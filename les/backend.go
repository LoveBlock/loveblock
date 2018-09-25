// Copyright 2016 The loveblock Authors
// This file is part of the loveblock library.
//
// The loveblock library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The loveblock library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the loveblock library. If not, see <http://www.gnu.org/licenses/>.

// Package les implements the Light LoveBlock Subprotocol.
package les

import (
	"fmt"
	"sync"
	"time"

	"github.com/LoveBlock/loveblock/accounts"
	"github.com/LoveBlock/loveblock/common"
	"github.com/LoveBlock/loveblock/common/hexutil"
	"github.com/LoveBlock/loveblock/consensus"
	"github.com/LoveBlock/loveblock/core"
	"github.com/LoveBlock/loveblock/core/bloombits"
	"github.com/LoveBlock/loveblock/core/types"
	"github.com/LoveBlock/loveblock/event"
	"github.com/LoveBlock/loveblock/internal/loveapi"
	"github.com/LoveBlock/loveblock/light"
	"github.com/LoveBlock/loveblock/log"
	"github.com/LoveBlock/loveblock/lovedb"
	"github.com/LoveBlock/loveblock/network"
	"github.com/LoveBlock/loveblock/network/downloader"
	"github.com/LoveBlock/loveblock/network/filters"
	"github.com/LoveBlock/loveblock/network/gasprice"
	"github.com/LoveBlock/loveblock/node"
	"github.com/LoveBlock/loveblock/p2p"
	"github.com/LoveBlock/loveblock/p2p/discv5"
	"github.com/LoveBlock/loveblock/params"
	rpc "github.com/LoveBlock/loveblock/rpc"
)

type LightLoveblock struct {
	config *network.Config

	odr         *LesOdr
	relay       *LesTxRelay
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan chan bool
	// Handlers
	peers           *peerSet
	txPool          *light.TxPool
	blockchain      *light.LightChain
	protocolManager *ProtocolManager
	serverPool      *serverPool
	reqDist         *requestDistributor
	retriever       *retrieveManager
	// DB interfaces
	chainDb lovedb.Database // Block chain database

	bloomRequests                              chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer, chtIndexer, bloomTrieIndexer *core.ChainIndexer

	ApiBackend *LesApiBackend

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	networkId     uint64
	netRPCService *loveapi.PublicNetAPI

	wg sync.WaitGroup
}

func New(ctx *node.ServiceContext, config *network.Config) (*LightLoveblock, error) {
	chainDb, err := network.CreateDB(ctx, config, "lightchaindata")
	if err != nil {
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, isCompat := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !isCompat {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	peers := newPeerSet()
	quitSync := make(chan struct{})

	lnetwork := &LightLoveblock{
		config:           config,
		chainConfig:      chainConfig,
		chainDb:          chainDb,
		eventMux:         ctx.EventMux,
		peers:            peers,
		reqDist:          newRequestDistributor(peers, quitSync),
		accountManager:   ctx.AccountManager,
		engine:           network.CreateConsensusEngine(ctx, chainConfig, chainDb, common.Address{}),
		shutdownChan:     make(chan bool),
		networkId:        config.NetworkId,
		bloomRequests:    make(chan chan *bloombits.Retrieval),
		bloomIndexer:     network.NewBloomIndexer(chainDb, light.BloomTrieFrequency),
		chtIndexer:       light.NewChtIndexer(chainDb, true),
		bloomTrieIndexer: light.NewBloomTrieIndexer(chainDb, true),
	}
	lnetwork.relay = NewLesTxRelay(peers, lnetwork.reqDist)
	lnetwork.serverPool = newServerPool(chainDb, quitSync, &lnetwork.wg)
	lnetwork.retriever = newRetrieveManager(peers, lnetwork.reqDist, lnetwork.serverPool)
	lnetwork.odr = NewLesOdr(chainDb, lnetwork.chtIndexer, lnetwork.bloomTrieIndexer, lnetwork.bloomIndexer, lnetwork.retriever)
	if lnetwork.blockchain, err = light.NewLightChain(lnetwork.odr, lnetwork.chainConfig, lnetwork.engine); err != nil {
		return nil, err
	}
	lnetwork.bloomIndexer.Start(lnetwork.blockchain)
	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		lnetwork.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}

	lnetwork.txPool = light.NewTxPool(lnetwork.chainConfig, lnetwork.blockchain, lnetwork.relay)
	if lnetwork.protocolManager, err = NewProtocolManager(lnetwork.chainConfig, true, ClientProtocolVersions, config.NetworkId, lnetwork.eventMux, lnetwork.engine, lnetwork.peers, lnetwork.blockchain, nil, chainDb, lnetwork.odr, lnetwork.relay, quitSync, &lnetwork.wg); err != nil {
		return nil, err
	}
	lnetwork.ApiBackend = &LesApiBackend{lnetwork, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	lnetwork.ApiBackend.gpo = gasprice.NewOracle(lnetwork.ApiBackend, gpoParams)
	return lnetwork, nil
}

func lesTopic(genesisHash common.Hash, protocolVersion uint) discv5.Topic {
	var name string
	switch protocolVersion {
	case lpv1:
		name = "LES"
	case lpv2:
		name = "LES2"
	default:
		panic(nil)
	}
	return discv5.Topic(name + "@" + common.Bytes2Hex(genesisHash.Bytes()[0:8]))
}

type LightDummyAPI struct{}

// Lovebase is the address that mining rewards will be send to
func (s *LightDummyAPI) Lovebase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Coinbase is the address that mining rewards will be send to (alias for Lovebase)
func (s *LightDummyAPI) Coinbase() (common.Address, error) {
	return common.Address{}, fmt.Errorf("not supported")
}

// Hashrate returns the POW hashrate
func (s *LightDummyAPI) Hashrate() hexutil.Uint {
	return 0
}

// Mining returns an indication if this node is currently mining.
func (s *LightDummyAPI) Mining() bool {
	return false
}

// APIs returns the collection of RPC services the LoveBlock package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *LightLoveblock) APIs() []rpc.API {
	return append(loveapi.GetAPIs(s.ApiBackend), []rpc.API{
		{
			Namespace: "network",
			Version:   "1.0",
			Service:   &LightDummyAPI{},
			Public:    true,
		}, {
			Namespace: "network",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "network",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, true),
			Public:    true,
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *LightLoveblock) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *LightLoveblock) BlockChain() *light.LightChain      { return s.blockchain }
func (s *LightLoveblock) TxPool() *light.TxPool              { return s.txPool }
func (s *LightLoveblock) Engine() consensus.Engine           { return s.engine }
func (s *LightLoveblock) LesVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *LightLoveblock) Downloader() *downloader.Downloader { return s.protocolManager.downloader }
func (s *LightLoveblock) EventMux() *event.TypeMux           { return s.eventMux }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *LightLoveblock) Protocols() []p2p.Protocol {
	return s.protocolManager.SubProtocols
}

// Start implements node.Service, starting all internal goroutines needed by the
// LoveBlock protocol implementation.
func (s *LightLoveblock) Start(srvr *p2p.Server) error {
	s.startBloomHandlers()
	log.Warn("Light client mode is an experimental feature")
	s.netRPCService = loveapi.NewPublicNetAPI(srvr, s.networkId)
	// clients are searching for the first advertised protocol in the list
	protocolVersion := AdvertiseProtocolVersions[0]
	s.serverPool.start(srvr, lesTopic(s.blockchain.Genesis().Hash(), protocolVersion))
	s.protocolManager.Start(s.config.LightPeers)
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// LoveBlock protocol.
func (s *LightLoveblock) Stop() error {
	s.odr.Stop()
	if s.bloomIndexer != nil {
		s.bloomIndexer.Close()
	}
	if s.chtIndexer != nil {
		s.chtIndexer.Close()
	}
	if s.bloomTrieIndexer != nil {
		s.bloomTrieIndexer.Close()
	}
	s.blockchain.Stop()
	s.protocolManager.Stop()
	s.txPool.Stop()

	s.eventMux.Stop()

	time.Sleep(time.Millisecond * 200)
	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}

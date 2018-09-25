// Copyright 2014 The loveblock Authors
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

// Package network implements the LoveBlock protocol.
package network

import (
	"errors"
	"fmt"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/LoveBlock/loveblock/accounts"
	"github.com/LoveBlock/loveblock/common"
	"github.com/LoveBlock/loveblock/common/hexutil"
	"github.com/LoveBlock/loveblock/consensus"
	"github.com/LoveBlock/loveblock/consensus/dpovp"
	"github.com/LoveBlock/loveblock/core"
	"github.com/LoveBlock/loveblock/core/bloombits"
	"github.com/LoveBlock/loveblock/core/types"
	"github.com/LoveBlock/loveblock/core/vm"
	"github.com/LoveBlock/loveblock/event"
	"github.com/LoveBlock/loveblock/internal/loveapi"
	"github.com/LoveBlock/loveblock/log"
	"github.com/LoveBlock/loveblock/lovedb"
	"github.com/LoveBlock/loveblock/miner"
	"github.com/LoveBlock/loveblock/network/downloader"
	"github.com/LoveBlock/loveblock/network/filters"
	"github.com/LoveBlock/loveblock/network/gasprice"
	"github.com/LoveBlock/loveblock/node"
	"github.com/LoveBlock/loveblock/p2p"
	"github.com/LoveBlock/loveblock/params"
	"github.com/LoveBlock/loveblock/rlp"
	"github.com/LoveBlock/loveblock/rpc"
)

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
	SetBloomBitsIndexer(bbIndexer *core.ChainIndexer)
}

// LoveBlock implements the LoveBlock full node service.
type Loveblock struct {
	config      *Config
	chainConfig *params.ChainConfig

	// Channel for shutting down the service
	shutdownChan  chan bool    // Channel for shutting down the LoveBlock
	stopDbUpgrade func() error // stop chain db sequential key upgrade

	// Handlers
	txPool          *core.TxPool
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer

	// DB interfaces
	chainDb lovedb.Database // Block chain database

	eventMux       *event.TypeMux
	engine         consensus.Engine
	accountManager *accounts.Manager

	bloomRequests chan chan *bloombits.Retrieval // Channel receiving bloom data retrieval requests
	bloomIndexer  *core.ChainIndexer             // Bloom indexer operating during block imports

	ApiBackend *LoveApiBackend

	miner       *miner.Miner
	gasPrice    *big.Int
	networkbase common.Address

	networkId     uint64
	netRPCService *loveapi.PublicNetAPI

	lock sync.RWMutex // Protects the variadic fields (e.g. gas price and networkbase)
}

func (s *Loveblock) AddLesServer(ls LesServer) {
	s.lesServer = ls
	ls.SetBloomBitsIndexer(s.bloomIndexer)
}

// New creates a new LoveBlock object (including the
// initialisation of the common LoveBlock object)
func New(ctx *node.ServiceContext, config *Config) (*Loveblock, error) {
	if config.SyncMode == downloader.LightSync {
		return nil, errors.New("can't run network.Loveblock in light sync mode, use les.LightLoveblock")
	}
	if !config.SyncMode.IsValid() {
		return nil, fmt.Errorf("invalid sync mode %d", config.SyncMode)
	}
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeDeduplicateData(chainDb)
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Info("Initialised chain configuration", "config", chainConfig)

	network := &Loveblock{
		config:         config,
		chainDb:        chainDb,
		chainConfig:    chainConfig,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		networkId:      config.NetworkId,
		gasPrice:       config.GasPrice,
		networkbase:    config.Lovebase,
		bloomRequests:  make(chan chan *bloombits.Retrieval),
		bloomIndexer:   NewBloomIndexer(chainDb, params.BloomBitsBlocks),
	}
	// sman modify
	network.engine = CreateConsensusEngine(ctx, chainConfig, chainDb, config.Lovebase)

	log.Info("Initialising Loveblock protocol", "versions", ProtocolVersions, "network", config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := core.GetBlockChainVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run loveblock upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		core.WriteBlockChainVersion(chainDb, core.BlockChainVersion)
	}
	var (
		vmConfig    = vm.Config{EnablePreimageRecording: config.EnablePreimageRecording}
		cacheConfig = &core.CacheConfig{Disabled: config.NoPruning, TrieNodeLimit: config.TrieCache, TrieTimeLimit: config.TrieTimeout}
	)
	network.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, network.chainConfig, network.engine, vmConfig)
	if err != nil {
		return nil, err
	}

	// Rewind the chain in case of an incompatible config upgrade.
	if compat, ok := genesisErr.(*params.ConfigCompatError); ok {
		log.Warn("Rewinding chain to upgrade configuration", "err", compat)
		network.blockchain.SetHead(compat.RewindTo)
		core.WriteChainConfig(chainDb, genesisHash, chainConfig)
	}
	network.bloomIndexer.Start(network.blockchain)

	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}
	network.txPool = core.NewTxPool(config.TxPool, network.chainConfig, network.blockchain)

	if network.protocolManager, err = NewProtocolManager(network.chainConfig, config.SyncMode, config.NetworkId, network.eventMux, network.txPool, network.engine, network.blockchain, chainDb); err != nil {
		return nil, err
	}
	network.miner = miner.New(network, network.chainConfig, network.EventMux(), network.engine)
	network.miner.SetExtra(makeExtraData(config.ExtraData))
	// sman
	if config.NodeMode == NodeModeStar {
		network.miner.SetStarNodeFlag()
	}
	// sman
	network.blockchain.SetIsStarNode(config.NodeMode == NodeModeStar)

	network.ApiBackend = &LoveApiBackend{network, nil}
	gpoParams := config.GPO
	if gpoParams.Default == nil {
		gpoParams.Default = config.GasPrice
	}
	network.ApiBackend.gpo = gasprice.NewOracle(network.ApiBackend, gpoParams)

	return network, nil
}

func makeExtraData(extra []byte) []byte {
	if len(extra) == 0 {
		// create default extradata
		extra, _ = rlp.EncodeToBytes([]interface{}{
			uint(params.VersionMajor<<16 | params.VersionMinor<<8 | params.VersionPatch),
			"loveblock",
			runtime.Version(),
			runtime.GOOS,
		})
	}
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		log.Warn("Miner extra data exceed limit", "extra", hexutil.Bytes(extra), "limit", params.MaximumExtraDataSize)
		extra = nil
	}
	return extra
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (lovedb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	if db, ok := db.(*lovedb.LDBDatabase); ok {
		db.Meter("network/db/chaindata/")
	}
	return db, nil
}

// CreateConsensusEngine creates the required type of consensus engine instance for an LoveBlock service
func CreateConsensusEngine(ctx *node.ServiceContext, chainConfig *params.ChainConfig, db lovedb.Database, coinbase common.Address) consensus.Engine {
	// sman 此处路由到我们的新的共识方法：DPOVP
	if chainConfig.Dpovp != nil {
		return dpovp.New(chainConfig.Dpovp, db, coinbase)
	} else {
		log.Error(`sman not dpovp`)
		return nil
	}
}

// APIs returns the collection of RPC services the LoveBlock package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Loveblock) APIs() []rpc.API {
	apis := loveapi.GetAPIs(s.ApiBackend)

	// Append any APIs exposed explicitly by the consensus engine
	apis = append(apis, s.engine.APIs(s.BlockChain())...)
	apis = append(apis, []rpc.API{
		{
			Namespace: "network",
			Version:   "1.0",
			Service:   NewPublicLoveblockAPI(s),
			Public:    true,
		}, {
			Namespace: "network",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "network",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "network",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
	// Append all the local APIs and return
	return apis
}

func (s *Loveblock) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Loveblock) Lovebase() (eb common.Address, err error) {
	s.lock.RLock()
	networkbase := s.networkbase
	s.lock.RUnlock()

	if networkbase != (common.Address{}) {
		return networkbase, nil
	}
	if wallets := s.AccountManager().Wallets(); len(wallets) > 0 {
		if accounts := wallets[0].Accounts(); len(accounts) > 0 {
			networkbase := accounts[0].Address

			s.lock.Lock()
			s.networkbase = networkbase
			s.lock.Unlock()

			log.Info("Lovebase automatically configured", "address", networkbase)
			return networkbase, nil
		}
	}
	return common.Address{}, fmt.Errorf("networkbase must be explicitly specified")
}

// set in js console via admin interface or wrapper from cli flags
func (self *Loveblock) SetLovebase(networkbase common.Address) {
	self.lock.Lock()
	self.networkbase = networkbase
	self.lock.Unlock()

	self.miner.SetLovebase(networkbase)
}

func (s *Loveblock) StartMining(local bool) error {
	eb, err := s.Lovebase()
	if err != nil {
		log.Error("Cannot start mining without networkbase", "err", err)
		return fmt.Errorf("networkbase missing: %v", err)
	}
	if local {
		// If local (CPU) mining is started, we can disable the transaction rejection
		// mechanism introduced to speed sync times. CPU mining on mainnet is ludicrous
		// so noone will ever hit this path, whereas marking sync done on CPU mining
		// will ensure that private networks work in single miner mode too.
		atomic.StoreUint32(&s.protocolManager.acceptTxs, 1)
	}
	go s.miner.Start(eb)
	return nil
}

func (s *Loveblock) StopMining()         { s.miner.Stop() }
func (s *Loveblock) IsMining() bool      { return s.miner.Mining() }
func (s *Loveblock) Miner() *miner.Miner { return s.miner }

func (s *Loveblock) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Loveblock) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Loveblock) TxPool() *core.TxPool               { return s.txPool }
func (s *Loveblock) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Loveblock) Engine() consensus.Engine           { return s.engine }
func (s *Loveblock) ChainDb() lovedb.Database           { return s.chainDb }
func (s *Loveblock) IsListening() bool                  { return true } // Always listening
func (s *Loveblock) LoveVersion() int                   { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Loveblock) NetVersion() uint64                 { return s.networkId }
func (s *Loveblock) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Loveblock) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	}
	return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
}

// Start implements node.Service, starting all internal goroutines needed by the
// LoveBlock protocol implementation.
func (s *Loveblock) Start(srvr *p2p.Server) error {
	// sman set coinbase to blockchain
	coinbase, _ := s.Lovebase()
	s.blockchain.SetCoinbase(coinbase)

	// Start the bloom bits servicing goroutines
	s.startBloomHandlers()

	// Start the RPC service
	s.netRPCService = loveapi.NewPublicNetAPI(srvr, s.NetVersion())

	// Figure out a max peers count based on the server limits
	maxPeers := srvr.MaxPeers
	if s.config.LightServ > 0 {
		if s.config.LightPeers >= srvr.MaxPeers {
			return fmt.Errorf("invalid peer config: light peer count (%d) >= total peer count (%d)", s.config.LightPeers, srvr.MaxPeers)
		}
		maxPeers -= s.config.LightPeers
	}
	// Start the networking layer and the light server if requested
	s.protocolManager.Start(maxPeers)
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// LoveBlock protocol.
func (s *Loveblock) Stop() error {
	if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.bloomIndexer.Close()
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}

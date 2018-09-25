// Copyright 2015 The loveblock Authors
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

package network

import (
	"context"
	"math/big"

	"fmt"
	"github.com/LoveBlock/loveblock/accounts"
	"github.com/LoveBlock/loveblock/common"
	"github.com/LoveBlock/loveblock/common/math"
	"github.com/LoveBlock/loveblock/core"
	"github.com/LoveBlock/loveblock/core/bloombits"
	"github.com/LoveBlock/loveblock/core/state"
	"github.com/LoveBlock/loveblock/core/types"
	"github.com/LoveBlock/loveblock/core/vm"
	"github.com/LoveBlock/loveblock/event"
	"github.com/LoveBlock/loveblock/lovedb"
	"github.com/LoveBlock/loveblock/network/downloader"
	"github.com/LoveBlock/loveblock/network/gasprice"
	"github.com/LoveBlock/loveblock/params"
	"github.com/LoveBlock/loveblock/rpc"
)

// LoveApiBackend implements loveapi.Backend for full nodes
type LoveApiBackend struct {
	network *Loveblock
	gpo     *gasprice.Oracle
}

func (b *LoveApiBackend) ChainConfig() *params.ChainConfig {
	return b.network.chainConfig
}

func (b *LoveApiBackend) CurrentBlock() *types.Block {
	return b.network.blockchain.CurrentBlock()
}

func (b *LoveApiBackend) SetHead(number uint64) {
	b.network.protocolManager.downloader.Cancel()
	b.network.blockchain.SetHead(number)
}

func (b *LoveApiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.network.miner.PendingBlock()
		if block == nil {
			return nil, fmt.Errorf("PendingBlock is nil")
		}
		return block.Header(), nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.network.blockchain.CurrentBlock().Header(), nil
	}
	return b.network.blockchain.GetHeaderByNumber(uint64(blockNr)), nil
}

func (b *LoveApiBackend) BlockByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Block, error) {
	// Pending block is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block := b.network.miner.PendingBlock()
		if block == nil {
			return nil, fmt.Errorf("PendingBlock is nil")
		}
		return block, nil
	}
	// Otherwise resolve and return the block
	if blockNr == rpc.LatestBlockNumber {
		return b.network.blockchain.CurrentBlock(), nil
	}
	return b.network.blockchain.GetBlockByNumber(uint64(blockNr)), nil
}

func (b *LoveApiBackend) StateAndHeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*state.StateDB, *types.Header, error) {
	// Pending state is only known by the miner
	if blockNr == rpc.PendingBlockNumber {
		block, state := b.network.miner.Pending()
		return state, block.Header(), nil
	}
	// Otherwise resolve the block number and return its state
	header, err := b.HeaderByNumber(ctx, blockNr)
	if header == nil || err != nil {
		return nil, nil, err
	}
	stateDb, err := b.network.BlockChain().StateAt(header.Root)
	return stateDb, header, err
}

func (b *LoveApiBackend) GetBlock(ctx context.Context, blockHash common.Hash) (*types.Block, error) {
	return b.network.blockchain.GetBlockByHash(blockHash), nil
}

func (b *LoveApiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (types.Receipts, error) {
	return core.GetBlockReceipts(b.network.chainDb, blockHash, core.GetBlockNumber(b.network.chainDb, blockHash)), nil
}

func (b *LoveApiBackend) GetLogs(ctx context.Context, blockHash common.Hash) ([][]*types.Log, error) {
	receipts := core.GetBlockReceipts(b.network.chainDb, blockHash, core.GetBlockNumber(b.network.chainDb, blockHash))
	if receipts == nil {
		return nil, nil
	}
	logs := make([][]*types.Log, len(receipts))
	for i, receipt := range receipts {
		logs[i] = receipt.Logs
	}
	return logs, nil
}

func (b *LoveApiBackend) GetTd(blockHash common.Hash) *big.Int {
	return b.network.blockchain.GetTdByHash(blockHash)
}

func (b *LoveApiBackend) GetEVM(ctx context.Context, msg core.Message, state *state.StateDB, header *types.Header, vmCfg vm.Config) (*vm.EVM, func() error, error) {
	state.SetBalance(msg.From(), math.MaxBig256)
	vmError := func() error { return nil }

	context := core.NewEVMContext(msg, header, b.network.BlockChain(), nil)
	return vm.NewEVM(context, state, b.network.chainConfig, vmCfg), vmError, nil
}

func (b *LoveApiBackend) SubscribeRemovedLogsEvent(ch chan<- core.RemovedLogsEvent) event.Subscription {
	return b.network.BlockChain().SubscribeRemovedLogsEvent(ch)
}

func (b *LoveApiBackend) SubscribeChainEvent(ch chan<- core.ChainEvent) event.Subscription {
	return b.network.BlockChain().SubscribeChainEvent(ch)
}

func (b *LoveApiBackend) SubscribeChainHeadEvent(ch chan<- core.ChainHeadEvent) event.Subscription {
	return b.network.BlockChain().SubscribeChainHeadEvent(ch)
}

func (b *LoveApiBackend) SubscribeChainSideEvent(ch chan<- core.ChainSideEvent) event.Subscription {
	return b.network.BlockChain().SubscribeChainSideEvent(ch)
}

func (b *LoveApiBackend) SubscribeLogsEvent(ch chan<- []*types.Log) event.Subscription {
	return b.network.BlockChain().SubscribeLogsEvent(ch)
}

func (b *LoveApiBackend) SendTx(ctx context.Context, signedTx *types.Transaction) error {
	return b.network.txPool.AddLocal(signedTx)
}

func (b *LoveApiBackend) GetPoolTransactions() (types.Transactions, error) {
	pending, err := b.network.txPool.Pending()
	if err != nil {
		return nil, err
	}
	var txs types.Transactions
	for _, batch := range pending {
		txs = append(txs, batch...)
	}
	return txs, nil
}

func (b *LoveApiBackend) GetPoolTransaction(hash common.Hash) *types.Transaction {
	return b.network.txPool.Get(hash)
}

func (b *LoveApiBackend) GetPoolNonce(ctx context.Context, addr common.Address) (uint64, error) {
	return b.network.txPool.State().GetNonce(addr), nil
}

func (b *LoveApiBackend) Stats() (pending int, queued int) {
	return b.network.txPool.Stats()
}

func (b *LoveApiBackend) TxPoolContent() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	return b.network.TxPool().Content()
}

func (b *LoveApiBackend) SubscribeTxPreEvent(ch chan<- core.TxPreEvent) event.Subscription {
	return b.network.TxPool().SubscribeTxPreEvent(ch)
}

func (b *LoveApiBackend) Downloader() *downloader.Downloader {
	return b.network.Downloader()
}

func (b *LoveApiBackend) ProtocolVersion() int {
	return b.network.LoveVersion()
}

func (b *LoveApiBackend) SuggestPrice(ctx context.Context) (*big.Int, error) {
	return b.gpo.SuggestPrice(ctx)
}

func (b *LoveApiBackend) ChainDb() lovedb.Database {
	return b.network.ChainDb()
}

func (b *LoveApiBackend) EventMux() *event.TypeMux {
	return b.network.EventMux()
}

func (b *LoveApiBackend) AccountManager() *accounts.Manager {
	return b.network.AccountManager()
}

func (b *LoveApiBackend) BloomStatus() (uint64, uint64) {
	sections, _, _ := b.network.bloomIndexer.Sections()
	return params.BloomBitsBlocks, sections
}

func (b *LoveApiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < bloomFilterThreads; i++ {
		go session.Multiplex(bloomRetrievalBatch, bloomRetrievalWait, b.network.bloomRequests)
	}
}

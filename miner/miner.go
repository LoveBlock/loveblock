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

// Package miner implements LoveBlock block creation and mining.
package miner

import (
	"fmt"
	"sync/atomic"

	"github.com/LoveBlock/loveblock/accounts"
	"github.com/LoveBlock/loveblock/common"
	dpovpCommon "github.com/LoveBlock/loveblock/common/dpovp"
	"github.com/LoveBlock/loveblock/consensus"
	"github.com/LoveBlock/loveblock/consensus/dpovp"
	"github.com/LoveBlock/loveblock/core"
	"github.com/LoveBlock/loveblock/core/state"
	"github.com/LoveBlock/loveblock/core/types"
	"github.com/LoveBlock/loveblock/event"
	"github.com/LoveBlock/loveblock/log"
	"github.com/LoveBlock/loveblock/lovedb"
	"github.com/LoveBlock/loveblock/network/downloader"
	"github.com/LoveBlock/loveblock/params"
)

// sman 封装挖矿用的所有方法
// Backend wraps all methods required for mining.
type Backend interface {
	AccountManager() *accounts.Manager
	BlockChain() *core.BlockChain
	TxPool() *core.TxPool
	ChainDb() lovedb.Database
}

// Miner creates blocks and searches for proof-of-work values.
type Miner struct {
	mux    *event.TypeMux
	worker *worker // sman delete

	coinbase common.Address   // sman 挖矿收益地址
	mining   int32            // sman 是否开启挖矿
	network  Backend          // sman 用于挖矿的所有封装
	engine   consensus.Engine // sman 确认共识机

	// sman 是否在同步区块
	canStart int32 // can start indicates whether we can start the mining operation
	// sman 标识是否同步区块后立即挖矿
	shouldStart  int32 // should start indicates whether we should start after sync
	starNodeFlag int32
}

func (self *Miner) CoinBase() common.Address {
	return self.coinbase
}

func New(network Backend, config *params.ChainConfig, mux *event.TypeMux, engine consensus.Engine) *Miner {
	miner := &Miner{
		network:      network,
		mux:          mux,
		engine:       engine,
		worker:       newWorker(config, engine, common.Address{}, network, mux),
		canStart:     1,
		starNodeFlag: 0,
	}
	miner.Register(NewCpuAgent(network.BlockChain(), engine))
	go miner.update()

	return miner
}

// update keeps track of the downloader events. Please be aware that this is a one shot type of update loop.
// It's entered once and as soon as `Done` or `Failed` has been broadcasted the events are unregistered and
// the loop is exited. This to prevent a major security vuln where external parties can DOS you with blocks
// and halt your mining operation for as long as the DOS continues.
func (self *Miner) update() {
	events := self.mux.Subscribe(downloader.StartEvent{}, downloader.DoneEvent{}, downloader.FailedEvent{})
out:
	for ev := range events.Chan() {
		switch ev.Data.(type) {
		case downloader.StartEvent:
			atomic.StoreInt32(&self.canStart, 0)
			if self.Mining() {
				self.Stop()
				atomic.StoreInt32(&self.shouldStart, 1)
				log.Info("Mining aborted due to sync")
			}
		case downloader.DoneEvent, downloader.FailedEvent:
			shouldStart := atomic.LoadInt32(&self.shouldStart) == 1

			atomic.StoreInt32(&self.canStart, 1)
			atomic.StoreInt32(&self.shouldStart, 0)
			if shouldStart {
				self.Start(self.coinbase)
			}
			// unsubscribe. we're only interested in this event once
			events.Unsubscribe()
			// stop immediately and ignore all further pending events
			break out
		}
	}
}

func (self *Miner) Start(coinbase common.Address) {
	if atomic.LoadInt32(&self.starNodeFlag) == 0 {
		log.Warn("Not a Star Node!")
		return
	}
	if dpovpCommon.GetCoreNodesCount() == 0 {
		log.Warn("At least one star node!")
		return
	}
	atomic.StoreInt32(&self.shouldStart, 1)
	self.SetLovebase(coinbase)

	if atomic.LoadInt32(&self.canStart) == 0 {
		log.Info("Network syncing, will start miner afterwards")
		return
	}
	atomic.StoreInt32(&self.mining, 1)

	log.Info("Starting mining operation")
	self.worker.start()
}

func (self *Miner) Stop() {
	self.worker.stop()
	atomic.StoreInt32(&self.mining, 0)
	atomic.StoreInt32(&self.shouldStart, 0)
}

func (self *Miner) Register(agent Agent) {
	if self.Mining() {
		agent.Start()
	}
	self.worker.register(agent)
}

func (self *Miner) Unregister(agent Agent) {
	self.worker.unregister(agent)
}

func (self *Miner) Mining() bool {
	return atomic.LoadInt32(&self.mining) > 0
}

func (self *Miner) SetExtra(extra []byte) error {
	if uint64(len(extra)) > params.MaximumExtraDataSize {
		return fmt.Errorf("Extra exceeds max length. %d > %v", len(extra), params.MaximumExtraDataSize)
	}
	self.worker.setExtra(extra)
	return nil
}

// Pending returns the currently pending block and associated state.
func (self *Miner) Pending() (*types.Block, *state.StateDB) {
	return self.worker.pending()
}

// PendingBlock returns the currently pending block.
//
// Note, to access both the pending block and the pending state
// simultaneously, please use Pending(), as the pending state can
// change between multiple method calls
func (self *Miner) PendingBlock() *types.Block {
	return self.worker.pendingBlock()
}

func (self *Miner) SetLovebase(addr common.Address) {
	self.coinbase = addr
	self.worker.setLovebase(addr)
	self.engine.(*dpovp.Dpovp).SetCoinbase(addr)
}

func (self *Miner) SetStarNodeFlag() {
	atomic.StoreInt32(&self.starNodeFlag, 1)
}

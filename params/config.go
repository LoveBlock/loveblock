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

package params

import (
	"fmt"
	"math/big"

	"github.com/LoveBlock/loveblock/common"
)

var (
	// TODO
	MainnetGenesisHash = common.HexToHash("0xff4a7c7beda3c3eb3fb6eb2fd90f75db88f1ad0894ec773db43028787121691e") // Mainnet genesis hash to enforce below configs on
	TestnetGenesisHash = common.HexToHash("0x682c33863395ea58a7e08c63b0a655ffcdc707d1c03c8dda791f3642ba441a96") // Testnet genesis hash to enforce below configs on
)

var (
	// MainnetChainConfig is the chain parameters to run a node on the main network.
	MainnetChainConfig = &ChainConfig{
		ChainId: big.NewInt(20180627),
		Dpovp:   &DpovpConfig{Timeout: 10 * 1000, Sleeptime: 3 * 1000},
	}

	// TestnetChainConfig contains the chain parameters to run a node on the Test network.
	TestnetChainConfig = &ChainConfig{
		ChainId: big.NewInt(100),
		Dpovp:   &DpovpConfig{Timeout: 10 * 1000, Sleeptime: 3 * 1000},
	}

	// AllLovehashProtocolChanges contains every protocol change (EIPs) introduced
	// and accepted by the LoveBlock core developers into the Lovehash consensus.
	//
	// This configuration is intentionally not using keyed fields to force anyone
	// adding flags to the config to also have to set these fields.
	// AllLovehashProtocolChanges = &ChainConfig{big.NewInt(1337), &DpovpConfig{Timeout: 10 * 1000, Sleeptime: 3 * 1000}}
	AllLovehashProtocolChanges = &ChainConfig{big.NewInt(20180627), &DpovpConfig{Timeout: 10 * 1000, Sleeptime: 3 * 1000}}

	TestChainConfig = &ChainConfig{big.NewInt(1), new(DpovpConfig)}
)

// ChainConfig is the core config which determines the blockchain settings.
//
// ChainConfig is stored in the database on a per block basis. This means
// that any network, identified by its genesis block, can have its own
// set of configuration options.
type ChainConfig struct {
	ChainId *big.Int `json:"chainId"` // Chain id identifies the current chain and is used for replay protection

	// Various consensus engines
	Dpovp *DpovpConfig `json:"dpovp,omitempty"` // sman for dpovp
}

// sman DpovpConfig is the consensus engine configs for dpos
type DpovpConfig struct {
	Timeout   int64 `json:"Timeout"`   // Number of timeout between blocks to produce millsecond
	Sleeptime int64 `json:"Sleeptime"` // Time of one block is produced and before ohter node begin produce another block millsecond
}

// String implements the fmt.Stringer interface.
func (c *ChainConfig) String() string {
	var engine interface{}
	switch {
	case c.Dpovp != nil:
		engine = c.Dpovp
	default:
		engine = "unknown"
	}
	return fmt.Sprintf("{ChainID: %v Engine: %v}",
		c.ChainId,
		engine,
	)
}

// ConfigCompatError is raised if the locally-stored blockchain is initialised with a
// ChainConfig that would alter the past.
type ConfigCompatError struct {
	What string
	// block numbers of the stored and new configurations
	StoredConfig, NewConfig *big.Int
	// the block number to which the local chain must be rewound to correct the error
	RewindTo uint64
}

func (err *ConfigCompatError) Error() string {
	return fmt.Sprintf("mismatching %s in database (have %d, want %d, rewindto %d)", err.What, err.StoredConfig, err.NewConfig, err.RewindTo)
}

// Rules wraps ChainConfig and is merely syntatic sugar or can be used for functions
// that do not have or require information about the block.
//
// Rules is a one time interface meaning that it shouldn't be used in between transition
// phases.
type Rules struct {
	ChainId *big.Int
}

func (c *ChainConfig) Rules(num *big.Int) Rules {
	chainId := c.ChainId
	if chainId == nil {
		chainId = new(big.Int)
	}
	return Rules{ChainId: new(big.Int).Set(chainId)}
}

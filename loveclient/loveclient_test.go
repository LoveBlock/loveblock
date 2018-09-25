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

package loveclient

import "github.com/LoveBlock/loveblock"

// Verify that Client implements the LoveBlock interfaces.
var (
	_ = loveblock.ChainReader(&Client{})
	_ = loveblock.TransactionReader(&Client{})
	_ = loveblock.ChainStateReader(&Client{})
	_ = loveblock.ChainSyncReader(&Client{})
	_ = loveblock.ContractCaller(&Client{})
	_ = loveblock.GasEstimator(&Client{})
	_ = loveblock.GasPricer(&Client{})
	_ = loveblock.LogFilterer(&Client{})
	_ = loveblock.PendingStateReader(&Client{})
	// _ = LoveBlock.PendingStateEventer(&Client{})
	_ = loveblock.PendingContractCaller(&Client{})
)

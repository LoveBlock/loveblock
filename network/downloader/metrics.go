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

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/LoveBlock/loveblock/metrics"
)

var (
	headerInMeter      = metrics.NewRegisteredMeter("network/downloader/headers/in", nil)
	headerReqTimer     = metrics.NewRegisteredTimer("network/downloader/headers/req", nil)
	headerDropMeter    = metrics.NewRegisteredMeter("network/downloader/headers/drop", nil)
	headerTimeoutMeter = metrics.NewRegisteredMeter("network/downloader/headers/timeout", nil)

	bodyInMeter      = metrics.NewRegisteredMeter("network/downloader/bodies/in", nil)
	bodyReqTimer     = metrics.NewRegisteredTimer("network/downloader/bodies/req", nil)
	bodyDropMeter    = metrics.NewRegisteredMeter("network/downloader/bodies/drop", nil)
	bodyTimeoutMeter = metrics.NewRegisteredMeter("network/downloader/bodies/timeout", nil)

	receiptInMeter      = metrics.NewRegisteredMeter("network/downloader/receipts/in", nil)
	receiptReqTimer     = metrics.NewRegisteredTimer("network/downloader/receipts/req", nil)
	receiptDropMeter    = metrics.NewRegisteredMeter("network/downloader/receipts/drop", nil)
	receiptTimeoutMeter = metrics.NewRegisteredMeter("network/downloader/receipts/timeout", nil)

	stateInMeter   = metrics.NewRegisteredMeter("network/downloader/states/in", nil)
	stateDropMeter = metrics.NewRegisteredMeter("network/downloader/states/drop", nil)
)

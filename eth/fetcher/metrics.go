// Copyright 2015 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with go-ethereum.  If not, see <http://www.gnu.org/licenses/>.

// Contains the metrics collected by the fetcher.

package fetcher

import (
	"github.com/ethereum/go-ethereum/metrics"
)

var (
	announceMeter  = metrics.NewMeter("eth/sync/RemoteAnnounces")
	announceTimer  = metrics.NewTimer("eth/sync/LocalAnnounces")
	broadcastMeter = metrics.NewMeter("eth/sync/RemoteBroadcasts")
	broadcastTimer = metrics.NewTimer("eth/sync/LocalBroadcasts")
	discardMeter   = metrics.NewMeter("eth/sync/DiscardedBlocks")
	futureMeter    = metrics.NewMeter("eth/sync/FutureBlocks")
)

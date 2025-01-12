// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package bootstrap

import (
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/network/p2p"
	"github.com/shubhamdubey02/cryftgoftgo/snow"
	"github.com/shubhamdubey02/cryftgoftgo/snow/engine/avalanche/bootstrap/queue"
	"github.com/shubhamdubey02/cryftgoftgo/snow/engine/avalanche/vertex"
	"github.com/shubhamdubey02/cryftgoftgo/snow/engine/common"
	"github.com/shubhamdubey02/cryftgoftgo/snow/engine/common/tracker"
)

type Config struct {
	common.AllGetsServer

	Ctx *snow.ConsensusContext

	StartupTracker tracker.Startup
	Sender         common.Sender

	// PeerTracker manages the set of nodes that we fetch the next block from.
	PeerTracker *p2p.PeerTracker

	// This node will only consider the first [AncestorsMaxContainersReceived]
	// containers in an ancestors message it receives.
	AncestorsMaxContainersReceived int

	// VtxBlocked tracks operations that are blocked on vertices
	VtxBlocked *queue.JobsWithMissing
	// TxBlocked tracks operations that are blocked on transactions
	TxBlocked *queue.Jobs

	Manager vertex.Manager
	VM      vertex.LinearizableVM

	// If StopVertexID is empty, the engine will generate the stop vertex based
	// on the current state.
	StopVertexID ids.ID
}

// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package bootstrap

import (
	"github.com/shubhamdubey02/cryftgo/database"
	"github.com/shubhamdubey02/cryftgoftgo/network/p2p"
	"github.com/shubhamdubey02/cryftgoftgo/snow"
	"github.com/shubhamdubey02/cryftgoftgo/snow/engine/common"
	"github.com/shubhamdubey02/cryftgoftgo/snow/engine/common/tracker"
	"github.com/shubhamdubey02/cryftgoftgo/snow/engine/snowman/block"
	"github.com/shubhamdubey02/cryftgoftgo/snow/validators"
)

type Config struct {
	common.AllGetsServer

	Ctx     *snow.ConsensusContext
	Beacons validators.Manager

	SampleK          int
	StartupTracker   tracker.Startup
	Sender           common.Sender
	BootstrapTracker common.BootstrapTracker
	Timer            common.Timer

	// PeerTracker manages the set of nodes that we fetch the next block from.
	PeerTracker *p2p.PeerTracker

	// This node will only consider the first [AncestorsMaxContainersReceived]
	// containers in an ancestors message it receives.
	AncestorsMaxContainersReceived int

	// Database used to track the fetched, but not yet executed, blocks during
	// bootstrapping.
	DB database.Database

	VM block.ChainVM

	Bootstrapped func()
}

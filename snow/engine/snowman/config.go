// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowman

import (
	"github.com/shubhamdubey02/cryftgo/snow"
	"github.com/shubhamdubey02/cryftgoftgo/snow/consensus/snowball"
	"github.com/shubhamdubey02/cryftgoftgo/snow/consensus/snowman"
	"github.com/shubhamdubey02/cryftgoftgo/snow/engine/common"
	"github.com/shubhamdubey02/cryftgoftgo/snow/engine/common/tracker"
	"github.com/shubhamdubey02/cryftgoftgo/snow/engine/snowman/block"
	"github.com/shubhamdubey02/cryftgoftgo/snow/validators"
)

// Config wraps all the parameters needed for a snowman engine
type Config struct {
	common.AllGetsServer

	Ctx                 *snow.ConsensusContext
	VM                  block.ChainVM
	Sender              common.Sender
	Validators          validators.Manager
	ConnectedValidators tracker.Peers
	Params              snowball.Parameters
	Consensus           snowman.Consensus
	PartialSync         bool
}

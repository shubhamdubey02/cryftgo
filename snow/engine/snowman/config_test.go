// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowman

import (
	"testing"

	"github.com/shubhamdubey02/cryftgo/snow/consensus/snowball"
	"github.com/shubhamdubey02/cryftgo/snow/consensus/snowman"
	"github.com/shubhamdubey02/cryftgo/snow/engine/common"
	"github.com/shubhamdubey02/cryftgo/snow/engine/common/tracker"
	"github.com/shubhamdubey02/cryftgo/snow/engine/snowman/block"
	"github.com/shubhamdubey02/cryftgo/snow/snowtest"
	"github.com/shubhamdubey02/cryftgo/snow/validators"
)

func DefaultConfig(t testing.TB) Config {
	ctx := snowtest.Context(t, snowtest.PChainID)

	return Config{
		Ctx:                 snowtest.ConsensusContext(ctx),
		VM:                  &block.TestVM{},
		Sender:              &common.SenderTest{},
		Validators:          validators.NewManager(),
		ConnectedValidators: tracker.NewPeers(),
		Params: snowball.Parameters{
			K:                     1,
			AlphaPreference:       1,
			AlphaConfidence:       1,
			Beta:                  1,
			ConcurrentRepolls:     1,
			OptimalProcessing:     100,
			MaxOutstandingItems:   1,
			MaxItemProcessingTime: 1,
		},
		Consensus: &snowman.Topological{},
	}
}

// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowman

import (
	"context"

	"go.uber.org/zap"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/utils/bag"
	"github.com/shubhamdubey02/cryftgo/utils/set"
)

// Voter records chits received from [vdr] once its dependencies are met.
type voter struct {
	t               *Transitive
	vdr             ids.NodeID
	requestID       uint32
	responseOptions []ids.ID
	deps            set.Set[ids.ID]
}

func (v *voter) Dependencies() set.Set[ids.ID] {
	return v.deps
}

// Mark that a dependency has been met.
func (v *voter) Fulfill(ctx context.Context, id ids.ID) {
	v.deps.Remove(id)
	v.Update(ctx)
}

// Abandon this attempt to record chits.
func (v *voter) Abandon(ctx context.Context, id ids.ID) {
	v.Fulfill(ctx, id)
}

func (v *voter) Update(ctx context.Context) {
	if v.deps.Len() != 0 || v.t.errs.Errored() {
		return
	}

	var (
		vote       ids.ID
		shouldVote bool
		voteIndex  int
	)
	for i, voteOption := range v.responseOptions {
		// To prevent any potential deadlocks with undisclosed dependencies,
		// votes must be bubbled to the nearest valid block
		vote, shouldVote = v.t.getProcessingAncestor(ctx, voteOption)
		if shouldVote {
			voteIndex = i
			break
		}
	}

	var results []bag.Bag[ids.ID]
	if shouldVote {
		v.t.selectedVoteIndex.Observe(float64(voteIndex))
		results = v.t.polls.Vote(v.requestID, v.vdr, vote)
	} else {
		results = v.t.polls.Drop(v.requestID, v.vdr)
	}

	if len(results) == 0 {
		return
	}

	for _, result := range results {
		result := result
		v.t.Ctx.Log.Debug("finishing poll",
			zap.Stringer("result", &result),
		)
		if err := v.t.Consensus.RecordPoll(ctx, result); err != nil {
			v.t.errs.Add(err)
		}
	}

	if v.t.errs.Errored() {
		return
	}

	if err := v.t.VM.SetPreference(ctx, v.t.Consensus.Preference()); err != nil {
		v.t.errs.Add(err)
		return
	}

	if v.t.Consensus.NumProcessing() == 0 {
		v.t.Ctx.Log.Debug("Snowman engine can quiesce")
		return
	}

	v.t.Ctx.Log.Debug("Snowman engine can't quiesce")
	v.t.repoll(ctx)
}

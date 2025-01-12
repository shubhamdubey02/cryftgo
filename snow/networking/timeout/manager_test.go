// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package timeout

import (
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/snow/networking/benchlist"
	"github.com/shubhamdubey02/cryftgoftgo/utils/timer"
)

func TestManagerFire(t *testing.T) {
	benchlist := benchlist.NewNoBenchlist()
	manager, err := NewManager(
		&timer.AdaptiveTimeoutConfig{
			InitialTimeout:     time.Millisecond,
			MinimumTimeout:     time.Millisecond,
			MaximumTimeout:     10 * time.Second,
			TimeoutCoefficient: 1.25,
			TimeoutHalflife:    5 * time.Minute,
		},
		benchlist,
		"",
		prometheus.NewRegistry(),
	)
	require.NoError(t, err)
	go manager.Dispatch()
	defer manager.Stop()

	wg := sync.WaitGroup{}
	wg.Add(1)

	manager.RegisterRequest(
		ids.EmptyNodeID,
		ids.ID{},
		true,
		ids.RequestID{},
		wg.Done,
	)

	wg.Wait()
}

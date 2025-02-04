// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package handler

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/network/p2p"
	"github.com/shubhamdubey02/cryftgo/snow"
	"github.com/shubhamdubey02/cryftgo/snow/consensus/snowball"
	"github.com/shubhamdubey02/cryftgo/snow/engine/common"
	"github.com/shubhamdubey02/cryftgo/snow/networking/tracker"
	"github.com/shubhamdubey02/cryftgo/snow/snowtest"
	"github.com/shubhamdubey02/cryftgo/snow/validators"
	"github.com/shubhamdubey02/cryftgo/subnets"
	"github.com/shubhamdubey02/cryftgo/utils/logging"
	"github.com/shubhamdubey02/cryftgo/utils/math/meter"
	"github.com/shubhamdubey02/cryftgo/utils/resource"
	"github.com/shubhamdubey02/cryftgo/utils/set"
	"github.com/shubhamdubey02/cryftgo/version"

	p2ppb "github.com/shubhamdubey02/cryftgo/proto/pb/p2p"
	commontracker "github.com/shubhamdubey02/cryftgo/snow/engine/common/tracker"
)

func TestHealthCheckSubnet(t *testing.T) {
	tests := map[string]struct {
		consensusParams snowball.Parameters
	}{
		"default consensus params": {
			consensusParams: snowball.DefaultParameters,
		},
		"custom consensus params": {
			func() snowball.Parameters {
				params := snowball.DefaultParameters
				params.K = params.AlphaConfidence
				return params
			}(),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			snowCtx := snowtest.Context(t, snowtest.CChainID)
			ctx := snowtest.ConsensusContext(snowCtx)

			vdrs := validators.NewManager()

			resourceTracker, err := tracker.NewResourceTracker(
				prometheus.NewRegistry(),
				resource.NoUsage,
				meter.ContinuousFactory{},
				time.Second,
			)
			require.NoError(err)

			peerTracker := commontracker.NewPeers()
			vdrs.RegisterSetCallbackListener(ctx.SubnetID, peerTracker)

			sb := subnets.New(
				ctx.NodeID,
				subnets.Config{
					ConsensusParameters: test.consensusParams,
				},
			)

			p2pTracker, err := p2p.NewPeerTracker(
				logging.NoLog{},
				"",
				prometheus.NewRegistry(),
				nil,
				version.CurrentApp,
			)
			require.NoError(err)

			handlerIntf, err := New(
				ctx,
				vdrs,
				nil,
				time.Second,
				testThreadPoolSize,
				resourceTracker,
				validators.UnhandledSubnetConnector,
				sb,
				peerTracker,
				p2pTracker,
			)
			require.NoError(err)

			bootstrapper := &common.BootstrapperTest{
				EngineTest: common.EngineTest{
					T: t,
				},
			}
			bootstrapper.Default(false)

			engine := &common.EngineTest{T: t}
			engine.Default(false)
			engine.ContextF = func() *snow.ConsensusContext {
				return ctx
			}

			handlerIntf.SetEngineManager(&EngineManager{
				Snowman: &Engine{
					Bootstrapper: bootstrapper,
					Consensus:    engine,
				},
			})

			ctx.State.Set(snow.EngineState{
				Type:  p2ppb.EngineType_ENGINE_TYPE_SNOWMAN,
				State: snow.NormalOp, // assumed bootstrap is done
			})

			bootstrapper.StartF = func(context.Context, uint32) error {
				return nil
			}

			handlerIntf.Start(context.Background(), false)

			testVdrCount := 4
			vdrIDs := set.NewSet[ids.NodeID](testVdrCount)
			for i := 0; i < testVdrCount; i++ {
				vdrID := ids.GenerateTestNodeID()
				vdrIDs.Add(vdrID)

				require.NoError(vdrs.AddStaker(ctx.SubnetID, vdrID, nil, ids.Empty, 100))
			}

			for index, nodeID := range vdrIDs.List() {
				require.NoError(peerTracker.Connected(context.Background(), nodeID, nil))

				details, err := handlerIntf.HealthCheck(context.Background())
				expectedPercentConnected := float64(index+1) / float64(testVdrCount)
				conf := sb.Config()
				minPercentConnected := conf.ConsensusParameters.MinPercentConnectedHealthy()
				if expectedPercentConnected >= minPercentConnected {
					require.NoError(err)
					continue
				}
				require.ErrorIs(err, ErrNotConnectedEnoughStake)

				detailsMap, ok := details.(map[string]interface{})
				require.True(ok)
				networkingMap, ok := detailsMap["networking"]
				require.True(ok)
				networkingDetails, ok := networkingMap.(map[string]float64)
				require.True(ok)
				percentConnected, ok := networkingDetails["percentConnected"]
				require.True(ok)
				require.Equal(expectedPercentConnected, percentConnected)
			}
		})
	}
}

// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/utils/set"
	"github.com/shubhamdubey02/cryftgo/utils/units"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/state"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/txs/txstest"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/utxo"
	"github.com/shubhamdubey02/cryftgo/vms/secp256k1fx"
)

func TestCreateSubnetTxAP3FeeChange(t *testing.T) {
	ap3Time := defaultGenesisTime.Add(time.Hour)
	tests := []struct {
		name        string
		time        time.Time
		fee         uint64
		expectedErr error
	}{
		{
			name:        "pre-fork - correctly priced",
			time:        defaultGenesisTime,
			fee:         0,
			expectedErr: nil,
		},
		{
			name:        "post-fork - incorrectly priced",
			time:        ap3Time,
			fee:         100*defaultTxFee - 1*units.NanoCryft,
			expectedErr: utxo.ErrInsufficientUnlockedFunds,
		},
		{
			name:        "post-fork - correctly priced",
			time:        ap3Time,
			fee:         100 * defaultTxFee,
			expectedErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require := require.New(t)

			env := newEnvironment(t, apricotPhase3)
			env.config.UpgradeConfig.ApricotPhase3Time = ap3Time
			env.ctx.Lock.Lock()
			defer env.ctx.Lock.Unlock()

			env.state.SetTimestamp(test.time) // to duly set fee

			addrs := set.NewSet[ids.ShortID](len(preFundedKeys))
			for _, key := range preFundedKeys {
				addrs.Add(key.Address())
			}

			cfg := *env.config
			cfg.StaticFeeConfig.CreateSubnetTxFee = test.fee
			builder := txstest.NewBuilder(env.ctx, &cfg, env.state)
			tx, err := builder.NewCreateSubnetTx(
				&secp256k1fx.OutputOwners{},
				preFundedKeys,
			)
			require.NoError(err)

			stateDiff, err := state.NewDiff(lastAcceptedID, env)
			require.NoError(err)

			stateDiff.SetTimestamp(test.time)

			executor := StandardTxExecutor{
				Backend: &env.backend,
				State:   stateDiff,
				Tx:      tx,
			}
			err = tx.Unsigned.Visit(&executor)
			require.ErrorIs(err, test.expectedErr)
		})
	}
}

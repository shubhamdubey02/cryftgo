// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/upgrade/upgradetest"
	"github.com/shubhamdubey02/cryftgo/utils/set"
	"github.com/shubhamdubey02/cryftgo/utils/units"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/genesis/genesistest"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/state"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/utxo"
	"github.com/shubhamdubey02/cryftgo/vms/secp256k1fx"
)

func TestCreateSubnetTxAP3FeeChange(t *testing.T) {
	ap3Time := genesistest.DefaultValidatorStartTime.Add(time.Hour)
	tests := []struct {
		name        string
		time        time.Time
		fee         uint64
		expectedErr error
	}{
		{
			name:        "pre-fork - correctly priced",
			time:        genesistest.DefaultValidatorStartTime,
			fee:         0,
			expectedErr: nil,
		},
		{
			name:        "post-fork - incorrectly priced",
			time:        ap3Time,
			fee:         100*defaultTxFee - 1*units.NanoAvax,
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

			env := newEnvironment(t, upgradetest.ApricotPhase3)
			env.config.UpgradeConfig.ApricotPhase3Time = ap3Time
			env.ctx.Lock.Lock()
			defer env.ctx.Lock.Unlock()

			env.state.SetTimestamp(test.time) // to duly set fee

			addrs := set.NewSet[ids.ShortID](len(genesistest.DefaultFundedKeys))
			for _, key := range genesistest.DefaultFundedKeys {
				addrs.Add(key.Address())
			}

			config := *env.config
			config.StaticFeeConfig.CreateSubnetTxFee = test.fee
			wallet := newWallet(t, env, walletConfig{
				config: &config,
			})

			tx, err := wallet.IssueCreateSubnetTx(
				&secp256k1fx.OutputOwners{},
			)
			require.NoError(err)

			stateDiff, err := state.NewDiff(lastAcceptedID, env)
			require.NoError(err)

			stateDiff.SetTimestamp(test.time)

			feeCalculator := state.PickFeeCalculator(env.config, stateDiff)
			executor := StandardTxExecutor{
				Backend:       &env.backend,
				FeeCalculator: feeCalculator,
				State:         stateDiff,
				Tx:            tx,
			}
			err = tx.Unsigned.Visit(&executor)
			require.ErrorIs(err, test.expectedErr)
		})
	}
}

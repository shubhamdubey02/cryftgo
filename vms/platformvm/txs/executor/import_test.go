// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/chains/atomic"
	"github.com/shubhamdubey02/cryftgo/database/prefixdb"
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/utils/crypto/secp256k1"
	"github.com/shubhamdubey02/cryftgo/vms/components/cryft"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/state"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/txs"
	"github.com/shubhamdubey02/cryftgo/vms/secp256k1fx"
	"github.com/shubhamdubey02/cryftgo/wallet/chain/p/builder"
)

var fundedSharedMemoryCalls byte

func TestNewImportTx(t *testing.T) {
	env := newEnvironment(t, apricotPhase5)

	type test struct {
		description   string
		sourceChainID ids.ID
		sharedMemory  atomic.SharedMemory
		sourceKeys    []*secp256k1.PrivateKey
		timestamp     time.Time
		expectedErr   error
	}

	sourceKey, err := secp256k1.NewPrivateKey()
	require.NoError(t, err)

	customAssetID := ids.GenerateTestID()
	// generate a constant random source generator.
	randSrc := rand.NewSource(0)
	tests := []test{
		{
			description:   "can't pay fee",
			sourceChainID: env.ctx.XChainID,
			sharedMemory: fundedSharedMemory(
				t,
				env,
				sourceKey,
				env.ctx.XChainID,
				map[ids.ID]uint64{
					env.ctx.CRYFTAssetID: env.config.StaticFeeConfig.TxFee - 1,
				},
				randSrc,
			),
			sourceKeys:  []*secp256k1.PrivateKey{sourceKey},
			expectedErr: builder.ErrInsufficientFunds,
		},
		{
			description:   "can barely pay fee",
			sourceChainID: env.ctx.XChainID,
			sharedMemory: fundedSharedMemory(
				t,
				env,
				sourceKey,
				env.ctx.XChainID,
				map[ids.ID]uint64{
					env.ctx.CRYFTAssetID: env.config.StaticFeeConfig.TxFee,
				},
				randSrc,
			),
			sourceKeys:  []*secp256k1.PrivateKey{sourceKey},
			expectedErr: nil,
		},
		{
			description:   "attempting to import from C-chain",
			sourceChainID: env.ctx.CChainID,
			sharedMemory: fundedSharedMemory(
				t,
				env,
				sourceKey,
				env.ctx.CChainID,
				map[ids.ID]uint64{
					env.ctx.CRYFTAssetID: env.config.StaticFeeConfig.TxFee,
				},
				randSrc,
			),
			sourceKeys:  []*secp256k1.PrivateKey{sourceKey},
			timestamp:   env.config.UpgradeConfig.ApricotPhase5Time,
			expectedErr: nil,
		},
		{
			description:   "attempting to import non-cryft from X-chain",
			sourceChainID: env.ctx.XChainID,
			sharedMemory: fundedSharedMemory(
				t,
				env,
				sourceKey,
				env.ctx.XChainID,
				map[ids.ID]uint64{
					env.ctx.CRYFTAssetID: env.config.StaticFeeConfig.TxFee,
					customAssetID:        1,
				},
				randSrc,
			),
			sourceKeys:  []*secp256k1.PrivateKey{sourceKey},
			timestamp:   env.config.UpgradeConfig.ApricotPhase5Time,
			expectedErr: nil,
		},
	}

	to := &secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs:     []ids.ShortID{ids.GenerateTestShortID()},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			require := require.New(t)

			env.msm.SharedMemory = tt.sharedMemory
			tx, err := env.txBuilder.NewImportTx(
				tt.sourceChainID,
				to,
				tt.sourceKeys,
			)
			require.ErrorIs(err, tt.expectedErr)
			if tt.expectedErr != nil {
				return
			}
			require.NoError(err)

			unsignedTx := tx.Unsigned.(*txs.ImportTx)
			require.NotEmpty(unsignedTx.ImportedInputs)
			numInputs := len(unsignedTx.Ins) + len(unsignedTx.ImportedInputs)
			require.Equal(len(tx.Creds), numInputs, "should have the same number of credentials as inputs")

			totalIn := uint64(0)
			for _, in := range unsignedTx.Ins {
				totalIn += in.Input().Amount()
			}
			for _, in := range unsignedTx.ImportedInputs {
				totalIn += in.Input().Amount()
			}
			totalOut := uint64(0)
			for _, out := range unsignedTx.Outs {
				totalOut += out.Out.Amount()
			}

			require.Equal(env.config.StaticFeeConfig.TxFee, totalIn-totalOut)

			stateDiff, err := state.NewDiff(lastAcceptedID, env)
			require.NoError(err)

			stateDiff.SetTimestamp(tt.timestamp)

			verifier := StandardTxExecutor{
				Backend: &env.backend,
				State:   stateDiff,
				Tx:      tx,
			}
			require.NoError(tx.Unsigned.Visit(&verifier))
		})
	}
}

// Returns a shared memory where GetDatabase returns a database
// where [recipientKey] has a balance of [amt]
func fundedSharedMemory(
	t *testing.T,
	env *environment,
	sourceKey *secp256k1.PrivateKey,
	peerChain ids.ID,
	assets map[ids.ID]uint64,
	randSrc rand.Source,
) atomic.SharedMemory {
	fundedSharedMemoryCalls++
	m := atomic.NewMemory(prefixdb.New([]byte{fundedSharedMemoryCalls}, env.baseDB))

	sm := m.NewSharedMemory(env.ctx.ChainID)
	peerSharedMemory := m.NewSharedMemory(peerChain)

	for assetID, amt := range assets {
		utxo := &cryft.UTXO{
			UTXOID: cryft.UTXOID{
				TxID:        ids.GenerateTestID(),
				OutputIndex: uint32(randSrc.Int63()),
			},
			Asset: cryft.Asset{ID: assetID},
			Out: &secp256k1fx.TransferOutput{
				Amt: amt,
				OutputOwners: secp256k1fx.OutputOwners{
					Locktime:  0,
					Addrs:     []ids.ShortID{sourceKey.PublicKey().Address()},
					Threshold: 1,
				},
			},
		}
		utxoBytes, err := txs.Codec.Marshal(txs.CodecVersion, utxo)
		require.NoError(t, err)

		inputID := utxo.InputID()
		require.NoError(t, peerSharedMemory.Apply(map[ids.ID]*atomic.Requests{
			env.ctx.ChainID: {
				PutRequests: []*atomic.Element{
					{
						Key:   inputID[:],
						Value: utxoBytes,
						Traits: [][]byte{
							sourceKey.PublicKey().Address().Bytes(),
						},
					},
				},
			},
		}))
	}

	return sm
}

// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package utxo

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cryft-labs/cryftgo/ids"
	"github.com/cryft-labs/cryftgo/snow/snowtest"
	"github.com/cryft-labs/cryftgo/utils/crypto/secp256k1"
	"github.com/cryft-labs/cryftgo/utils/timer/mockable"
	"github.com/cryft-labs/cryftgo/vms/components/cryft"
	"github.com/cryft-labs/cryftgo/vms/components/verify"
	"github.com/cryft-labs/cryftgo/vms/platformvm/stakeable"
	"github.com/cryft-labs/cryftgo/vms/platformvm/txs"
	"github.com/cryft-labs/cryftgo/vms/secp256k1fx"

	safemath "github.com/cryft-labs/cryftgo/utils/math"
)

var _ txs.UnsignedTx = (*dummyUnsignedTx)(nil)

type dummyUnsignedTx struct {
	txs.BaseTx
}

func (*dummyUnsignedTx) Visit(txs.Visitor) error {
	return nil
}

func TestVerifySpendUTXOs(t *testing.T) {
	fx := &secp256k1fx.Fx{}

	require.NoError(t, fx.InitializeVM(&secp256k1fx.TestVM{}))
	require.NoError(t, fx.Bootstrapped())

	ctx := snowtest.Context(t, snowtest.PChainID)

	h := &verifier{
		ctx: ctx,
		clk: &mockable.Clock{},
		fx:  fx,
	}

	// The handler time during a test, unless [chainTimestamp] is set
	now := time.Unix(1607133207, 0)

	unsignedTx := dummyUnsignedTx{
		BaseTx: txs.BaseTx{},
	}
	unsignedTx.SetBytes([]byte{0})

	customAssetID := ids.GenerateTestID()

	// Note that setting [chainTimestamp] also set's the handler's clock.
	// Adjust input/output locktimes accordingly.
	tests := []struct {
		description     string
		utxos           []*cryft.UTXO
		ins             []*cryft.TransferableInput
		outs            []*cryft.TransferableOutput
		creds           []verify.Verifiable
		producedAmounts map[ids.ID]uint64
		expectedErr     error
	}{
		{
			description:     "no inputs, no outputs, no fee",
			utxos:           []*cryft.UTXO{},
			ins:             []*cryft.TransferableInput{},
			outs:            []*cryft.TransferableOutput{},
			creds:           []verify.Verifiable{},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     nil,
		},
		{
			description: "no inputs, no outputs, positive fee",
			utxos:       []*cryft.UTXO{},
			ins:         []*cryft.TransferableInput{},
			outs:        []*cryft.TransferableOutput{},
			creds:       []verify.Verifiable{},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
			},
			expectedErr: ErrInsufficientUnlockedFunds,
		},
		{
			description: "wrong utxo assetID, one input, no outputs, no fee",
			utxos: []*cryft.UTXO{{
				Asset: cryft.Asset{ID: customAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*cryft.TransferableInput{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     errAssetIDMismatch,
		},
		{
			description: "one wrong assetID input, no outputs, no fee",
			utxos: []*cryft.UTXO{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*cryft.TransferableInput{{
				Asset: cryft.Asset{ID: customAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     errAssetIDMismatch,
		},
		{
			description: "one input, one wrong assetID output, no fee",
			utxos: []*cryft.UTXO{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*cryft.TransferableInput{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     ErrInsufficientUnlockedFunds,
		},
		{
			description: "attempt to consume locked output as unlocked",
			utxos: []*cryft.UTXO{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				Out: &stakeable.LockOut{
					Locktime: uint64(now.Add(time.Second).Unix()),
					TransferableOut: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			}},
			ins: []*cryft.TransferableInput{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     errLockedFundsNotMarkedAsLocked,
		},
		{
			description: "attempt to modify locktime",
			utxos: []*cryft.UTXO{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				Out: &stakeable.LockOut{
					Locktime: uint64(now.Add(time.Second).Unix()),
					TransferableOut: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			}},
			ins: []*cryft.TransferableInput{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				In: &stakeable.LockIn{
					Locktime: uint64(now.Unix()),
					TransferableIn: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			}},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     errLocktimeMismatch,
		},
		{
			description: "one input, no outputs, positive fee",
			utxos: []*cryft.UTXO{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*cryft.TransferableInput{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
			},
			expectedErr: nil,
		},
		{
			description: "wrong number of credentials",
			utxos: []*cryft.UTXO{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*cryft.TransferableInput{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs:  []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
			},
			expectedErr: errWrongNumberCredentials,
		},
		{
			description: "wrong number of UTXOs",
			utxos:       []*cryft.UTXO{},
			ins: []*cryft.TransferableInput{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
			},
			expectedErr: errWrongNumberUTXOs,
		},
		{
			description: "invalid credential",
			utxos: []*cryft.UTXO{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*cryft.TransferableInput{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				(*secp256k1fx.Credential)(nil),
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
			},
			expectedErr: secp256k1fx.ErrNilCredential,
		},
		{
			description: "invalid signature",
			utxos: []*cryft.UTXO{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs: []ids.ShortID{
							ids.GenerateTestShortID(),
						},
					},
				},
			}},
			ins: []*cryft.TransferableInput{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
					Input: secp256k1fx.Input{
						SigIndices: []uint32{0},
					},
				},
			}},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{
					Sigs: [][secp256k1.SignatureLen]byte{
						{},
					},
				},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
			},
			expectedErr: secp256k1.ErrInvalidSig,
		},
		{
			description: "one input, no outputs, positive fee",
			utxos: []*cryft.UTXO{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 1,
				},
			}},
			ins: []*cryft.TransferableInput{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				In: &secp256k1fx.TransferInput{
					Amt: 1,
				},
			}},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
			},
			expectedErr: nil,
		},
		{
			description: "locked one input, no outputs, no fee",
			utxos: []*cryft.UTXO{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				Out: &stakeable.LockOut{
					Locktime: uint64(now.Unix()) + 1,
					TransferableOut: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			}},
			ins: []*cryft.TransferableInput{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				In: &stakeable.LockIn{
					Locktime: uint64(now.Unix()) + 1,
					TransferableIn: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			}},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     nil,
		},
		{
			description: "locked one input, no outputs, positive fee",
			utxos: []*cryft.UTXO{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				Out: &stakeable.LockOut{
					Locktime: uint64(now.Unix()) + 1,
					TransferableOut: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			}},
			ins: []*cryft.TransferableInput{{
				Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
				In: &stakeable.LockIn{
					Locktime: uint64(now.Unix()) + 1,
					TransferableIn: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			}},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
			},
			expectedErr: ErrInsufficientUnlockedFunds,
		},
		{
			description: "one locked and one unlocked input, one locked output, positive fee",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Unix()) + 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 1,
						},
					},
				},
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &stakeable.LockIn{
						Locktime: uint64(now.Unix()) + 1,
						TransferableIn: &secp256k1fx.TransferInput{
							Amt: 1,
						},
					},
				},
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Unix()) + 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 1,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
			},
			expectedErr: nil,
		},
		{
			description: "one locked and one unlocked input, one locked output, positive fee, partially locked",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Unix()) + 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 1,
						},
					},
				},
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 2,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &stakeable.LockIn{
						Locktime: uint64(now.Unix()) + 1,
						TransferableIn: &secp256k1fx.TransferInput{
							Amt: 1,
						},
					},
				},
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 2,
					},
				},
			},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Unix()) + 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 2,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
			},
			expectedErr: nil,
		},
		{
			description: "one unlocked input, one locked output, zero fee",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Unix()) - 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 1,
						},
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     nil,
		},
		{
			description: "attempted overflow",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 2,
					},
				},
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: math.MaxUint64,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     safemath.ErrOverflow,
		},
		{
			description: "attempted mint",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &stakeable.LockOut{
						Locktime: 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 2,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     ErrInsufficientLockedFunds,
		},
		{
			description: "attempted mint through locking",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &stakeable.LockOut{
						Locktime: 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 2,
						},
					},
				},
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &stakeable.LockOut{
						Locktime: 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: math.MaxUint64,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     safemath.ErrOverflow,
		},
		{
			description: "attempted mint through mixed locking (low then high)",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 2,
					},
				},
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &stakeable.LockOut{
						Locktime: 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: math.MaxUint64,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     ErrInsufficientLockedFunds,
		},
		{
			description: "attempted mint through mixed locking (high then low)",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: math.MaxUint64,
					},
				},
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &stakeable.LockOut{
						Locktime: 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 2,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     ErrInsufficientLockedFunds,
		},
		{
			description: "transfer non-cryft asset",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     nil,
		},
		{
			description: "lock non-cryft asset",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Add(time.Second).Unix()),
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 1,
						},
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     nil,
		},
		{
			description: "attempted asset conversion",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{},
			expectedErr:     ErrInsufficientUnlockedFunds,
		},
		{
			description: "attempted asset conversion with burn",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
			},
			expectedErr: ErrInsufficientUnlockedFunds,
		},
		{
			description: "two inputs, one output with custom asset, with fee",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
				{
					Asset: cryft.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
			},
			expectedErr: nil,
		},
		{
			description: "one input, fee, custom asset",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
			},
			expectedErr: ErrInsufficientUnlockedFunds,
		},
		{
			description: "one input, custom fee",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				customAssetID: 1,
			},
			expectedErr: nil,
		},
		{
			description: "one input, custom fee, wrong burn",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				customAssetID: 1,
			},
			expectedErr: ErrInsufficientUnlockedFunds,
		},
		{
			description: "two inputs, multiple fee",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: h.ctx.CRYFTAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
				{
					Asset: cryft.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
				&secp256k1fx.Credential{},
			},
			producedAmounts: map[ids.ID]uint64{
				h.ctx.CRYFTAssetID: 1,
				customAssetID:     1,
			},
			expectedErr: nil,
		},
		{
			description: "one unlock input, one locked output, zero fee, unlocked, custom asset",
			utxos: []*cryft.UTXO{
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &stakeable.LockOut{
						Locktime: uint64(now.Unix()) - 1,
						TransferableOut: &secp256k1fx.TransferOutput{
							Amt: 1,
						},
					},
				},
			},
			ins: []*cryft.TransferableInput{
				{
					Asset: cryft.Asset{ID: customAssetID},
					In: &secp256k1fx.TransferInput{
						Amt: 1,
					},
				},
			},
			outs: []*cryft.TransferableOutput{
				{
					Asset: cryft.Asset{ID: customAssetID},
					Out: &secp256k1fx.TransferOutput{
						Amt: 1,
					},
				},
			},
			creds: []verify.Verifiable{
				&secp256k1fx.Credential{},
			},
			producedAmounts: make(map[ids.ID]uint64),
			expectedErr:     nil,
		},
	}

	for _, test := range tests {
		h.clk.Set(now)

		t.Run(test.description, func(t *testing.T) {
			err := h.VerifySpendUTXOs(
				&unsignedTx,
				test.utxos,
				test.ins,
				test.outs,
				test.creds,
				test.producedAmounts,
			)
			require.ErrorIs(t, err, test.expectedErr)
		})
	}
}

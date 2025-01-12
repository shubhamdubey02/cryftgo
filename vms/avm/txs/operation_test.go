// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/codec"
	"github.com/shubhamdubey02/cryftgo/codec/linearcodec"
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/snow"
	"github.com/shubhamdubey02/cryftgo/vms/components/cryft"
	"github.com/shubhamdubey02/cryftgo/vms/components/verify"
)

type testOperable struct {
	cryft.TestTransferable `serialize:"true"`

	Outputs []verify.State `serialize:"true"`
}

func (*testOperable) InitCtx(*snow.Context) {}

func (o *testOperable) Outs() []verify.State {
	return o.Outputs
}

func TestOperationVerifyNil(t *testing.T) {
	op := (*Operation)(nil)
	err := op.Verify()
	require.ErrorIs(t, err, ErrNilOperation)
}

func TestOperationVerifyEmpty(t *testing.T) {
	op := &Operation{
		Asset: cryft.Asset{ID: ids.Empty},
	}
	err := op.Verify()
	require.ErrorIs(t, err, ErrNilFxOperation)
}

func TestOperationVerifyUTXOIDsNotSorted(t *testing.T) {
	op := &Operation{
		Asset: cryft.Asset{ID: ids.Empty},
		UTXOIDs: []*cryft.UTXOID{
			{
				TxID:        ids.Empty,
				OutputIndex: 1,
			},
			{
				TxID:        ids.Empty,
				OutputIndex: 0,
			},
		},
		Op: &testOperable{},
	}
	err := op.Verify()
	require.ErrorIs(t, err, ErrNotSortedAndUniqueUTXOIDs)
}

func TestOperationVerify(t *testing.T) {
	assetID := ids.GenerateTestID()
	op := &Operation{
		Asset: cryft.Asset{ID: assetID},
		UTXOIDs: []*cryft.UTXOID{
			{
				TxID:        assetID,
				OutputIndex: 1,
			},
		},
		Op: &testOperable{},
	}
	require.NoError(t, op.Verify())
}

func TestOperationSorting(t *testing.T) {
	require := require.New(t)

	c := linearcodec.NewDefault()
	require.NoError(c.RegisterType(&testOperable{}))

	m := codec.NewDefaultManager()
	require.NoError(m.RegisterCodec(CodecVersion, c))

	ops := []*Operation{
		{
			Asset: cryft.Asset{ID: ids.Empty},
			UTXOIDs: []*cryft.UTXOID{
				{
					TxID:        ids.Empty,
					OutputIndex: 1,
				},
			},
			Op: &testOperable{},
		},
		{
			Asset: cryft.Asset{ID: ids.Empty},
			UTXOIDs: []*cryft.UTXOID{
				{
					TxID:        ids.Empty,
					OutputIndex: 0,
				},
			},
			Op: &testOperable{},
		},
	}
	require.False(IsSortedAndUniqueOperations(ops, m))
	SortOperations(ops, m)
	require.True(IsSortedAndUniqueOperations(ops, m))
	ops = append(ops, &Operation{
		Asset: cryft.Asset{ID: ids.Empty},
		UTXOIDs: []*cryft.UTXOID{
			{
				TxID:        ids.Empty,
				OutputIndex: 1,
			},
		},
		Op: &testOperable{},
	})
	require.False(IsSortedAndUniqueOperations(ops, m))
}

func TestOperationTxNotState(t *testing.T) {
	intf := interface{}(&OperationTx{})
	_, ok := intf.(verify.State)
	require.False(t, ok)
}

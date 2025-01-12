// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/database"
	"github.com/shubhamdubey02/cryftgoftgo/database/memdb"
	"github.com/shubhamdubey02/cryftgoftgo/database/versiondb"
	"github.com/shubhamdubey02/cryftgoftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgoftgo/utils/crypto/secp256k1"
	"github.com/shubhamdubey02/cryftgoftgo/utils/units"
	"github.com/shubhamdubey02/cryftgoftgo/vms/avm/block"
	"github.com/shubhamdubey02/cryftgoftgo/vms/avm/fxs"
	"github.com/shubhamdubey02/cryftgoftgo/vms/avm/state"
	"github.com/shubhamdubey02/cryftgoftgo/vms/avm/txs"
	"github.com/shubhamdubey02/cryftgoftgo/vms/components/cryft"
	"github.com/shubhamdubey02/cryftgoftgo/vms/components/verify"
	"github.com/shubhamdubey02/cryftgoftgo/vms/secp256k1fx"
)

const trackChecksums = false

var (
	chainID = ids.ID{5, 4, 3, 2, 1}
	assetID = ids.ID{1, 2, 3}
)

func TestBaseTxExecutor(t *testing.T) {
	require := require.New(t)

	secpFx := &secp256k1fx.Fx{}
	parser, err := block.NewParser(
		[]fxs.Fx{secpFx},
	)
	require.NoError(err)
	codec := parser.Codec()

	db := memdb.New()
	vdb := versiondb.New(db)
	registerer := prometheus.NewRegistry()
	state, err := state.New(vdb, parser, registerer, trackChecksums)
	require.NoError(err)

	utxoID := cryft.UTXOID{
		TxID:        ids.GenerateTestID(),
		OutputIndex: 1,
	}

	addr := keys[0].Address()
	utxo := &cryft.UTXO{
		UTXOID: utxoID,
		Asset:  cryft.Asset{ID: assetID},
		Out: &secp256k1fx.TransferOutput{
			Amt: 20 * units.KiloCryft,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
	}

	// Populate the UTXO that we will be consuming
	state.AddUTXO(utxo)
	require.NoError(state.Commit())

	baseTx := &txs.Tx{Unsigned: &txs.BaseTx{BaseTx: cryft.BaseTx{
		NetworkID:    constants.UnitTestID,
		BlockchainID: chainID,
		Ins: []*cryft.TransferableInput{{
			UTXOID: utxoID,
			Asset:  cryft.Asset{ID: assetID},
			In: &secp256k1fx.TransferInput{
				Amt: 20 * units.KiloCryft,
				Input: secp256k1fx.Input{
					SigIndices: []uint32{
						0,
					},
				},
			},
		}},
		Outs: []*cryft.TransferableOutput{{
			Asset: cryft.Asset{ID: assetID},
			Out: &secp256k1fx.TransferOutput{
				Amt: 10 * units.KiloCryft,
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
			},
		}},
	}}}
	require.NoError(baseTx.SignSECP256K1Fx(codec, [][]*secp256k1.PrivateKey{{keys[0]}}))

	executor := &Executor{
		Codec: codec,
		State: state,
		Tx:    baseTx,
	}

	// Execute baseTx
	require.NoError(baseTx.Unsigned.Visit(executor))

	// Verify the consumed UTXO was removed from the state
	_, err = executor.State.GetUTXO(utxoID.InputID())
	require.ErrorIs(err, database.ErrNotFound)

	// Verify the produced UTXO was added to the state
	expectedOutputUTXO := &cryft.UTXO{
		UTXOID: cryft.UTXOID{
			TxID:        baseTx.TxID,
			OutputIndex: 0,
		},
		Asset: cryft.Asset{
			ID: assetID,
		},
		Out: &secp256k1fx.TransferOutput{
			Amt: 10 * units.KiloCryft,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs:     []ids.ShortID{addr},
			},
		},
	}
	expectedOutputUTXOID := expectedOutputUTXO.InputID()
	outputUTXO, err := executor.State.GetUTXO(expectedOutputUTXOID)
	require.NoError(err)

	outputUTXOID := outputUTXO.InputID()
	require.Equal(expectedOutputUTXOID, outputUTXOID)
	require.Equal(expectedOutputUTXO, outputUTXO)
}

func TestCreateAssetTxExecutor(t *testing.T) {
	require := require.New(t)

	secpFx := &secp256k1fx.Fx{}
	parser, err := block.NewParser(
		[]fxs.Fx{secpFx},
	)
	require.NoError(err)
	codec := parser.Codec()

	db := memdb.New()
	vdb := versiondb.New(db)
	registerer := prometheus.NewRegistry()
	state, err := state.New(vdb, parser, registerer, trackChecksums)
	require.NoError(err)

	utxoID := cryft.UTXOID{
		TxID:        ids.GenerateTestID(),
		OutputIndex: 1,
	}

	addr := keys[0].Address()
	utxo := &cryft.UTXO{
		UTXOID: utxoID,
		Asset:  cryft.Asset{ID: assetID},
		Out: &secp256k1fx.TransferOutput{
			Amt: 20 * units.KiloCryft,
			OutputOwners: secp256k1fx.OutputOwners{
				Threshold: 1,
				Addrs: []ids.ShortID{
					addr,
				},
			},
		},
	}

	// Populate the UTXO that we will be consuming
	state.AddUTXO(utxo)
	require.NoError(state.Commit())

	createAssetTx := &txs.Tx{Unsigned: &txs.CreateAssetTx{
		BaseTx: txs.BaseTx{BaseTx: cryft.BaseTx{
			NetworkID:    constants.UnitTestID,
			BlockchainID: chainID,
			Ins: []*cryft.TransferableInput{{
				UTXOID: utxoID,
				Asset:  cryft.Asset{ID: assetID},
				In: &secp256k1fx.TransferInput{
					Amt: 20 * units.KiloCryft,
					Input: secp256k1fx.Input{
						SigIndices: []uint32{
							0,
						},
					},
				},
			}},
			Outs: []*cryft.TransferableOutput{{
				Asset: cryft.Asset{ID: assetID},
				Out: &secp256k1fx.TransferOutput{
					Amt: 10 * units.KiloCryft,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs:     []ids.ShortID{addr},
					},
				},
			}},
		}},
		Name:         "name",
		Symbol:       "symb",
		Denomination: 0,
		States: []*txs.InitialState{
			{
				FxIndex: 0,
				Outs: []verify.State{
					&secp256k1fx.MintOutput{
						OutputOwners: secp256k1fx.OutputOwners{
							Threshold: 1,
							Addrs:     []ids.ShortID{addr},
						},
					},
				},
			},
		},
	}}
	require.NoError(createAssetTx.SignSECP256K1Fx(codec, [][]*secp256k1.PrivateKey{{keys[0]}}))

	executor := &Executor{
		Codec: codec,
		State: state,
		Tx:    createAssetTx,
	}

	// Execute createAssetTx
	require.NoError(createAssetTx.Unsigned.Visit(executor))

	// Verify the consumed UTXO was removed from the state
	_, err = executor.State.GetUTXO(utxoID.InputID())
	require.ErrorIs(err, database.ErrNotFound)

	// Verify the produced UTXOs were added to the state
	txID := createAssetTx.ID()
	expectedOutputUTXOs := []*cryft.UTXO{
		{
			UTXOID: cryft.UTXOID{
				TxID:        txID,
				OutputIndex: 0,
			},
			Asset: cryft.Asset{
				ID: assetID,
			},
			Out: &secp256k1fx.TransferOutput{
				Amt: 10 * units.KiloCryft,
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
			},
		},
		{
			UTXOID: cryft.UTXOID{
				TxID:        txID,
				OutputIndex: 1,
			},
			Asset: cryft.Asset{
				ID: txID,
			},
			Out: &secp256k1fx.MintOutput{
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{addr},
				},
			},
		},
	}
	for _, expectedOutputUTXO := range expectedOutputUTXOs {
		expectedOutputUTXOID := expectedOutputUTXO.InputID()
		outputUTXO, err := executor.State.GetUTXO(expectedOutputUTXOID)
		require.NoError(err)

		outputUTXOID := outputUTXO.InputID()
		require.Equal(expectedOutputUTXOID, outputUTXOID)
		require.Equal(expectedOutputUTXO, outputUTXO)
	}
}

func TestOperationTxExecutor(t *testing.T) {
	require := require.New(t)

	secpFx := &secp256k1fx.Fx{}
	parser, err := block.NewParser(
		[]fxs.Fx{secpFx},
	)
	require.NoError(err)
	codec := parser.Codec()

	db := memdb.New()
	vdb := versiondb.New(db)
	registerer := prometheus.NewRegistry()
	state, err := state.New(vdb, parser, registerer, trackChecksums)
	require.NoError(err)

	outputOwners := secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs: []ids.ShortID{
			keys[0].Address(),
		},
	}

	utxoID := cryft.UTXOID{
		TxID:        ids.GenerateTestID(),
		OutputIndex: 1,
	}
	utxo := &cryft.UTXO{
		UTXOID: utxoID,
		Asset:  cryft.Asset{ID: assetID},
		Out: &secp256k1fx.TransferOutput{
			Amt:          20 * units.KiloCryft,
			OutputOwners: outputOwners,
		},
	}

	opUTXOID := cryft.UTXOID{
		TxID:        ids.GenerateTestID(),
		OutputIndex: 1,
	}
	opUTXO := &cryft.UTXO{
		UTXOID: opUTXOID,
		Asset:  cryft.Asset{ID: assetID},
		Out: &secp256k1fx.MintOutput{
			OutputOwners: outputOwners,
		},
	}

	// Populate the UTXOs that we will be consuming
	state.AddUTXO(utxo)
	state.AddUTXO(opUTXO)
	require.NoError(state.Commit())

	operationTx := &txs.Tx{Unsigned: &txs.OperationTx{
		BaseTx: txs.BaseTx{BaseTx: cryft.BaseTx{
			NetworkID:    constants.UnitTestID,
			BlockchainID: chainID,
			Ins: []*cryft.TransferableInput{{
				UTXOID: utxoID,
				Asset:  cryft.Asset{ID: assetID},
				In: &secp256k1fx.TransferInput{
					Amt: 20 * units.KiloCryft,
					Input: secp256k1fx.Input{
						SigIndices: []uint32{
							0,
						},
					},
				},
			}},
			Outs: []*cryft.TransferableOutput{{
				Asset: cryft.Asset{ID: assetID},
				Out: &secp256k1fx.TransferOutput{
					Amt:          10 * units.KiloCryft,
					OutputOwners: outputOwners,
				},
			}},
		}},
		Ops: []*txs.Operation{{
			Asset: cryft.Asset{ID: assetID},
			UTXOIDs: []*cryft.UTXOID{
				&opUTXOID,
			},
			Op: &secp256k1fx.MintOperation{
				MintInput: secp256k1fx.Input{
					SigIndices: []uint32{0},
				},
				MintOutput: secp256k1fx.MintOutput{
					OutputOwners: outputOwners,
				},
				TransferOutput: secp256k1fx.TransferOutput{
					Amt:          12345,
					OutputOwners: outputOwners,
				},
			},
		}},
	}}
	require.NoError(operationTx.SignSECP256K1Fx(
		codec,
		[][]*secp256k1.PrivateKey{
			{keys[0]},
			{keys[0]},
		},
	))

	executor := &Executor{
		Codec: codec,
		State: state,
		Tx:    operationTx,
	}

	// Execute operationTx
	require.NoError(operationTx.Unsigned.Visit(executor))

	// Verify the consumed UTXOs were removed from the state
	_, err = executor.State.GetUTXO(utxo.InputID())
	require.ErrorIs(err, database.ErrNotFound)
	_, err = executor.State.GetUTXO(opUTXO.InputID())
	require.ErrorIs(err, database.ErrNotFound)

	// Verify the produced UTXOs were added to the state
	txID := operationTx.ID()
	expectedOutputUTXOs := []*cryft.UTXO{
		{
			UTXOID: cryft.UTXOID{
				TxID:        txID,
				OutputIndex: 0,
			},
			Asset: cryft.Asset{
				ID: assetID,
			},
			Out: &secp256k1fx.TransferOutput{
				Amt:          10 * units.KiloCryft,
				OutputOwners: outputOwners,
			},
		},
		{
			UTXOID: cryft.UTXOID{
				TxID:        txID,
				OutputIndex: 1,
			},
			Asset: cryft.Asset{
				ID: assetID,
			},
			Out: &secp256k1fx.MintOutput{
				OutputOwners: outputOwners,
			},
		},
		{
			UTXOID: cryft.UTXOID{
				TxID:        txID,
				OutputIndex: 2,
			},
			Asset: cryft.Asset{
				ID: assetID,
			},
			Out: &secp256k1fx.TransferOutput{
				Amt:          12345,
				OutputOwners: outputOwners,
			},
		},
	}
	for _, expectedOutputUTXO := range expectedOutputUTXOs {
		expectedOutputUTXOID := expectedOutputUTXO.InputID()
		outputUTXO, err := executor.State.GetUTXO(expectedOutputUTXOID)
		require.NoError(err)

		outputUTXOID := outputUTXO.InputID()
		require.Equal(expectedOutputUTXOID, outputUTXOID)
		require.Equal(expectedOutputUTXO, outputUTXO)
	}
}

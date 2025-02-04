// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgo/utils/crypto/secp256k1"
	"github.com/shubhamdubey02/cryftgo/vms/avm/txs"
	"github.com/shubhamdubey02/cryftgo/vms/components/cryft"
	"github.com/shubhamdubey02/cryftgo/vms/components/verify"
	"github.com/shubhamdubey02/cryftgo/vms/nftfx"
	"github.com/shubhamdubey02/cryftgo/vms/secp256k1fx"
)

func TestVerifyFxUsage(t *testing.T) {
	require := require.New(t)

	env := setup(t, &envConfig{
		vmStaticConfig: noFeesTestConfig,
	})
	env.vm.ctx.Lock.Unlock()
	defer func() {
		env.vm.ctx.Lock.Lock()
		require.NoError(env.vm.Shutdown(context.Background()))
		env.vm.ctx.Lock.Unlock()
	}()

	createAssetTx := &txs.Tx{Unsigned: &txs.CreateAssetTx{
		BaseTx: txs.BaseTx{BaseTx: cryft.BaseTx{
			NetworkID:    constants.UnitTestID,
			BlockchainID: env.vm.ctx.XChainID,
		}},
		Name:         "Team Rocket",
		Symbol:       "TR",
		Denomination: 0,
		States: []*txs.InitialState{
			{
				FxIndex: 0,
				Outs: []verify.State{
					&secp256k1fx.TransferOutput{
						Amt: 1,
						OutputOwners: secp256k1fx.OutputOwners{
							Threshold: 1,
							Addrs:     []ids.ShortID{keys[0].PublicKey().Address()},
						},
					},
				},
			},
			{
				FxIndex: 1,
				Outs: []verify.State{
					&nftfx.MintOutput{
						GroupID: 1,
						OutputOwners: secp256k1fx.OutputOwners{
							Threshold: 1,
							Addrs:     []ids.ShortID{keys[0].PublicKey().Address()},
						},
					},
				},
			},
		},
	}}
	require.NoError(createAssetTx.Initialize(env.vm.parser.Codec()))
	issueAndAccept(require, env.vm, env.issuer, createAssetTx)

	mintNFTTx := &txs.Tx{Unsigned: &txs.OperationTx{
		BaseTx: txs.BaseTx{BaseTx: cryft.BaseTx{
			NetworkID:    constants.UnitTestID,
			BlockchainID: env.vm.ctx.XChainID,
		}},
		Ops: []*txs.Operation{{
			Asset: cryft.Asset{ID: createAssetTx.ID()},
			UTXOIDs: []*cryft.UTXOID{{
				TxID:        createAssetTx.ID(),
				OutputIndex: 1,
			}},
			Op: &nftfx.MintOperation{
				MintInput: secp256k1fx.Input{
					SigIndices: []uint32{0},
				},
				GroupID: 1,
				Payload: []byte{'h', 'e', 'l', 'l', 'o'},
				Outputs: []*secp256k1fx.OutputOwners{{}},
			},
		}},
	}}
	require.NoError(mintNFTTx.SignNFTFx(env.vm.parser.Codec(), [][]*secp256k1.PrivateKey{{keys[0]}}))
	issueAndAccept(require, env.vm, env.issuer, mintNFTTx)

	spendTx := &txs.Tx{Unsigned: &txs.BaseTx{BaseTx: cryft.BaseTx{
		NetworkID:    constants.UnitTestID,
		BlockchainID: env.vm.ctx.XChainID,
		Ins: []*cryft.TransferableInput{{
			UTXOID: cryft.UTXOID{
				TxID:        createAssetTx.ID(),
				OutputIndex: 0,
			},
			Asset: cryft.Asset{ID: createAssetTx.ID()},
			In: &secp256k1fx.TransferInput{
				Amt: 1,
				Input: secp256k1fx.Input{
					SigIndices: []uint32{0},
				},
			},
		}},
	}}}
	require.NoError(spendTx.SignSECP256K1Fx(env.vm.parser.Codec(), [][]*secp256k1.PrivateKey{{keys[0]}}))
	issueAndAccept(require, env.vm, env.issuer, spendTx)
}

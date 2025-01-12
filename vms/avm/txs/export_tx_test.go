// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/utils/crypto/secp256k1"
	"github.com/shubhamdubey02/cryftgo/vms/avm/fxs"
	"github.com/shubhamdubey02/cryftgo/vms/components/cryft"
	"github.com/shubhamdubey02/cryftgo/vms/components/verify"
	"github.com/shubhamdubey02/cryftgo/vms/secp256k1fx"
)

func TestExportTxSerialization(t *testing.T) {
	require := require.New(t)

	expected := []byte{
		// Codec version:
		0x00, 0x00,
		// txID:
		0x00, 0x00, 0x00, 0x04,
		// networkID:
		0x00, 0x00, 0x00, 0x02,
		// blockchainID:
		0xff, 0xff, 0xff, 0xff, 0xee, 0xee, 0xee, 0xee,
		0xdd, 0xdd, 0xdd, 0xdd, 0xcc, 0xcc, 0xcc, 0xcc,
		0xbb, 0xbb, 0xbb, 0xbb, 0xaa, 0xaa, 0xaa, 0xaa,
		0x99, 0x99, 0x99, 0x99, 0x88, 0x88, 0x88, 0x88,
		// number of outs:
		0x00, 0x00, 0x00, 0x00,
		// number of inputs:
		0x00, 0x00, 0x00, 0x01,
		// utxoID:
		0x0f, 0x2f, 0x4f, 0x6f, 0x8e, 0xae, 0xce, 0xee,
		0x0d, 0x2d, 0x4d, 0x6d, 0x8c, 0xac, 0xcc, 0xec,
		0x0b, 0x2b, 0x4b, 0x6b, 0x8a, 0xaa, 0xca, 0xea,
		0x09, 0x29, 0x49, 0x69, 0x88, 0xa8, 0xc8, 0xe8,
		// output index
		0x00, 0x00, 0x00, 0x00,
		// assetID:
		0x1f, 0x3f, 0x5f, 0x7f, 0x9e, 0xbe, 0xde, 0xfe,
		0x1d, 0x3d, 0x5d, 0x7d, 0x9c, 0xbc, 0xdc, 0xfc,
		0x1b, 0x3b, 0x5b, 0x7b, 0x9a, 0xba, 0xda, 0xfa,
		0x19, 0x39, 0x59, 0x79, 0x98, 0xb8, 0xd8, 0xf8,
		// input:
		// input ID:
		0x00, 0x00, 0x00, 0x05,
		// amount:
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0xe8,
		// num sig indices:
		0x00, 0x00, 0x00, 0x01,
		// sig index[0]:
		0x00, 0x00, 0x00, 0x00,
		// Memo length:
		0x00, 0x00, 0x00, 0x04,
		// Memo:
		0x00, 0x01, 0x02, 0x03,
		// Destination Chain ID:
		0x1f, 0x8f, 0x9f, 0x0f, 0x1e, 0x8e, 0x9e, 0x0e,
		0x2d, 0x7d, 0xad, 0xfd, 0x2c, 0x7c, 0xac, 0xfc,
		0x3b, 0x6b, 0xbb, 0xeb, 0x3a, 0x6a, 0xba, 0xea,
		0x49, 0x59, 0xc9, 0xd9, 0x48, 0x58, 0xc8, 0xd8,
		// number of exported outs:
		0x00, 0x00, 0x00, 0x00,
		// number of credentials:
		0x00, 0x00, 0x00, 0x00,
	}

	tx := &Tx{Unsigned: &ExportTx{
		BaseTx: BaseTx{BaseTx: cryft.BaseTx{
			NetworkID: 2,
			BlockchainID: ids.ID{
				0xff, 0xff, 0xff, 0xff, 0xee, 0xee, 0xee, 0xee,
				0xdd, 0xdd, 0xdd, 0xdd, 0xcc, 0xcc, 0xcc, 0xcc,
				0xbb, 0xbb, 0xbb, 0xbb, 0xaa, 0xaa, 0xaa, 0xaa,
				0x99, 0x99, 0x99, 0x99, 0x88, 0x88, 0x88, 0x88,
			},
			Ins: []*cryft.TransferableInput{{
				UTXOID: cryft.UTXOID{TxID: ids.ID{
					0x0f, 0x2f, 0x4f, 0x6f, 0x8e, 0xae, 0xce, 0xee,
					0x0d, 0x2d, 0x4d, 0x6d, 0x8c, 0xac, 0xcc, 0xec,
					0x0b, 0x2b, 0x4b, 0x6b, 0x8a, 0xaa, 0xca, 0xea,
					0x09, 0x29, 0x49, 0x69, 0x88, 0xa8, 0xc8, 0xe8,
				}},
				Asset: cryft.Asset{ID: ids.ID{
					0x1f, 0x3f, 0x5f, 0x7f, 0x9e, 0xbe, 0xde, 0xfe,
					0x1d, 0x3d, 0x5d, 0x7d, 0x9c, 0xbc, 0xdc, 0xfc,
					0x1b, 0x3b, 0x5b, 0x7b, 0x9a, 0xba, 0xda, 0xfa,
					0x19, 0x39, 0x59, 0x79, 0x98, 0xb8, 0xd8, 0xf8,
				}},
				In: &secp256k1fx.TransferInput{
					Amt:   1000,
					Input: secp256k1fx.Input{SigIndices: []uint32{0}},
				},
			}},
			Memo: []byte{0x00, 0x01, 0x02, 0x03},
		}},
		DestinationChain: ids.ID{
			0x1f, 0x8f, 0x9f, 0x0f, 0x1e, 0x8e, 0x9e, 0x0e,
			0x2d, 0x7d, 0xad, 0xfd, 0x2c, 0x7c, 0xac, 0xfc,
			0x3b, 0x6b, 0xbb, 0xeb, 0x3a, 0x6a, 0xba, 0xea,
			0x49, 0x59, 0xc9, 0xd9, 0x48, 0x58, 0xc8, 0xd8,
		},
	}}

	parser, err := NewParser(
		[]fxs.Fx{
			&secp256k1fx.Fx{},
		},
	)
	require.NoError(err)

	require.NoError(tx.Initialize(parser.Codec()))
	require.Equal("2PKJE4TrKYpgynBFCpNPpV3GHK7d9QTgrL5mpYG6abHKDvNBG3", tx.ID().String())

	result := tx.Bytes()
	require.Equal(expected, result)

	credBytes := []byte{
		// type id
		0x00, 0x00, 0x00, 0x09,

		// there are two signers (thus two signatures)
		0x00, 0x00, 0x00, 0x02,

		// 65 bytes
		0x61, 0xdd, 0x9b, 0xff, 0xc0, 0x49, 0x95, 0x6e, 0xd7, 0xf8,
		0xcd, 0x92, 0xec, 0xda, 0x03, 0x6e, 0xac, 0xb8, 0x16, 0x9e,
		0x53, 0x83, 0xc0, 0x3a, 0x2e, 0x88, 0x5b, 0x5f, 0xc6, 0xef,
		0x2e, 0xbe, 0x50, 0x59, 0x72, 0x8d, 0x0f, 0xa6, 0x59, 0x66,
		0x93, 0x28, 0x88, 0xb4, 0x56, 0x3b, 0x77, 0x7c, 0x59, 0xa5,
		0x8f, 0xe0, 0x2a, 0xf3, 0xcc, 0x31, 0x32, 0xef, 0xfe, 0x7d,
		0x3d, 0x9f, 0x14, 0x94, 0x01,

		// 65 bytes
		0x61, 0xdd, 0x9b, 0xff, 0xc0, 0x49, 0x95, 0x6e, 0xd7, 0xf8,
		0xcd, 0x92, 0xec, 0xda, 0x03, 0x6e, 0xac, 0xb8, 0x16, 0x9e,
		0x53, 0x83, 0xc0, 0x3a, 0x2e, 0x88, 0x5b, 0x5f, 0xc6, 0xef,
		0x2e, 0xbe, 0x50, 0x59, 0x72, 0x8d, 0x0f, 0xa6, 0x59, 0x66,
		0x93, 0x28, 0x88, 0xb4, 0x56, 0x3b, 0x77, 0x7c, 0x59, 0xa5,
		0x8f, 0xe0, 0x2a, 0xf3, 0xcc, 0x31, 0x32, 0xef, 0xfe, 0x7d,
		0x3d, 0x9f, 0x14, 0x94, 0x01,

		// type id
		0x00, 0x00, 0x00, 0x09,

		// there are two signers (thus two signatures)
		0x00, 0x00, 0x00, 0x02,

		// 65 bytes
		0x61, 0xdd, 0x9b, 0xff, 0xc0, 0x49, 0x95, 0x6e, 0xd7, 0xf8,
		0xcd, 0x92, 0xec, 0xda, 0x03, 0x6e, 0xac, 0xb8, 0x16, 0x9e,
		0x53, 0x83, 0xc0, 0x3a, 0x2e, 0x88, 0x5b, 0x5f, 0xc6, 0xef,
		0x2e, 0xbe, 0x50, 0x59, 0x72, 0x8d, 0x0f, 0xa6, 0x59, 0x66,
		0x93, 0x28, 0x88, 0xb4, 0x56, 0x3b, 0x77, 0x7c, 0x59, 0xa5,
		0x8f, 0xe0, 0x2a, 0xf3, 0xcc, 0x31, 0x32, 0xef, 0xfe, 0x7d,
		0x3d, 0x9f, 0x14, 0x94, 0x01,

		// 65 bytes
		0x61, 0xdd, 0x9b, 0xff, 0xc0, 0x49, 0x95, 0x6e, 0xd7, 0xf8,
		0xcd, 0x92, 0xec, 0xda, 0x03, 0x6e, 0xac, 0xb8, 0x16, 0x9e,
		0x53, 0x83, 0xc0, 0x3a, 0x2e, 0x88, 0x5b, 0x5f, 0xc6, 0xef,
		0x2e, 0xbe, 0x50, 0x59, 0x72, 0x8d, 0x0f, 0xa6, 0x59, 0x66,
		0x93, 0x28, 0x88, 0xb4, 0x56, 0x3b, 0x77, 0x7c, 0x59, 0xa5,
		0x8f, 0xe0, 0x2a, 0xf3, 0xcc, 0x31, 0x32, 0xef, 0xfe, 0x7d,
		0x3d, 0x9f, 0x14, 0x94, 0x01,
	}
	require.NoError(tx.SignSECP256K1Fx(
		parser.Codec(),
		[][]*secp256k1.PrivateKey{
			{keys[0], keys[0]},
			{keys[0], keys[0]},
		},
	))
	require.Equal("2oG52e7Cb7XF1yUzv3pRFndAypgbpswWRcSAKD5SH5VgaiTm5D", tx.ID().String())

	// there are two credentials
	expected[len(expected)-1] = 0x02
	expected = append(expected, credBytes...)
	result = tx.Bytes()
	require.Equal(expected, result)
}

func TestExportTxNotState(t *testing.T) {
	require := require.New(t)

	intf := interface{}(&ExportTx{})
	_, ok := intf.(verify.State)
	require.False(ok, "should not be marked as state")
}

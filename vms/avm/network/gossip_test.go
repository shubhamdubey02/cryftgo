// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package network

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/snow/engine/common"
	"github.com/shubhamdubey02/cryftgo/utils/logging"
	"github.com/shubhamdubey02/cryftgo/vms/avm/fxs"
	"github.com/shubhamdubey02/cryftgo/vms/avm/txs"
	"github.com/shubhamdubey02/cryftgo/vms/avm/txs/mempool"
	"github.com/shubhamdubey02/cryftgo/vms/components/cryft"
	"github.com/shubhamdubey02/cryftgo/vms/secp256k1fx"
)

var _ TxVerifier = (*testVerifier)(nil)

type testVerifier struct {
	err error
}

func (v testVerifier) VerifyTx(*txs.Tx) error {
	return v.err
}

func TestMarshaller(t *testing.T) {
	require := require.New(t)

	parser, err := txs.NewParser(
		[]fxs.Fx{
			&secp256k1fx.Fx{},
		},
	)
	require.NoError(err)

	marhsaller := txParser{
		parser: parser,
	}

	want := &txs.Tx{Unsigned: &txs.BaseTx{}}
	require.NoError(want.Initialize(parser.Codec()))

	bytes, err := marhsaller.MarshalGossip(want)
	require.NoError(err)

	got, err := marhsaller.UnmarshalGossip(bytes)
	require.NoError(err)
	require.Equal(want.GossipID(), got.GossipID())
}

func TestGossipMempoolAdd(t *testing.T) {
	require := require.New(t)

	metrics := prometheus.NewRegistry()
	toEngine := make(chan common.Message, 1)

	baseMempool, err := mempool.New("", metrics, toEngine)
	require.NoError(err)

	parser, err := txs.NewParser(nil)
	require.NoError(err)

	mempool, err := newGossipMempool(
		baseMempool,
		metrics,
		logging.NoLog{},
		testVerifier{},
		parser,
		DefaultConfig.ExpectedBloomFilterElements,
		DefaultConfig.ExpectedBloomFilterFalsePositiveProbability,
		DefaultConfig.MaxBloomFilterFalsePositiveProbability,
	)
	require.NoError(err)

	tx := &txs.Tx{
		Unsigned: &txs.BaseTx{
			BaseTx: cryft.BaseTx{
				Ins: []*cryft.TransferableInput{},
			},
		},
		TxID: ids.GenerateTestID(),
	}

	require.NoError(mempool.Add(tx))
	require.True(mempool.bloom.Has(tx))
}

func TestGossipMempoolAddVerified(t *testing.T) {
	require := require.New(t)

	metrics := prometheus.NewRegistry()
	toEngine := make(chan common.Message, 1)

	baseMempool, err := mempool.New("", metrics, toEngine)
	require.NoError(err)

	parser, err := txs.NewParser(nil)
	require.NoError(err)

	mempool, err := newGossipMempool(
		baseMempool,
		metrics,
		logging.NoLog{},
		testVerifier{
			err: errTest, // We shouldn't be attempting to verify the tx in this flow
		},
		parser,
		DefaultConfig.ExpectedBloomFilterElements,
		DefaultConfig.ExpectedBloomFilterFalsePositiveProbability,
		DefaultConfig.MaxBloomFilterFalsePositiveProbability,
	)
	require.NoError(err)

	tx := &txs.Tx{
		Unsigned: &txs.BaseTx{
			BaseTx: cryft.BaseTx{
				Ins: []*cryft.TransferableInput{},
			},
		},
		TxID: ids.GenerateTestID(),
	}

	require.NoError(mempool.AddWithoutVerification(tx))
	require.True(mempool.bloom.Has(tx))
}

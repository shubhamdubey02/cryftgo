// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txstest

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/snow"
	"github.com/shubhamdubey02/cryftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgo/vms/components/avax"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/config"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/fx"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/state"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/txs"
	"github.com/shubhamdubey02/cryftgo/vms/secp256k1fx"
	"github.com/shubhamdubey02/cryftgo/wallet/chain/p/builder"
	"github.com/shubhamdubey02/cryftgo/wallet/chain/p/signer"
	"github.com/shubhamdubey02/cryftgo/wallet/chain/p/wallet"
	"github.com/shubhamdubey02/cryftgo/wallet/subnet/primary/common"
)

func NewWallet(
	t testing.TB,
	ctx *snow.Context,
	config *config.Config,
	state state.State,
	kc *secp256k1fx.Keychain,
	subnetIDs []ids.ID,
	chainIDs []ids.ID,
) wallet.Wallet {
	var (
		require = require.New(t)
		addrs   = kc.Addresses()
		utxos   = common.NewUTXOs()
	)

	pChainUTXOs, err := avax.GetAllUTXOs(state, addrs)
	require.NoError(err)

	for _, utxo := range pChainUTXOs {
		require.NoError(utxos.AddUTXO(
			context.Background(),
			constants.PlatformChainID,
			constants.PlatformChainID,
			utxo,
		))
	}

	for _, chainID := range chainIDs {
		remoteChainUTXOs, _, _, err := avax.GetAtomicUTXOs(
			ctx.SharedMemory,
			txs.Codec,
			chainID,
			addrs,
			ids.ShortEmpty,
			ids.Empty,
			math.MaxInt,
		)
		require.NoError(err)

		for _, utxo := range remoteChainUTXOs {
			require.NoError(utxos.AddUTXO(
				context.Background(),
				chainID,
				constants.PlatformChainID,
				utxo,
			))
		}
	}

	owners := make(map[ids.ID]fx.Owner, len(subnetIDs))
	for _, subnetID := range subnetIDs {
		owner, err := state.GetSubnetOwner(subnetID)
		require.NoError(err)
		owners[subnetID] = owner
	}

	builderContext := newContext(ctx, config, state)
	backend := wallet.NewBackend(
		builderContext,
		common.NewChainUTXOs(constants.PlatformChainID, utxos),
		owners,
	)
	return wallet.New(
		&client{
			backend: backend,
		},
		builder.New(
			addrs,
			builderContext,
			backend,
		),
		signer.New(
			kc,
			backend,
		),
	)
}

type client struct {
	backend wallet.Backend
}

func (c *client) IssueTx(
	tx *txs.Tx,
	options ...common.Option,
) error {
	ops := common.NewOptions(options)
	if f := ops.PostIssuanceFunc(); f != nil {
		txID := tx.ID()
		f(txID)
	}
	ctx := ops.Context()
	return c.backend.AcceptTx(ctx, tx)
}

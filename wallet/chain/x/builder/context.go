// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package builder

import (
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/snow"
	"github.com/shubhamdubey02/cryftgoftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgoftgo/utils/logging"
)

const Alias = "X"

type Context struct {
	NetworkID        uint32
	BlockchainID     ids.ID
	CRYFTAssetID     ids.ID
	BaseTxFee        uint64
	CreateAssetTxFee uint64
}

func NewSnowContext(
	networkID uint32,
	blockchainID ids.ID,
	cryftAssetID ids.ID,
) (*snow.Context, error) {
	lookup := ids.NewAliaser()
	return &snow.Context{
		NetworkID:    networkID,
		SubnetID:     constants.PrimaryNetworkID,
		ChainID:      blockchainID,
		XChainID:     blockchainID,
		CRYFTAssetID: cryftAssetID,
		Log:          logging.NoLog{},
		BCLookup:     lookup,
	}, lookup.Alias(blockchainID, Alias)
}

// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package builder

import (
	"context"

	"github.com/shubhamdubey02/cryftgo/api/info"
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/snow"
	"github.com/shubhamdubey02/cryftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgo/utils/logging"
	"github.com/shubhamdubey02/cryftgo/vms/avm"
)

const Alias = "P"

type Context struct {
	NetworkID                     uint32
	CRYFTAssetID                  ids.ID
	BaseTxFee                     uint64
	CreateSubnetTxFee             uint64
	TransformSubnetTxFee          uint64
	CreateBlockchainTxFee         uint64
	AddPrimaryNetworkValidatorFee uint64
	AddPrimaryNetworkDelegatorFee uint64
	AddSubnetValidatorFee         uint64
	AddSubnetDelegatorFee         uint64
}

func NewContextFromURI(ctx context.Context, uri string) (*Context, error) {
	infoClient := info.NewClient(uri)
	xChainClient := avm.NewClient(uri, "X")
	return NewContextFromClients(ctx, infoClient, xChainClient)
}

func NewContextFromClients(
	ctx context.Context,
	infoClient info.Client,
	xChainClient avm.Client,
) (*Context, error) {
	networkID, err := infoClient.GetNetworkID(ctx)
	if err != nil {
		return nil, err
	}

	asset, err := xChainClient.GetAssetDescription(ctx, "CRYFT")
	if err != nil {
		return nil, err
	}

	txFees, err := infoClient.GetTxFee(ctx)
	if err != nil {
		return nil, err
	}

	return &Context{
		NetworkID:                     networkID,
		CRYFTAssetID:                  asset.AssetID,
		BaseTxFee:                     uint64(txFees.TxFee),
		CreateSubnetTxFee:             uint64(txFees.CreateSubnetTxFee),
		TransformSubnetTxFee:          uint64(txFees.TransformSubnetTxFee),
		CreateBlockchainTxFee:         uint64(txFees.CreateBlockchainTxFee),
		AddPrimaryNetworkValidatorFee: uint64(txFees.AddPrimaryNetworkValidatorFee),
		AddPrimaryNetworkDelegatorFee: uint64(txFees.AddPrimaryNetworkDelegatorFee),
		AddSubnetValidatorFee:         uint64(txFees.AddSubnetValidatorFee),
		AddSubnetDelegatorFee:         uint64(txFees.AddSubnetDelegatorFee),
	}, nil
}

func NewSnowContext(networkID uint32, cryftAssetID ids.ID) (*snow.Context, error) {
	lookup := ids.NewAliaser()
	return &snow.Context{
		NetworkID:    networkID,
		SubnetID:     constants.PrimaryNetworkID,
		ChainID:      constants.PlatformChainID,
		CRYFTAssetID: cryftAssetID,
		Log:          logging.NoLog{},
		BCLookup:     lookup,
	}, lookup.Alias(constants.PlatformChainID, Alias)
}

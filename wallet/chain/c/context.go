// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package c

import (
	"github.com/shubhamdubey02/cryftgo/api/info"
	"github.com/shubhamdubey02/cryftgoftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/snow"
	"github.com/shubhamdubey02/cryftgoftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgoftgo/utils/logging"
	"github.com/shubhamdubey02/cryftgoftgo/vms/avm"

	stdcontext "context"
)

const Alias = "C"

var _ Context = (*context)(nil)

type Context interface {
	NetworkID() uint32
	BlockchainID() ids.ID
	CRYFTAssetID() ids.ID
}

type context struct {
	networkID    uint32
	blockchainID ids.ID
	cryftAssetID ids.ID
}

func NewContextFromURI(ctx stdcontext.Context, uri string) (Context, error) {
	infoClient := info.NewClient(uri)
	xChainClient := avm.NewClient(uri, "X")
	return NewContextFromClients(ctx, infoClient, xChainClient)
}

func NewContextFromClients(
	ctx stdcontext.Context,
	infoClient info.Client,
	xChainClient avm.Client,
) (Context, error) {
	networkID, err := infoClient.GetNetworkID(ctx)
	if err != nil {
		return nil, err
	}

	chainID, err := infoClient.GetBlockchainID(ctx, Alias)
	if err != nil {
		return nil, err
	}

	asset, err := xChainClient.GetAssetDescription(ctx, "CRYFT")
	if err != nil {
		return nil, err
	}

	return NewContext(
		networkID,
		chainID,
		asset.AssetID,
	), nil
}

func NewContext(
	networkID uint32,
	blockchainID ids.ID,
	cryftAssetID ids.ID,
) Context {
	return &context{
		networkID:    networkID,
		blockchainID: blockchainID,
		cryftAssetID: cryftAssetID,
	}
}

func (c *context) NetworkID() uint32 {
	return c.networkID
}

func (c *context) BlockchainID() ids.ID {
	return c.blockchainID
}

func (c *context) CRYFTAssetID() ids.ID {
	return c.cryftAssetID
}

func newSnowContext(c Context) (*snow.Context, error) {
	chainID := c.BlockchainID()
	lookup := ids.NewAliaser()
	return &snow.Context{
		NetworkID:    c.NetworkID(),
		SubnetID:     constants.PrimaryNetworkID,
		ChainID:      chainID,
		CChainID:     chainID,
		CRYFTAssetID: c.CRYFTAssetID(),
		Log:          logging.NoLog{},
		BCLookup:     lookup,
	}, lookup.Alias(chainID, Alias)
}

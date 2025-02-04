// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package primary

import (
	"context"
	"fmt"

	"github.com/cryft-labs/coreth/ethclient"
	"github.com/cryft-labs/coreth/plugin/evm"

	"github.com/shubhamdubey02/cryftgo/api/info"
	"github.com/shubhamdubey02/cryftgo/codec"
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgo/utils/rpc"
	"github.com/shubhamdubey02/cryftgo/utils/set"
	"github.com/shubhamdubey02/cryftgo/vms/avm"
	"github.com/shubhamdubey02/cryftgo/vms/components/cryft"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/txs"
	"github.com/shubhamdubey02/cryftgo/wallet/chain/c"
	"github.com/shubhamdubey02/cryftgo/wallet/chain/x"

	ethcommon "github.com/ethereum/go-ethereum/common"
	pbuilder "github.com/shubhamdubey02/cryftgo/wallet/chain/p/builder"
	xbuilder "github.com/shubhamdubey02/cryftgo/wallet/chain/x/builder"
	walletcommon "github.com/shubhamdubey02/cryftgo/wallet/subnet/primary/common"
)

const (
	MainnetAPIURI = "https://api.cryft.network"
	MustangAPIURI = "https://api.cryft-test.network"
	LocalAPIURI   = "http://localhost:9650"

	fetchLimit = 1024
)

// TODO: Refactor UTXOClient definition to allow the client implementations to
// perform their own assertions.
var (
	_ UTXOClient = platformvm.Client(nil)
	_ UTXOClient = avm.Client(nil)
)

type UTXOClient interface {
	GetAtomicUTXOs(
		ctx context.Context,
		addrs []ids.ShortID,
		sourceChain string,
		limit uint32,
		startAddress ids.ShortID,
		startUTXOID ids.ID,
		options ...rpc.Option,
	) ([][]byte, ids.ShortID, ids.ID, error)
}

type CRYFTState struct {
	PClient platformvm.Client
	PCTX    *pbuilder.Context
	XClient avm.Client
	XCTX    *xbuilder.Context
	CClient evm.Client
	CCTX    c.Context
	UTXOs   walletcommon.UTXOs
}

func FetchState(
	ctx context.Context,
	uri string,
	addrs set.Set[ids.ShortID],
) (
	*CRYFTState,
	error,
) {
	infoClient := info.NewClient(uri)
	pClient := platformvm.NewClient(uri)
	xClient := avm.NewClient(uri, "X")
	cClient := evm.NewCChainClient(uri)

	pCTX, err := pbuilder.NewContextFromClients(ctx, infoClient, xClient)
	if err != nil {
		return nil, err
	}

	xCTX, err := x.NewContextFromClients(ctx, infoClient, xClient)
	if err != nil {
		return nil, err
	}

	cCTX, err := c.NewContextFromClients(ctx, infoClient, xClient)
	if err != nil {
		return nil, err
	}

	utxos := walletcommon.NewUTXOs()
	addrList := addrs.List()
	chains := []struct {
		id     ids.ID
		client UTXOClient
		codec  codec.Manager
	}{
		{
			id:     constants.PlatformChainID,
			client: pClient,
			codec:  txs.Codec,
		},
		{
			id:     xCTX.BlockchainID,
			client: xClient,
			codec:  xbuilder.Parser.Codec(),
		},
		{
			id:     cCTX.BlockchainID(),
			client: cClient,
			codec:  evm.Codec,
		},
	}
	for _, destinationChain := range chains {
		for _, sourceChain := range chains {
			err = AddAllUTXOs(
				ctx,
				utxos,
				destinationChain.client,
				destinationChain.codec,
				sourceChain.id,
				destinationChain.id,
				addrList,
			)
			if err != nil {
				return nil, err
			}
		}
	}
	return &CRYFTState{
		PClient: pClient,
		PCTX:    pCTX,
		XClient: xClient,
		XCTX:    xCTX,
		CClient: cClient,
		CCTX:    cCTX,
		UTXOs:   utxos,
	}, nil
}

type EthState struct {
	Client   ethclient.Client
	Accounts map[ethcommon.Address]*c.Account
}

func FetchEthState(
	ctx context.Context,
	uri string,
	addrs set.Set[ethcommon.Address],
) (*EthState, error) {
	path := fmt.Sprintf(
		"%s/ext/%s/C/rpc",
		uri,
		constants.ChainAliasPrefix,
	)
	client, err := ethclient.Dial(path)
	if err != nil {
		return nil, err
	}

	accounts := make(map[ethcommon.Address]*c.Account, addrs.Len())
	for addr := range addrs {
		balance, err := client.BalanceAt(ctx, addr, nil)
		if err != nil {
			return nil, err
		}
		nonce, err := client.NonceAt(ctx, addr, nil)
		if err != nil {
			return nil, err
		}
		accounts[addr] = &c.Account{
			Balance: balance,
			Nonce:   nonce,
		}
	}
	return &EthState{
		Client:   client,
		Accounts: accounts,
	}, nil
}

// AddAllUTXOs fetches all the UTXOs referenced by [addresses] that were sent
// from [sourceChainID] to [destinationChainID] from the [client]. It then uses
// [codec] to parse the returned UTXOs and it adds them into [utxos]. If [ctx]
// expires, then the returned error will be immediately reported.
func AddAllUTXOs(
	ctx context.Context,
	utxos walletcommon.UTXOs,
	client UTXOClient,
	codec codec.Manager,
	sourceChainID ids.ID,
	destinationChainID ids.ID,
	addrs []ids.ShortID,
) error {
	var (
		sourceChainIDStr = sourceChainID.String()
		startAddr        ids.ShortID
		startUTXO        ids.ID
	)
	for {
		utxosBytes, endAddr, endUTXO, err := client.GetAtomicUTXOs(
			ctx,
			addrs,
			sourceChainIDStr,
			fetchLimit,
			startAddr,
			startUTXO,
		)
		if err != nil {
			return err
		}

		for _, utxoBytes := range utxosBytes {
			var utxo cryft.UTXO
			_, err := codec.Unmarshal(utxoBytes, &utxo)
			if err != nil {
				return err
			}

			if err := utxos.AddUTXO(ctx, sourceChainID, destinationChainID, &utxo); err != nil {
				return err
			}
		}

		if len(utxosBytes) < fetchLimit {
			break
		}

		// Update the vars to query the next page of UTXOs.
		startAddr = endAddr
		startUTXO = endUTXO
	}
	return nil
}

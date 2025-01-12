// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"log"
	"time"

	"github.com/shubhamdubey02/cryftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgoftgo/utils/formatting/address"
	"github.com/shubhamdubey02/cryftgoftgo/utils/set"
	"github.com/shubhamdubey02/cryftgoftgo/wallet/chain/p"
	"github.com/shubhamdubey02/cryftgoftgo/wallet/chain/p/builder"
	"github.com/shubhamdubey02/cryftgoftgo/wallet/subnet/primary"
	"github.com/shubhamdubey02/cryftgoftgo/wallet/subnet/primary/common"
)

func main() {
	uri := primary.LocalAPIURI
	addrStr := "P-local18jma8ppw3nhx5r4ap8clazz0dps7rv5u00z96u"

	addr, err := address.ParseToID(addrStr)
	if err != nil {
		log.Fatalf("failed to parse address: %s\n", err)
	}

	addresses := set.Of(addr)

	ctx := context.Background()

	fetchStartTime := time.Now()
	state, err := primary.FetchState(ctx, uri, addresses)
	if err != nil {
		log.Fatalf("failed to fetch state: %s\n", err)
	}
	log.Printf("fetched state of %s in %s\n", addrStr, time.Since(fetchStartTime))

	pUTXOs := common.NewChainUTXOs(constants.PlatformChainID, state.UTXOs)
	pBackend := p.NewBackend(state.PCTX, pUTXOs, nil)
	pBuilder := builder.New(addresses, state.PCTX, pBackend)

	currentBalances, err := pBuilder.GetBalance()
	if err != nil {
		log.Fatalf("failed to get the balance: %s\n", err)
	}

	cryftID := state.PCTX.CRYFTAssetID
	cryftBalance := currentBalances[cryftID]
	log.Printf("current CRYFT balance of %s is %d nCRYFT\n", addrStr, cryftBalance)
}

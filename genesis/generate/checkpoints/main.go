// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/indexer"
	"github.com/shubhamdubey02/cryftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgo/utils/perms"
	"github.com/shubhamdubey02/cryftgo/utils/set"
)

const (
	mustangURI = "http://localhost:9650"
	mainnetURI = "http://localhost:9660"

	maxNumCheckpoints = 100
)

var (
	mustangXChainID = ids.FromStringOrPanic("2JVSBoinj9C2J33VntvzYtVJNZdN2NKiwwKjcumHUWEb5DbBrm")
	mustangCChainID = ids.FromStringOrPanic("yH8D7ThNJkxmtkuv2jgBa4P1Rn3Qpr4pPr7QYNfcdoS6k6HWp")
	mainnetXChainID = ids.FromStringOrPanic("2oYMBNV4eNHyqk2fjjV5nVQLDbtmNJzq5s3qs3Lo6ftnC6FByM")
	mainnetCChainID = ids.FromStringOrPanic("2q9e4r6Mu3U68nU1fYjgbR6JvwrRx36CohpAX5UQxse55x1Q5")
)

// This fetches IDs of blocks periodically accepted on the P-chain, X-chain, and
// C-chain on both Mustang and Mainnet.
//
// This expects to be able to communicate with a Mustang node at [mustangURI] and a
// Mainnet node at [mainnetURI]. Both nodes must have the index API enabled.
func main() {
	ctx := context.Background()

	mustangPChainCheckpoints, err := getCheckpoints(ctx, mustangURI, "P")
	if err != nil {
		log.Fatalf("failed to fetch Mustang P-chain checkpoints: %v", err)
	}
	mustangXChainCheckpoints, err := getCheckpoints(ctx, mustangURI, "X")
	if err != nil {
		log.Fatalf("failed to fetch Mustang X-chain checkpoints: %v", err)
	}
	mustangCChainCheckpoints, err := getCheckpoints(ctx, mustangURI, "C")
	if err != nil {
		log.Fatalf("failed to fetch Mustang C-chain checkpoints: %v", err)
	}

	mainnetPChainCheckpoints, err := getCheckpoints(ctx, mainnetURI, "P")
	if err != nil {
		log.Fatalf("failed to fetch Mainnet P-chain checkpoints: %v", err)
	}
	mainnetXChainCheckpoints, err := getCheckpoints(ctx, mainnetURI, "X")
	if err != nil {
		log.Fatalf("failed to fetch Mainnet X-chain checkpoints: %v", err)
	}
	mainnetCChainCheckpoints, err := getCheckpoints(ctx, mainnetURI, "C")
	if err != nil {
		log.Fatalf("failed to fetch Mainnet C-chain checkpoints: %v", err)
	}

	checkpoints := map[string]map[ids.ID]set.Set[ids.ID]{
		constants.MustangName: {
			constants.PlatformChainID: mustangPChainCheckpoints,
			mustangXChainID:           mustangXChainCheckpoints,
			mustangCChainID:           mustangCChainCheckpoints,
		},
		constants.MainnetName: {
			constants.PlatformChainID: mainnetPChainCheckpoints,
			mainnetXChainID:           mainnetXChainCheckpoints,
			mainnetCChainID:           mainnetCChainCheckpoints,
		},
	}
	checkpointsJSON, err := json.MarshalIndent(checkpoints, "", "\t")
	if err != nil {
		log.Fatalf("failed to marshal checkpoints: %v", err)
	}

	if err := perms.WriteFile("checkpoints.json", checkpointsJSON, perms.ReadWrite); err != nil {
		log.Fatalf("failed to write checkpoints: %v", err)
	}
}

func getCheckpoints(
	ctx context.Context,
	uri string,
	chainAlias string,
) (set.Set[ids.ID], error) {
	var (
		chainURI = fmt.Sprintf("%s/ext/index/%s/block", uri, chainAlias)
		client   = indexer.NewClient(chainURI)
	)

	// If there haven't been any blocks accepted, this will return an error.
	_, lastIndex, err := client.GetLastAccepted(ctx)
	if err != nil {
		return nil, err
	}

	var (
		numAccepted = lastIndex + 1
		// interval is rounded up to ensure that the number of checkpoints
		// fetched is at most maxNumCheckpoints.
		interval    = (numAccepted + maxNumCheckpoints - 1) / maxNumCheckpoints
		checkpoints set.Set[ids.ID]
	)
	for index := interval - 1; index <= lastIndex; index += interval {
		container, err := client.GetContainerByIndex(ctx, index)
		if err != nil {
			return nil, err
		}

		checkpoints.Add(container.ID)
	}
	return checkpoints, nil
}

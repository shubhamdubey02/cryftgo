// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"encoding/json"
	"fmt"

	_ "embed"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgo/utils/ips"
	"github.com/shubhamdubey02/cryftgo/utils/sampler"
)

var (
	//go:embed bootstrappers.json
	bootstrappersPerNetworkJSON []byte

	bootstrappersPerNetwork map[string][]Bootstrapper
)

func init() {
	if err := json.Unmarshal(bootstrappersPerNetworkJSON, &bootstrappersPerNetwork); err != nil {
		panic(fmt.Sprintf("failed to decode bootstrappers.json %v", err))
	}
}

// Represents the relationship between the nodeID and the nodeIP.
// The bootstrapper is sometimes called "anchor" or "beacon" node.
type Bootstrapper struct {
	ID ids.NodeID `json:"id"`
	IP ips.IPDesc `json:"ip"`
}

// GetBootstrappers returns all default bootstrappers for the provided network.
func GetBootstrappers(networkID uint32) []Bootstrapper {
	networkName := constants.NetworkIDToNetworkName[networkID]
	return bootstrappersPerNetwork[networkName]
}

// SampleBootstrappers returns the some beacons this node should connect to
func SampleBootstrappers(networkID uint32, count int) []Bootstrapper {
	bootstrappers := GetBootstrappers(networkID)
	count = min(count, len(bootstrappers))

	s := sampler.NewUniform()
	s.Initialize(uint64(len(bootstrappers)))
	indices, _ := s.Sample(count)

	sampled := make([]Bootstrapper, 0, len(indices))
	for _, index := range indices {
		sampled = append(sampled, bootstrappers[int(index)])
	}
	return sampled
}

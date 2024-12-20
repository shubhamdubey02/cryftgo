// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"encoding/json"
	"fmt"

	_ "embed"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgo/utils/set"
)

var (
	//go:embed validators.json
	validatorsPerNetworkJSON []byte

	validatorsPerNetwork map[string]set.Set[ids.NodeID]
)

func init() {
	if err := json.Unmarshal(validatorsPerNetworkJSON, &validatorsPerNetwork); err != nil {
		panic(fmt.Sprintf("failed to decode validators.json: %v", err))
	}
}

// GetValidators returns recent validators for the requested network.
func GetValidators(networkID uint32) set.Set[ids.NodeID] {
	networkName := constants.NetworkIDToNetworkName[networkID]
	return validatorsPerNetwork[networkName]
}

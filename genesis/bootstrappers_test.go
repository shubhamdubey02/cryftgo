// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/utils/constants"
)

func TestSampleBootstrappers(t *testing.T) {
	require := require.New(t)

	for networkID, networkName := range constants.NetworkIDToNetworkName {
		length := 10
		bootstrappers := SampleBootstrappers(networkID, length)
		t.Logf("%s bootstrappers: %+v", networkName, bootstrappers)

		if networkID == constants.MainnetID || networkID == constants.MustangID {
			require.Len(bootstrappers, length)
		}
	}
}

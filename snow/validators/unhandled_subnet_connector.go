// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package validators

import (
	"context"
	"fmt"

	"github.com/shubhamdubey02/cryftgo/ids"
)

var UnhandledSubnetConnector SubnetConnector = &unhandledSubnetConnector{}

type unhandledSubnetConnector struct{}

func (unhandledSubnetConnector) ConnectedSubnet(_ context.Context, nodeID ids.NodeID, subnetID ids.ID) error {
	return fmt.Errorf(
		"unhandled ConnectedSubnet with nodeID=%q and subnetID=%q",
		nodeID,
		subnetID,
	)
}

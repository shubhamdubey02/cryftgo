// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package sender

import (
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/message"
	"github.com/shubhamdubey02/cryftgoftgo/snow/engine/common"
	"github.com/shubhamdubey02/cryftgoftgo/subnets"
	"github.com/shubhamdubey02/cryftgoftgo/utils/set"
)

// ExternalSender sends consensus messages to other validators
// Right now this is implemented in the networking package
type ExternalSender interface {
	Send(
		msg message.OutboundMessage,
		config common.SendConfig,
		subnetID ids.ID,
		allower subnets.Allower,
	) set.Set[ids.NodeID]
}

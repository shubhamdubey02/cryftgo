// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package nftfx

import (
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/vms/fx"
)

const Name = "nftfx"

var (
	_ fx.Factory = (*Factory)(nil)

	// ID that this Fx uses when labeled
	ID = ids.ID{'n', 'f', 't', 'f', 'x'}
)

type Factory struct{}

func (*Factory) New() any {
	return &Fx{}
}

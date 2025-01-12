// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"reflect"

	"github.com/shubhamdubey02/cryftgo/codec"
	"github.com/shubhamdubey02/cryftgoftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/snow"
	"github.com/shubhamdubey02/cryftgoftgo/vms/avm/config"
	"github.com/shubhamdubey02/cryftgoftgo/vms/avm/fxs"
)

type Backend struct {
	Ctx           *snow.Context
	Config        *config.Config
	Fxs           []*fxs.ParsedFx
	TypeToFxIndex map[reflect.Type]int
	Codec         codec.Manager
	// Note: FeeAssetID may be different than ctx.CRYFTAssetID if this AVM is
	// running in a subnet.
	FeeAssetID   ids.ID
	Bootstrapped bool
}

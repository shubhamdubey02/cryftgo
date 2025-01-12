// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package propertyfx

import (
	"github.com/shubhamdubey02/cryftgo/snow"
	"github.com/shubhamdubey02/cryftgoftgo/vms/components/verify"
	"github.com/shubhamdubey02/cryftgoftgo/vms/secp256k1fx"
)

type BurnOperation struct {
	secp256k1fx.Input `serialize:"true"`
}

func (*BurnOperation) InitCtx(*snow.Context) {}

func (*BurnOperation) Outs() []verify.State {
	return nil
}

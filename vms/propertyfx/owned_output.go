// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package propertyfx

import (
	"github.com/shubhamdubey02/cryftgo/vms/components/verify"
	"github.com/shubhamdubey02/cryftgo/vms/secp256k1fx"
)

var _ verify.State = (*OwnedOutput)(nil)

type OwnedOutput struct {
	verify.IsState `json:"-"`

	secp256k1fx.OutputOwners `serialize:"true"`
}

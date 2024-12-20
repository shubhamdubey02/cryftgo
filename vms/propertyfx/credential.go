// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package propertyfx

import "github.com/shubhamdubey02/cryftgo/vms/secp256k1fx"

type Credential struct {
	secp256k1fx.Credential `serialize:"true"`
}

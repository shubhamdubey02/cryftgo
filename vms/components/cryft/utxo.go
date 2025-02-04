// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cryft

import (
	"errors"

	"github.com/shubhamdubey02/cryftgo/vms/components/verify"
)

var (
	errNilUTXO   = errors.New("nil utxo is not valid")
	errEmptyUTXO = errors.New("empty utxo is not valid")

	_ verify.Verifiable = (*UTXO)(nil)
)

type UTXO struct {
	UTXOID `serialize:"true"`
	Asset  `serialize:"true"`

	Out verify.State `serialize:"true" json:"output"`
}

func (utxo *UTXO) Verify() error {
	switch {
	case utxo == nil:
		return errNilUTXO
	case utxo.Out == nil:
		return errEmptyUTXO
	default:
		return verify.All(&utxo.UTXOID, &utxo.Asset, utxo.Out)
	}
}

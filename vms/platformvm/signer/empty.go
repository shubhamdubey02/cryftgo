// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import "github.com/shubhamdubey02/cryftgo/utils/crypto/bls"

var _ Signer = (*Empty)(nil)

type Empty struct{}

func (*Empty) Verify() error {
	return nil
}

func (*Empty) Key() *bls.PublicKey {
	return nil
}

// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowstorm

import (
	"context"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/snow/choices"
	"github.com/shubhamdubey02/cryftgo/utils/set"
)

var _ Tx = (*TestTx)(nil)

// TestTx is a useful test tx
type TestTx struct {
	choices.TestDecidable

	DependenciesV    set.Set[ids.ID]
	DependenciesErrV error
	VerifyV          error
	BytesV           []byte
}

func (t *TestTx) MissingDependencies() (set.Set[ids.ID], error) {
	return t.DependenciesV, t.DependenciesErrV
}

func (t *TestTx) Verify(context.Context) error {
	return t.VerifyV
}

func (t *TestTx) Bytes() []byte {
	return t.BytesV
}

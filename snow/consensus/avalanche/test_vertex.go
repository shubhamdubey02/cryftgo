// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avalanche

import (
	"context"

	"github.com/shubhamdubey02/cryftgo/snow/choices"
	"github.com/shubhamdubey02/cryftgo/snow/consensus/snowstorm"
)

var _ Vertex = (*TestVertex)(nil)

// TestVertex is a useful test vertex
type TestVertex struct {
	choices.TestDecidable

	ParentsV    []Vertex
	ParentsErrV error
	HeightV     uint64
	HeightErrV  error
	TxsV        []snowstorm.Tx
	TxsErrV     error
	BytesV      []byte
}

func (v *TestVertex) Parents() ([]Vertex, error) {
	return v.ParentsV, v.ParentsErrV
}

func (v *TestVertex) Height() (uint64, error) {
	return v.HeightV, v.HeightErrV
}

func (v *TestVertex) Txs(context.Context) ([]snowstorm.Tx, error) {
	return v.TxsV, v.TxsErrV
}

func (v *TestVertex) Bytes() []byte {
	return v.BytesV
}

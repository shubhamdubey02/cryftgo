// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vertex

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/snow/consensus/avalanche"
)

var (
	errParse = errors.New("unexpectedly called Parse")

	_ Parser = (*TestParser)(nil)
)

type TestParser struct {
	T            *testing.T
	CantParseVtx bool
	ParseVtxF    func(context.Context, []byte) (avalanche.Vertex, error)
}

func (p *TestParser) Default(cant bool) {
	p.CantParseVtx = cant
}

func (p *TestParser) ParseVtx(ctx context.Context, b []byte) (avalanche.Vertex, error) {
	if p.ParseVtxF != nil {
		return p.ParseVtxF(ctx, b)
	}
	if p.CantParseVtx && p.T != nil {
		require.FailNow(p.T, errParse.Error())
	}
	return nil, errParse
}

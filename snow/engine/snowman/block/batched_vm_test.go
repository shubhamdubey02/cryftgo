// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package block_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/database"
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/snow/consensus/snowman"
	"github.com/shubhamdubey02/cryftgo/snow/engine/snowman/block/blocktest"
	"github.com/shubhamdubey02/cryftgo/utils/logging"

	. "github.com/shubhamdubey02/cryftgo/snow/engine/snowman/block"
)

var errTest = errors.New("non-nil error")

func TestGetAncestorsDatabaseNotFound(t *testing.T) {
	require := require.New(t)

	vm := &blocktest.VM{}
	someID := ids.GenerateTestID()
	vm.GetBlockF = func(_ context.Context, id ids.ID) (snowman.Block, error) {
		require.Equal(someID, id)
		return nil, database.ErrNotFound
	}
	containers, err := GetAncestors(context.Background(), logging.NoLog{}, vm, someID, 10, 10, 1*time.Second)
	require.NoError(err)
	require.Empty(containers)
}

// TestGetAncestorsPropagatesErrors checks errors other than
// database.ErrNotFound propagate to caller.
func TestGetAncestorsPropagatesErrors(t *testing.T) {
	require := require.New(t)

	vm := &blocktest.VM{}
	someID := ids.GenerateTestID()
	vm.GetBlockF = func(_ context.Context, id ids.ID) (snowman.Block, error) {
		require.Equal(someID, id)
		return nil, errTest
	}
	containers, err := GetAncestors(context.Background(), logging.NoLog{}, vm, someID, 10, 10, 1*time.Second)
	require.Nil(containers)
	require.ErrorIs(err, errTest)
}

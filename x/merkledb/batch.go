// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package merkledb

import "github.com/shubhamdubey02/cryftgo/database"

var _ database.Batch = (*batch)(nil)

type batch struct {
	database.BatchOps

	db *merkleDB
}

// Assumes [b.db.lock] isn't held.
func (b *batch) Write() error {
	return b.db.commitBatch(b.Ops)
}

func (b *batch) Inner() database.Batch {
	return b
}

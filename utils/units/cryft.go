// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package units

// Denominations of value
const (
	NanoCryft  uint64 = 1
	MicroCryft uint64 = 1000 * NanoCryft
	Schmeckle uint64 = 49*MicroCryft + 463*NanoCryft
	MilliCryft uint64 = 1000 * MicroCryft
	Cryft      uint64 = 1000 * MilliCryft
	KiloCryft  uint64 = 1000 * Cryft
	MegaCryft  uint64 = 1000 * KiloCryft
)

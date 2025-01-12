// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package common

import "github.com/shubhamdubey02/cryftgo/ids"

// Fx wraps an instance of a feature extension
type Fx struct {
	ID ids.ID
	Fx interface{}
}

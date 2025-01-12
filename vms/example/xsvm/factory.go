// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package xsvm

import (
	"github.com/shubhamdubey02/cryftgo/utils/logging"
	"github.com/shubhamdubey02/cryftgoftgo/vms"
)

var _ vms.Factory = (*Factory)(nil)

type Factory struct{}

func (*Factory) New(logging.Logger) (interface{}, error) {
	return &VM{}, nil
}

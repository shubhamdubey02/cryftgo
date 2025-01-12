// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package platformvm

import (
	"github.com/shubhamdubey02/cryftgo/utils/logging"
	"github.com/shubhamdubey02/cryftgoftgo/vms"
	"github.com/shubhamdubey02/cryftgoftgo/vms/platformvm/config"
)

var _ vms.Factory = (*Factory)(nil)

// Factory can create new instances of the Platform Chain
type Factory struct {
	config.Config
}

// New returns a new instance of the Platform Chain
func (f *Factory) New(logging.Logger) (interface{}, error) {
	return &VM{Config: f.Config}, nil
}

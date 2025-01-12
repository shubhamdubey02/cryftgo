// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package executor

import (
	"github.com/shubhamdubey02/cryftgo/snow"
	"github.com/shubhamdubey02/cryftgoftgo/snow/uptime"
	"github.com/shubhamdubey02/cryftgoftgo/utils"
	"github.com/shubhamdubey02/cryftgoftgo/utils/timer/mockable"
	"github.com/shubhamdubey02/cryftgoftgo/vms/platformvm/config"
	"github.com/shubhamdubey02/cryftgoftgo/vms/platformvm/fx"
	"github.com/shubhamdubey02/cryftgoftgo/vms/platformvm/reward"
	"github.com/shubhamdubey02/cryftgoftgo/vms/platformvm/utxo"
)

type Backend struct {
	Config       *config.Config
	Ctx          *snow.Context
	Clk          *mockable.Clock
	Fx           fx.Fx
	FlowChecker  utxo.Verifier
	Uptimes      uptime.Calculator
	Rewards      reward.Calculator
	Bootstrapped *utils.Atomic[bool]
}

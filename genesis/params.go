// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"time"

	"github.com/shubhamdubey02/cryftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/reward"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/txs/fee"
)

type StakingConfig struct {
	// Staking uptime requirements
	UptimeRequirement float64 `json:"uptimeRequirement"`
	// Minimum stake, in nCRYFT, required to validate the primary network
	MinValidatorStake uint64 `json:"minValidatorStake"`
	// Maximum stake, in nCRYFT, allowed to be placed on a single validator in
	// the primary network
	MaxValidatorStake uint64 `json:"maxValidatorStake"`
	// Minimum stake, in nCRYFT, that can be delegated on the primary network
	MinDelegatorStake uint64 `json:"minDelegatorStake"`
	// Minimum delegation fee, in the range [0, 1000000], that can be charged
	// for delegation on the primary network.
	MinDelegationFee uint32 `json:"minDelegationFee"`
	// MinStakeDuration is the minimum amount of time a validator can validate
	// for in a single period.
	MinStakeDuration time.Duration `json:"minStakeDuration"`
	// MaxStakeDuration is the maximum amount of time a validator can validate
	// for in a single period.
	MaxStakeDuration time.Duration `json:"maxStakeDuration"`
	// RewardConfig is the config for the reward function.
	RewardConfig reward.Config `json:"rewardConfig"`
}

type Params struct {
	StakingConfig
	fee.StaticConfig
}

func GetTxFeeConfig(networkID uint32) fee.StaticConfig {
	switch networkID {
	case constants.MainnetID:
		return MainnetParams.StaticConfig
	case constants.MustangID:
		return MustangParams.StaticConfig
	case constants.LocalID:
		return LocalParams.StaticConfig
	default:
		return LocalParams.StaticConfig
	}
}

func GetStakingConfig(networkID uint32) StakingConfig {
	switch networkID {
	case constants.MainnetID:
		return MainnetParams.StakingConfig
	case constants.MustangID:
		return MustangParams.StakingConfig
	case constants.LocalID:
		return LocalParams.StakingConfig
	default:
		return LocalParams.StakingConfig
	}
}

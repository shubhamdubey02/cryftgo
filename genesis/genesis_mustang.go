// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"time"

	_ "embed"

	"github.com/cryft-labs/cryftgo/utils/units"
	"github.com/cryft-labs/cryftgo/vms/platformvm/reward"
	"github.com/cryft-labs/cryftgo/vms/platformvm/txs/fee"
)

var (
	//go:embed genesis_mustang.json
	mustangGenesisConfigJSON []byte

	// MustangParams are the params used for the mustang testnet
	MustangParams = Params{
		StaticConfig: fee.StaticConfig{
			TxFee:                         units.MilliCryft,
			CreateAssetTxFee:              10 * units.MilliCryft,
			CreateSubnetTxFee:             100 * units.MilliCryft,
			TransformSubnetTxFee:          1 * units.Cryft,
			CreateBlockchainTxFee:         100 * units.MilliCryft,
			AddPrimaryNetworkValidatorFee: 0,
			AddPrimaryNetworkDelegatorFee: 0,
			AddSubnetValidatorFee:         units.MilliCryft,
			AddSubnetDelegatorFee:         units.MilliCryft,
		},
		StakingConfig: StakingConfig{
			UptimeRequirement: .8, // 80%
			MinValidatorStake: 1 * units.Cryft,
			MaxValidatorStake: 3 * units.MegaCryft,
			MinDelegatorStake: 1 * units.Cryft,
			MinDelegationFee:  20000, // 2%
			MinStakeDuration:  24 * time.Hour,
			MaxStakeDuration:  365 * 24 * time.Hour,
			RewardConfig: reward.Config{
				MaxConsumptionRate: .12 * reward.PercentDenominator,
				MinConsumptionRate: .10 * reward.PercentDenominator,
				MintingPeriod:      365 * 24 * time.Hour,
				SupplyCap:          720 * units.MegaCryft,
			},
		},
	}
)

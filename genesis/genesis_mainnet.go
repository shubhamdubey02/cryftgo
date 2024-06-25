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
	//go:embed genesis_mainnet.json
	mainnetGenesisConfigJSON []byte

	// MainnetParams are the params used for mainnet
	MainnetParams = Params{
		StaticConfig: fee.StaticConfig{
			TxFee:                         units.MilliCryft,
			CreateAssetTxFee:              10 * units.MilliCryft,
			CreateSubnetTxFee:             1 * units.Cryft,
			TransformSubnetTxFee:          10 * units.Cryft,
			CreateBlockchainTxFee:         1 * units.Cryft,
			AddPrimaryNetworkValidatorFee: 0,
			AddPrimaryNetworkDelegatorFee: 0,
			AddSubnetValidatorFee:         units.MilliCryft,
			AddSubnetDelegatorFee:         units.MilliCryft,
		},
		StakingConfig: StakingConfig{
			UptimeRequirement: .8, // 80%
			MinValidatorStake: 2 * units.KiloCryft,
			MaxValidatorStake: 3 * units.MegaCryft,
			MinDelegatorStake: 25 * units.Cryft,
			MinDelegationFee:  20000, // 2%
			MinStakeDuration:  2 * 7 * 24 * time.Hour,
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

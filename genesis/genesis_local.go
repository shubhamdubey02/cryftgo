// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package genesis

import (
	"time"

	_ "embed"

	"github.com/shubhamdubey02/cryftgo/utils/cb58"
	"github.com/shubhamdubey02/cryftgo/utils/crypto/secp256k1"
	"github.com/shubhamdubey02/cryftgo/utils/units"
	"github.com/shubhamdubey02/cryftgo/utils/wrappers"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/reward"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/txs/fee"
)

// PrivateKey-vmRQiZeXEXYMyJhEiqdC2z5JhuDbxL8ix9UVvjgMu2Er1NepE => P-local1g65uqn6t77p656w64023nh8nd9updzmxyymev2
// PrivateKey-ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN => X-local18jma8ppw3nhx5r4ap8clazz0dps7rv5u00z96u
// 56289e99c94b6912bfc12adc093c9b51124f0dc54ac7a766b2bc5ccf558d8027 => 0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC

const (
	VMRQKeyStr          = "vmRQiZeXEXYMyJhEiqdC2z5JhuDbxL8ix9UVvjgMu2Er1NepE"
	VMRQKeyFormattedStr = secp256k1.PrivateKeyPrefix + VMRQKeyStr

	EWOQKeyStr          = "ewoqjP7PxY4yr3iLTpLisriqt94hdyDFNgchSxGGztUrTXtNN"
	EWOQKeyFormattedStr = secp256k1.PrivateKeyPrefix + EWOQKeyStr
)

var (
	VMRQKey *secp256k1.PrivateKey
	EWOQKey *secp256k1.PrivateKey

	//go:embed genesis_local.json
	localGenesisConfigJSON []byte

	// LocalParams are the params used for local networks
	LocalParams = Params{
		StaticConfig: fee.StaticConfig{
			TxFee:                         units.MilliCryft,
			CreateAssetTxFee:              units.MilliCryft,
			CreateSubnetTxFee:             100 * units.MilliCryft,
			TransformSubnetTxFee:          100 * units.MilliCryft,
			CreateBlockchainTxFee:         100 * units.MilliCryft,
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

func init() {
	errs := wrappers.Errs{}
	vmrqBytes, err := cb58.Decode(VMRQKeyStr)
	errs.Add(err)
	ewoqBytes, err := cb58.Decode(EWOQKeyStr)
	errs.Add(err)

	VMRQKey, err = secp256k1.ToPrivateKey(vmrqBytes)
	errs.Add(err)
	EWOQKey, err = secp256k1.ToPrivateKey(ewoqBytes)
	errs.Add(err)

	if errs.Err != nil {
		panic(errs.Err)
	}
}

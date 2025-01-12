// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package common

import (
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/utils/set"
	"github.com/shubhamdubey02/cryftgoftgo/vms/secp256k1fx"
)

// MatchOwners attempts to match a list of addresses up to the provided
// threshold.
func MatchOwners(
	owners *secp256k1fx.OutputOwners,
	addrs set.Set[ids.ShortID],
	minIssuanceTime uint64,
) ([]uint32, bool) {
	if owners.Locktime > minIssuanceTime {
		return nil, false
	}

	sigs := make([]uint32, 0, owners.Threshold)
	for i := uint32(0); i < uint32(len(owners.Addrs)) && uint32(len(sigs)) < owners.Threshold; i++ {
		addr := owners.Addrs[i]
		if addrs.Contains(addr) {
			sigs = append(sigs, i)
		}
	}
	return sigs, uint32(len(sigs)) == owners.Threshold
}

// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package types

import (
	"encoding/json"

	"github.com/shubhamdubey02/cryftgo/utils/formatting"
)

// JSONByteSlice represents [[]byte] that is json marshalled to hex
type JSONByteSlice []byte

func (b JSONByteSlice) MarshalJSON() ([]byte, error) {
	hexData, err := formatting.Encode(formatting.HexNC, b)
	if err != nil {
		return nil, err
	}
	return json.Marshal(hexData)
}

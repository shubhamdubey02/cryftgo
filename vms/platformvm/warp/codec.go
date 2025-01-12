// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package warp

import (
	"math"

	"github.com/shubhamdubey02/cryftgo/codec"
	"github.com/shubhamdubey02/cryftgoftgo/codec/linearcodec"
	"github.com/shubhamdubey02/cryftgoftgo/utils"
)

const CodecVersion = 0

var Codec codec.Manager

func init() {
	Codec = codec.NewManager(math.MaxInt)
	lc := linearcodec.NewDefault()

	err := utils.Err(
		lc.RegisterType(&BitSetSignature{}),
		Codec.RegisterCodec(CodecVersion, lc),
	)
	if err != nil {
		panic(err)
	}
}

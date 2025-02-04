// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package snowball

import "github.com/shubhamdubey02/cryftgo/ids"

var (
	SnowballFactory  Factory = snowballFactory{}
	SnowflakeFactory Factory = snowflakeFactory{}
)

type snowballFactory struct{}

func (snowballFactory) NewNnary(params Parameters, choice ids.ID) Nnary {
	sb := newNnarySnowball(params.AlphaPreference, params.AlphaConfidence, params.Beta, choice)
	return &sb
}

func (snowballFactory) NewUnary(params Parameters) Unary {
	sb := newUnarySnowball(params.AlphaPreference, params.AlphaConfidence, params.Beta)
	return &sb
}

type snowflakeFactory struct{}

func (snowflakeFactory) NewNnary(params Parameters, choice ids.ID) Nnary {
	sf := newNnarySnowflake(params.AlphaPreference, params.AlphaConfidence, params.Beta, choice)
	return &sf
}

func (snowflakeFactory) NewUnary(params Parameters) Unary {
	sf := newUnarySnowflake(params.AlphaPreference, params.AlphaConfidence, params.Beta)
	return &sf
}

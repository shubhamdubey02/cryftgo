// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package linearcodec

import (
	"testing"

	"github.com/shubhamdubey02/cryftgo/codec"
)

func TestVectors(t *testing.T) {
	for _, test := range codec.Tests {
		c := NewDefault()
		test(c, t)
	}
}

func TestMultipleTags(t *testing.T) {
	for _, test := range codec.MultipleTagsTests {
		c := New([]string{"tag1", "tag2"})
		test(c, t)
	}
}

func FuzzStructUnmarshalLinearCodec(f *testing.F) {
	c := NewDefault()
	codec.FuzzStructUnmarshal(c, f)
}

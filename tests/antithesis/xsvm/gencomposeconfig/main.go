// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"log"

	"github.com/shubhamdubey02/cryftgo/genesis"
	"github.com/shubhamdubey02/cryftgo/tests/antithesis"
	"github.com/shubhamdubey02/cryftgo/tests/fixture/subnet"
	"github.com/shubhamdubey02/cryftgo/tests/fixture/tmpnet"
)

const baseImageName = "antithesis-xsvm"

// Creates docker-compose.yml and its associated volumes in the target path.
func main() {
	network := tmpnet.LocalNetworkOrPanic()
	network.Subnets = []*tmpnet.Subnet{
		subnet.NewXSVMOrPanic("xsvm", genesis.VMRQKey, network.Nodes...),
	}
	if err := antithesis.GenerateComposeConfig(network, baseImageName, "" /* runtimePluginDir */); err != nil {
		log.Fatalf("failed to generate compose config: %v", err)
	}
}

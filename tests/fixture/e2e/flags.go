// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package e2e

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/shubhamdubey02/cryftgo/tests/fixture/tmpnet"
)

type FlagVars struct {
	cryftGoExecPath      string
	pluginDir            string
	networkDir           string
	reuseNetwork         bool
	networkShutdownDelay time.Duration
	stopNetwork          bool
	nodeCount            int
}

func (v *FlagVars) CryftGoExecPath() string {
	return v.cryftGoExecPath
}

func (v *FlagVars) PluginDir() string {
	return v.pluginDir
}

func (v *FlagVars) NetworkDir() string {
	if !v.reuseNetwork {
		return ""
	}
	if len(v.networkDir) > 0 {
		return v.networkDir
	}
	return os.Getenv(tmpnet.NetworkDirEnvName)
}

func (v *FlagVars) ReuseNetwork() bool {
	return v.reuseNetwork
}

func (v *FlagVars) NetworkShutdownDelay() time.Duration {
	return v.networkShutdownDelay
}

func (v *FlagVars) StopNetwork() bool {
	return v.stopNetwork
}

func (v *FlagVars) NodeCount() int {
	return v.nodeCount
}

func RegisterFlags() *FlagVars {
	vars := FlagVars{}
	flag.StringVar(
		&vars.cryftGoExecPath,
		"cryftgo-path",
		os.Getenv(tmpnet.CryftGoPathEnvName),
		fmt.Sprintf("cryftgo executable path (required if not using an existing network). Also possible to configure via the %s env variable.", tmpnet.CryftGoPathEnvName),
	)
	flag.StringVar(
		&vars.pluginDir,
		"plugin-dir",
		os.ExpandEnv("$HOME/.cryftgo/plugins"),
		"[optional] the dir containing VM plugins.",
	)
	flag.StringVar(
		&vars.networkDir,
		"network-dir",
		"",
		fmt.Sprintf("[optional] the dir containing the configuration of an existing network to target for testing. Will only be used if --reuse-network is specified. Also possible to configure via the %s env variable.", tmpnet.NetworkDirEnvName),
	)
	flag.BoolVar(
		&vars.reuseNetwork,
		"reuse-network",
		false,
		"[optional] reuse an existing network. If an existing network is not already running, create a new one and leave it running for subsequent usage.",
	)
	flag.DurationVar(
		&vars.networkShutdownDelay,
		"network-shutdown-delay",
		12*time.Second, // Make sure this value takes into account the scrape_interval defined in scripts/run_prometheus.sh
		"[optional] the duration to wait before shutting down the test network at the end of the test run. A value greater than the scrape interval is suggested. 0 avoids waiting for shutdown.",
	)
	flag.BoolVar(
		&vars.stopNetwork,
		"stop-network",
		false,
		"[optional] stop an existing network and exit without executing any tests.",
	)
	flag.IntVar(
		&vars.nodeCount,
		"node-count",
		tmpnet.DefaultNodeCount,
		"number of nodes the network should initially consist of",
	)

	return &vars
}

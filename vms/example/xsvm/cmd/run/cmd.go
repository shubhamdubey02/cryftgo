// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package run

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/shubhamdubey02/cryftgo/vms/example/xsvm"
	"github.com/shubhamdubey02/cryftgo/vms/rpcchainvm"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "xsvm",
		Short: "Runs an XSVM plugin",
		RunE:  runFunc,
	}
}

func runFunc(*cobra.Command, []string) error {
	return rpcchainvm.Serve(context.Background(), &xsvm.VM{})
}

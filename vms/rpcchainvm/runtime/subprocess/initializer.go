// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package subprocess

import (
	"context"
	"fmt"
	"sync"

	"github.com/shubhamdubey02/cryftgo/version"
	"github.com/shubhamdubey02/cryftgo/vms/rpcchainvm/runtime"
)

var _ runtime.Initializer = (*initializer)(nil)

// Subprocess VM Runtime intializer.
type initializer struct {
	once sync.Once
	// Address of the RPC Chain VM server
	vmAddr string
	// Error, if one occurred, during Initialization
	err error
	// Initialized is closed once Initialize is called
	initialized chan struct{}
}

func newInitializer() *initializer {
	return &initializer{
		initialized: make(chan struct{}),
	}
}

func (i *initializer) Initialize(_ context.Context, protocolVersion uint, vmAddr string) error {
	i.once.Do(func() {
		if version.RPCChainVMProtocol != protocolVersion {
			i.err = fmt.Errorf("%w. CryftGo version %s implements RPCChainVM protocol version %d. The VM implements RPCChainVM protocol version %d. Please make sure that there is an exact match of the protocol versions. This can be achieved by updating your VM or running an older/newer version of CryftGo. Please be advised that some virtual machines may not yet support the latest RPCChainVM protocol version",
				runtime.ErrProtocolVersionMismatch,
				version.Current,
				version.RPCChainVMProtocol,
				protocolVersion,
			)
		}
		i.vmAddr = vmAddr
		close(i.initialized)
	})
	return i.err
}

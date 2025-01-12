// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package galiasreader

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/vms/rpcchainvm/grpcutils"

	aliasreaderpb "github.com/shubhamdubey02/cryftgoftgo/proto/pb/aliasreader"
)

func TestInterface(t *testing.T) {
	require := require.New(t)

	for _, test := range ids.AliasTests {
		listener, err := grpcutils.NewListener()
		require.NoError(err)
		serverCloser := grpcutils.ServerCloser{}
		w := ids.NewAliaser()

		server := grpcutils.NewServer()
		aliasreaderpb.RegisterAliasReaderServer(server, NewServer(w))
		serverCloser.Add(server)

		go grpcutils.Serve(listener, server)

		conn, err := grpcutils.Dial(listener.Addr().String())
		require.NoError(err)

		r := NewClient(aliasreaderpb.NewAliasReaderClient(conn))
		test(require, r, w)

		serverCloser.Stop()
		_ = conn.Close()
		_ = listener.Close()
	}
}

// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package gwarp

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgo/utils/crypto/bls"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/warp"
	"github.com/shubhamdubey02/cryftgo/vms/rpcchainvm/grpcutils"

	pb "github.com/shubhamdubey02/cryftgo/proto/pb/warp"
)

type testSigner struct {
	client    *Client
	server    warp.Signer
	sk        *bls.SecretKey
	networkID uint32
	chainID   ids.ID
}

func setupSigner(t testing.TB) *testSigner {
	require := require.New(t)

	sk, err := bls.NewSecretKey()
	require.NoError(err)

	chainID := ids.GenerateTestID()

	s := &testSigner{
		server:    warp.NewSigner(sk, constants.UnitTestID, chainID),
		sk:        sk,
		networkID: constants.UnitTestID,
		chainID:   chainID,
	}

	listener, err := grpcutils.NewListener()
	require.NoError(err)
	serverCloser := grpcutils.ServerCloser{}

	server := grpcutils.NewServer()
	pb.RegisterSignerServer(server, NewServer(s.server))
	serverCloser.Add(server)

	go grpcutils.Serve(listener, server)

	conn, err := grpcutils.Dial(listener.Addr().String())
	require.NoError(err)

	s.client = NewClient(pb.NewSignerClient(conn))

	t.Cleanup(func() {
		serverCloser.Stop()
		_ = conn.Close()
		_ = listener.Close()
	})

	return s
}

func TestInterface(t *testing.T) {
	for name, test := range warp.SignerTests {
		t.Run(name, func(t *testing.T) {
			s := setupSigner(t)
			test(t, s.client, s.sk, s.networkID, s.chainID)
		})
	}
}

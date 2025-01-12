// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package peer

import (
	"context"
	"crypto"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/message"
	"github.com/shubhamdubey02/cryftgoftgo/network/throttling"
	"github.com/shubhamdubey02/cryftgoftgo/snow/networking/router"
	"github.com/shubhamdubey02/cryftgoftgo/snow/networking/tracker"
	"github.com/shubhamdubey02/cryftgoftgo/snow/uptime"
	"github.com/shubhamdubey02/cryftgoftgo/snow/validators"
	"github.com/shubhamdubey02/cryftgoftgo/staking"
	"github.com/shubhamdubey02/cryftgoftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgoftgo/utils/crypto/bls"
	"github.com/shubhamdubey02/cryftgoftgo/utils/ips"
	"github.com/shubhamdubey02/cryftgoftgo/utils/logging"
	"github.com/shubhamdubey02/cryftgoftgo/utils/math/meter"
	"github.com/shubhamdubey02/cryftgoftgo/utils/resource"
	"github.com/shubhamdubey02/cryftgoftgo/utils/set"
	"github.com/shubhamdubey02/cryftgoftgo/version"
)

const maxMessageToSend = 1024

// StartTestPeer provides a simple interface to create a peer that has finished
// the p2p handshake.
//
// This function will generate a new TLS key to use when connecting to the peer.
//
// The returned peer will not throttle inbound or outbound messages.
//
//   - [ctx] provides a way of canceling the connection request.
//   - [ip] is the remote that will be dialed to create the connection.
//   - [networkID] will be sent to the peer during the handshake. If the peer is
//     expecting a different [networkID], the handshake will fail and an error
//     will be returned.
//   - [router] will be called with all non-handshake messages received by the
//     peer.
func StartTestPeer(
	ctx context.Context,
	ip ips.IPPort,
	networkID uint32,
	router router.InboundHandler,
) (Peer, error) {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, constants.NetworkType, ip.String())
	if err != nil {
		return nil, err
	}

	tlsCert, err := staking.NewTLSCert()
	if err != nil {
		return nil, err
	}

	tlsConfg := TLSConfig(*tlsCert, nil)
	clientUpgrader := NewTLSClientUpgrader(
		tlsConfg,
		prometheus.NewCounter(prometheus.CounterOpts{}),
	)

	peerID, conn, cert, err := clientUpgrader.Upgrade(conn)
	if err != nil {
		return nil, err
	}

	mc, err := message.NewCreator(
		logging.NoLog{},
		prometheus.NewRegistry(),
		"",
		constants.DefaultNetworkCompressionType,
		10*time.Second,
	)
	if err != nil {
		return nil, err
	}

	metrics, err := NewMetrics(
		"",
		prometheus.NewRegistry(),
	)
	if err != nil {
		return nil, err
	}

	resourceTracker, err := tracker.NewResourceTracker(
		prometheus.NewRegistry(),
		resource.NoUsage,
		meter.ContinuousFactory{},
		10*time.Second,
	)
	if err != nil {
		return nil, err
	}

	signerIP := ips.NewDynamicIPPort(net.IPv6zero, 1)
	tlsKey := tlsCert.PrivateKey.(crypto.Signer)
	blsKey, err := bls.NewSecretKey()
	if err != nil {
		return nil, err
	}

	peer := Start(
		&Config{
			Metrics:              metrics,
			MessageCreator:       mc,
			Log:                  logging.NoLog{},
			InboundMsgThrottler:  throttling.NewNoInboundThrottler(),
			Network:              TestNetwork,
			Router:               router,
			VersionCompatibility: version.GetCompatibility(networkID),
			MySubnets:            set.Set[ids.ID]{},
			Beacons:              validators.NewManager(),
			Validators:           validators.NewManager(),
			NetworkID:            networkID,
			PingFrequency:        constants.DefaultPingFrequency,
			PongTimeout:          constants.DefaultPingPongTimeout,
			MaxClockDifference:   time.Minute,
			ResourceTracker:      resourceTracker,
			UptimeCalculator:     uptime.NoOpCalculator,
			IPSigner:             NewIPSigner(signerIP, tlsKey, blsKey),
		},
		conn,
		cert,
		peerID,
		NewBlockingMessageQueue(
			metrics,
			logging.NoLog{},
			maxMessageToSend,
		),
	)
	return peer, peer.AwaitReady(ctx)
}

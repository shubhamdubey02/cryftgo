// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package p

import (
	"math/big"
	"time"

	"github.com/cryft-labs/coreth/plugin/evm"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/require"

	"github.com/shubhamdubey02/cryftgo/api/info"
	"github.com/shubhamdubey02/cryftgo/config"
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/tests/fixture/e2e"
	"github.com/shubhamdubey02/cryftgo/tests/fixture/tmpnet"
	"github.com/shubhamdubey02/cryftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgo/utils/crypto/secp256k1"
	"github.com/shubhamdubey02/cryftgo/utils/set"
	"github.com/shubhamdubey02/cryftgo/utils/units"
	"github.com/shubhamdubey02/cryftgo/vms/components/cryft"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/reward"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/txs"
	"github.com/shubhamdubey02/cryftgo/vms/secp256k1fx"
	"github.com/shubhamdubey02/cryftgo/wallet/subnet/primary/common"

	ginkgo "github.com/onsi/ginkgo/v2"
)

var _ = e2e.DescribePChain("[Interchain Workflow]", ginkgo.Label(e2e.UsesCChainLabel), func() {
	require := require.New(ginkgo.GinkgoT())

	const (
		transferAmount = 10 * units.Cryft
		weight         = 2_000 * units.Cryft // Used for both validation and delegation
	)

	ginkgo.It("should ensure that funds can be transferred from the P-Chain to the X-Chain and the C-Chain", func() {
		network := e2e.Env.GetNetwork()

		ginkgo.By("checking that the network has a compatible minimum stake duration", func() {
			minStakeDuration := cast.ToDuration(network.DefaultFlags[config.MinStakeDurationKey])
			require.Equal(tmpnet.DefaultMinStakeDuration, minStakeDuration)
		})

		ginkgo.By("creating wallet with a funded key to send from and recipient key to deliver to")
		recipientKey, err := secp256k1.NewPrivateKey()
		require.NoError(err)
		keychain := e2e.Env.NewKeychain(1)
		keychain.Add(recipientKey)
		nodeURI := e2e.Env.GetRandomNodeURI()
		baseWallet := e2e.NewWallet(keychain, nodeURI)
		xWallet := baseWallet.X()
		cWallet := baseWallet.C()
		pWallet := baseWallet.P()

		xBuilder := xWallet.Builder()
		xContext := xBuilder.Context()
		pBuilder := pWallet.Builder()
		pContext := pBuilder.Context()

		ginkgo.By("defining common configuration")
		recipientEthAddress := evm.GetEthAddress(recipientKey)
		cryftAssetID := xContext.CRYFTAssetID
		// Use the same owner for sending to X-Chain and importing funds to P-Chain
		recipientOwner := secp256k1fx.OutputOwners{
			Threshold: 1,
			Addrs: []ids.ShortID{
				recipientKey.Address(),
			},
		}
		// Use the same outputs for both X-Chain and C-Chain exports
		exportOutputs := []*cryft.TransferableOutput{
			{
				Asset: cryft.Asset{
					ID: cryftAssetID,
				},
				Out: &secp256k1fx.TransferOutput{
					Amt: transferAmount,
					OutputOwners: secp256k1fx.OutputOwners{
						Threshold: 1,
						Addrs: []ids.ShortID{
							keychain.Keys[0].Address(),
						},
					},
				},
			},
		}

		ginkgo.By("adding new node and waiting for it to report healthy")
		node := e2e.AddEphemeralNode(network, tmpnet.FlagsMap{})
		e2e.WaitForHealthy(node)

		ginkgo.By("retrieving new node's id and pop")
		infoClient := info.NewClient(node.URI)
		nodeID, nodePOP, err := infoClient.GetNodeID(e2e.DefaultContext())
		require.NoError(err)

		// Adding a validator should not break interchain transfer.
		endTime := time.Now().Add(30 * time.Second)
		ginkgo.By("adding the new node as a validator", func() {
			rewardKey, err := secp256k1.NewPrivateKey()
			require.NoError(err)

			const (
				delegationPercent = 0.10 // 10%
				delegationShare   = reward.PercentDenominator * delegationPercent
			)

			_, err = pWallet.IssueAddPermissionlessValidatorTx(
				&txs.SubnetValidator{
					Validator: txs.Validator{
						NodeID: nodeID,
						End:    uint64(endTime.Unix()),
						Wght:   weight,
					},
					Subnet: constants.PrimaryNetworkID,
				},
				nodePOP,
				pContext.CRYFTAssetID,
				&secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{rewardKey.Address()},
				},
				&secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{rewardKey.Address()},
				},
				delegationShare,
				e2e.WithDefaultContext(),
			)
			require.NoError(err)
		})

		// Adding a delegator should not break interchain transfer.
		ginkgo.By("adding a delegator to the new node", func() {
			rewardKey, err := secp256k1.NewPrivateKey()
			require.NoError(err)

			_, err = pWallet.IssueAddPermissionlessDelegatorTx(
				&txs.SubnetValidator{
					Validator: txs.Validator{
						NodeID: nodeID,
						End:    uint64(endTime.Unix()),
						Wght:   weight,
					},
					Subnet: constants.PrimaryNetworkID,
				},
				pContext.CRYFTAssetID,
				&secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs:     []ids.ShortID{rewardKey.Address()},
				},
				e2e.WithDefaultContext(),
			)
			require.NoError(err)
		})

		ginkgo.By("exporting CRYFT from the P-Chain to the X-Chain", func() {
			_, err := pWallet.IssueExportTx(
				xContext.BlockchainID,
				exportOutputs,
				e2e.WithDefaultContext(),
			)
			require.NoError(err)
		})

		ginkgo.By("importing CRYFT from the P-Chain to the X-Chain", func() {
			_, err := xWallet.IssueImportTx(
				constants.PlatformChainID,
				&recipientOwner,
				e2e.WithDefaultContext(),
			)
			require.NoError(err)
		})

		ginkgo.By("checking that the recipient address has received imported funds on the X-Chain", func() {
			balances, err := xWallet.Builder().GetFTBalance(common.WithCustomAddresses(set.Of(
				recipientKey.Address(),
			)))
			require.NoError(err)
			require.Positive(balances[cryftAssetID])
		})

		ginkgo.By("exporting CRYFT from the P-Chain to the C-Chain", func() {
			_, err := pWallet.IssueExportTx(
				cWallet.BlockchainID(),
				exportOutputs,
				e2e.WithDefaultContext(),
			)
			require.NoError(err)
		})

		ginkgo.By("initializing a new eth client")
		ethClient := e2e.NewEthClient(nodeURI)

		ginkgo.By("importing CRYFT from the P-Chain to the C-Chain", func() {
			_, err := cWallet.IssueImportTx(
				constants.PlatformChainID,
				recipientEthAddress,
				e2e.WithDefaultContext(),
				e2e.WithSuggestedGasPrice(ethClient),
			)
			require.NoError(err)
		})

		ginkgo.By("checking that the recipient address has received imported funds on the C-Chain")
		balance, err := ethClient.BalanceAt(e2e.DefaultContext(), recipientEthAddress, nil)
		require.NoError(err)
		require.Positive(balance.Cmp(big.NewInt(0)))

		ginkgo.By("stopping validator node to free up resources for a bootstrap check")
		require.NoError(node.Stop(e2e.DefaultContext()))

		e2e.CheckBootstrapIsPossible(network)
	})
})

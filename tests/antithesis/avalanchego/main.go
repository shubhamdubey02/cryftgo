// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package main

import (
	"context"
	"crypto/rand"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/shubhamdubey02/cryftgo/database"
	"github.com/shubhamdubey02/cryftgo/genesis"
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgo/snow/choices"
	"github.com/shubhamdubey02/cryftgo/tests/antithesis"
	"github.com/shubhamdubey02/cryftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgo/utils/crypto/secp256k1"
	"github.com/shubhamdubey02/cryftgo/utils/set"
	"github.com/shubhamdubey02/cryftgo/utils/units"
	"github.com/shubhamdubey02/cryftgo/vms/avm"
	"github.com/shubhamdubey02/cryftgo/vms/components/cryft"
	"github.com/shubhamdubey02/cryftgo/vms/components/verify"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm"
	"github.com/shubhamdubey02/cryftgo/vms/platformvm/status"
	"github.com/shubhamdubey02/cryftgo/vms/propertyfx"
	"github.com/shubhamdubey02/cryftgo/vms/secp256k1fx"
	"github.com/shubhamdubey02/cryftgo/wallet/subnet/primary"
	"github.com/shubhamdubey02/cryftgo/wallet/subnet/primary/common"

	xtxs "github.com/shubhamdubey02/cryftgo/vms/avm/txs"
	ptxs "github.com/shubhamdubey02/cryftgo/vms/platformvm/txs"
	xbuilder "github.com/shubhamdubey02/cryftgo/wallet/chain/x/builder"
)

const NumKeys = 5

func main() {
	c, err := antithesis.NewConfig(os.Args)
	if err != nil {
		log.Fatalf("invalid config: %s", err)
	}

	ctx := context.Background()
	if err := antithesis.AwaitHealthyNodes(ctx, c.URIs); err != nil {
		log.Fatalf("failed to await healthy nodes: %s", err)
	}

	kc := secp256k1fx.NewKeychain(genesis.EWOQKey)
	walletSyncStartTime := time.Now()
	wallet, err := primary.MakeWallet(ctx, &primary.WalletConfig{
		URI:           c.URIs[0],
		CRYFTKeychain: kc,
		EthKeychain:   kc,
	})
	if err != nil {
		log.Fatalf("failed to initialize wallet: %s", err)
	}
	log.Printf("synced wallet in %s", time.Since(walletSyncStartTime))

	genesisWorkload := &workload{
		id:     0,
		wallet: wallet,
		addrs:  set.Of(genesis.EWOQKey.Address()),
		uris:   c.URIs,
	}

	workloads := make([]*workload, NumKeys)
	workloads[0] = genesisWorkload

	var (
		genesisXWallet  = wallet.X()
		genesisXBuilder = genesisXWallet.Builder()
		genesisXContext = genesisXBuilder.Context()
		cryftAssetID    = genesisXContext.CRYFTAssetID
	)
	for i := 1; i < NumKeys; i++ {
		key, err := secp256k1.NewPrivateKey()
		if err != nil {
			log.Fatalf("failed to generate key: %s", err)
		}

		var (
			addr          = key.Address()
			baseStartTime = time.Now()
		)
		baseTx, err := genesisXWallet.IssueBaseTx([]*cryft.TransferableOutput{{
			Asset: cryft.Asset{
				ID: cryftAssetID,
			},
			Out: &secp256k1fx.TransferOutput{
				Amt: 100 * units.KiloCryft,
				OutputOwners: secp256k1fx.OutputOwners{
					Threshold: 1,
					Addrs: []ids.ShortID{
						addr,
					},
				},
			},
		}})
		if err != nil {
			log.Fatalf("failed to issue initial funding X-chain baseTx: %s", err)
		}
		log.Printf("issued initial funding X-chain baseTx %s in %s", baseTx.ID(), time.Since(baseStartTime))

		genesisWorkload.confirmXChainTx(ctx, baseTx)

		uri := c.URIs[i%len(c.URIs)]
		kc := secp256k1fx.NewKeychain(key)
		walletSyncStartTime := time.Now()
		wallet, err := primary.MakeWallet(ctx, &primary.WalletConfig{
			URI:           uri,
			CRYFTKeychain: kc,
			EthKeychain:   kc,
		})
		if err != nil {
			log.Fatalf("failed to initialize wallet: %s", err)
		}
		log.Printf("synced wallet in %s", time.Since(walletSyncStartTime))

		workloads[i] = &workload{
			id:     i,
			wallet: wallet,
			addrs:  set.Of(addr),
			uris:   c.URIs,
		}
	}

	for _, w := range workloads[1:] {
		go w.run(ctx)
	}
	genesisWorkload.run(ctx)
}

type workload struct {
	id     int
	wallet primary.Wallet
	addrs  set.Set[ids.ShortID]
	uris   []string
}

func (w *workload) run(ctx context.Context) {
	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}

	var (
		xWallet  = w.wallet.X()
		xBuilder = xWallet.Builder()
		pWallet  = w.wallet.P()
		pBuilder = pWallet.Builder()
	)
	xBalances, err := xBuilder.GetFTBalance()
	if err != nil {
		log.Fatalf("failed to fetch X-chain balances: %s", err)
	}
	pBalances, err := pBuilder.GetBalance()
	if err != nil {
		log.Fatalf("failed to fetch P-chain balances: %s", err)
	}
	var (
		xContext     = xBuilder.Context()
		cryftAssetID = xContext.CRYFTAssetID
		xCRYFT       = xBalances[cryftAssetID]
		pCRYFT       = pBalances[cryftAssetID]
	)
	log.Printf("wallet starting with %d X-chain nCRYFT and %d P-chain nCRYFT", xCRYFT, pCRYFT)

	for {
		val, err := rand.Int(rand.Reader, big.NewInt(5))
		if err != nil {
			log.Fatalf("failed to read randomness: %s", err)
		}

		flowID := val.Int64()
		log.Printf("wallet %d executing flow %d", w.id, flowID)
		switch flowID {
		case 0:
			w.issueXChainBaseTx(ctx)
		case 1:
			w.issueXChainCreateAssetTx(ctx)
		case 2:
			w.issueXChainOperationTx(ctx)
		case 3:
			w.issueXToPTransfer(ctx)
		case 4:
			w.issuePToXTransfer(ctx)
		}

		val, err = rand.Int(rand.Reader, big.NewInt(int64(time.Second)))
		if err != nil {
			log.Fatalf("failed to read randomness: %s", err)
		}

		timer.Reset(time.Duration(val.Int64()))
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}
	}
}

func (w *workload) issueXChainBaseTx(ctx context.Context) {
	var (
		xWallet  = w.wallet.X()
		xBuilder = xWallet.Builder()
	)
	balances, err := xBuilder.GetFTBalance()
	if err != nil {
		log.Printf("failed to fetch X-chain balances: %s", err)
		return
	}

	var (
		xContext      = xBuilder.Context()
		cryftAssetID  = xContext.CRYFTAssetID
		cryftBalance  = balances[cryftAssetID]
		baseTxFee     = xContext.BaseTxFee
		neededBalance = baseTxFee + units.Schmeckle
	)
	if cryftBalance < neededBalance {
		log.Printf("skipping X-chain tx issuance due to insufficient balance: %d < %d", cryftBalance, neededBalance)
		return
	}

	var (
		owner         = w.makeOwner()
		baseStartTime = time.Now()
	)
	baseTx, err := xWallet.IssueBaseTx(
		[]*cryft.TransferableOutput{
			{
				Asset: cryft.Asset{
					ID: cryftAssetID,
				},
				Out: &secp256k1fx.TransferOutput{
					Amt:          units.Schmeckle,
					OutputOwners: owner,
				},
			},
		},
	)
	if err != nil {
		log.Printf("failed to issue X-chain baseTx: %s", err)
		return
	}
	log.Printf("issued new X-chain baseTx %s in %s", baseTx.ID(), time.Since(baseStartTime))

	w.confirmXChainTx(ctx, baseTx)
	w.verifyXChainTxConsumedUTXOs(ctx, baseTx)
}

func (w *workload) issueXChainCreateAssetTx(ctx context.Context) {
	var (
		xWallet  = w.wallet.X()
		xBuilder = xWallet.Builder()
	)
	balances, err := xBuilder.GetFTBalance()
	if err != nil {
		log.Printf("failed to fetch X-chain balances: %s", err)
		return
	}

	var (
		xContext      = xBuilder.Context()
		cryftAssetID  = xContext.CRYFTAssetID
		cryftBalance  = balances[cryftAssetID]
		neededBalance = xContext.CreateAssetTxFee
	)
	if cryftBalance < neededBalance {
		log.Printf("skipping X-chain tx issuance due to insufficient balance: %d < %d", cryftBalance, neededBalance)
		return
	}

	var (
		owner                = w.makeOwner()
		createAssetStartTime = time.Now()
	)
	createAssetTx, err := xWallet.IssueCreateAssetTx(
		"HI",
		"HI",
		1,
		map[uint32][]verify.State{
			0: {
				&secp256k1fx.TransferOutput{
					Amt:          units.Schmeckle,
					OutputOwners: owner,
				},
			},
		},
	)
	if err != nil {
		log.Printf("failed to issue X-chain create asset transaction: %s", err)
		return
	}
	log.Printf("created new X-chain asset %s in %s", createAssetTx.ID(), time.Since(createAssetStartTime))

	w.confirmXChainTx(ctx, createAssetTx)
	w.verifyXChainTxConsumedUTXOs(ctx, createAssetTx)
}

func (w *workload) issueXChainOperationTx(ctx context.Context) {
	var (
		xWallet  = w.wallet.X()
		xBuilder = xWallet.Builder()
	)
	balances, err := xBuilder.GetFTBalance()
	if err != nil {
		log.Printf("failed to fetch X-chain balances: %s", err)
		return
	}

	var (
		xContext         = xBuilder.Context()
		cryftAssetID     = xContext.CRYFTAssetID
		cryftBalance     = balances[cryftAssetID]
		createAssetTxFee = xContext.CreateAssetTxFee
		baseTxFee        = xContext.BaseTxFee
		neededBalance    = createAssetTxFee + baseTxFee
	)
	if cryftBalance < neededBalance {
		log.Printf("skipping X-chain tx issuance due to insufficient balance: %d < %d", cryftBalance, neededBalance)
		return
	}

	var (
		owner                = w.makeOwner()
		createAssetStartTime = time.Now()
	)
	createAssetTx, err := xWallet.IssueCreateAssetTx(
		"HI",
		"HI",
		1,
		map[uint32][]verify.State{
			2: {
				&propertyfx.MintOutput{
					OutputOwners: owner,
				},
			},
		},
	)
	if err != nil {
		log.Printf("failed to issue X-chain create asset transaction: %s", err)
		return
	}
	log.Printf("created new X-chain asset %s in %s", createAssetTx.ID(), time.Since(createAssetStartTime))

	operationStartTime := time.Now()
	operationTx, err := xWallet.IssueOperationTxMintProperty(
		createAssetTx.ID(),
		&owner,
	)
	if err != nil {
		log.Printf("failed to issue X-chain operation transaction: %s", err)
		return
	}
	log.Printf("issued X-chain operation tx %s in %s", operationTx.ID(), time.Since(operationStartTime))

	w.confirmXChainTx(ctx, createAssetTx)
	w.verifyXChainTxConsumedUTXOs(ctx, createAssetTx)
	w.confirmXChainTx(ctx, operationTx)
	w.verifyXChainTxConsumedUTXOs(ctx, operationTx)
}

func (w *workload) issueXToPTransfer(ctx context.Context) {
	var (
		xWallet  = w.wallet.X()
		pWallet  = w.wallet.P()
		xBuilder = xWallet.Builder()
	)
	balances, err := xBuilder.GetFTBalance()
	if err != nil {
		log.Printf("failed to fetch X-chain balances: %s", err)
		return
	}

	var (
		xContext      = xBuilder.Context()
		cryftAssetID  = xContext.CRYFTAssetID
		cryftBalance  = balances[cryftAssetID]
		xBaseTxFee    = xContext.BaseTxFee
		pBuilder      = pWallet.Builder()
		pContext      = pBuilder.Context()
		pBaseTxFee    = pContext.BaseTxFee
		txFees        = xBaseTxFee + pBaseTxFee
		neededBalance = txFees + units.Cryft
	)
	if cryftBalance < neededBalance {
		log.Printf("skipping X-chain tx issuance due to insufficient balance: %d < %d", cryftBalance, neededBalance)
		return
	}

	var (
		owner           = w.makeOwner()
		exportStartTime = time.Now()
	)
	exportTx, err := xWallet.IssueExportTx(
		constants.PlatformChainID,
		[]*cryft.TransferableOutput{{
			Asset: cryft.Asset{
				ID: cryftAssetID,
			},
			Out: &secp256k1fx.TransferOutput{
				Amt: units.Cryft,
			},
		}},
	)
	if err != nil {
		log.Printf("failed to issue X-chain export transaction: %s", err)
		return
	}
	log.Printf("created X-chain export transaction %s in %s", exportTx.ID(), time.Since(exportStartTime))

	var (
		xChainID        = xContext.BlockchainID
		importStartTime = time.Now()
	)
	importTx, err := pWallet.IssueImportTx(
		xChainID,
		&owner,
	)
	if err != nil {
		log.Printf("failed to issue P-chain import transaction: %s", err)
		return
	}
	log.Printf("created P-chain import transaction %s in %s", importTx.ID(), time.Since(importStartTime))

	w.confirmXChainTx(ctx, exportTx)
	w.verifyXChainTxConsumedUTXOs(ctx, exportTx)
	w.confirmPChainTx(ctx, importTx)
	w.verifyPChainTxConsumedUTXOs(ctx, importTx)
}

func (w *workload) issuePToXTransfer(ctx context.Context) {
	var (
		xWallet  = w.wallet.X()
		pWallet  = w.wallet.P()
		xBuilder = xWallet.Builder()
		pBuilder = pWallet.Builder()
	)
	balances, err := pBuilder.GetBalance()
	if err != nil {
		log.Printf("failed to fetch P-chain balances: %s", err)
		return
	}

	var (
		xContext      = xBuilder.Context()
		pContext      = pBuilder.Context()
		cryftAssetID  = pContext.CRYFTAssetID
		cryftBalance  = balances[cryftAssetID]
		pBaseTxFee    = pContext.BaseTxFee
		xBaseTxFee    = xContext.BaseTxFee
		txFees        = pBaseTxFee + xBaseTxFee
		neededBalance = txFees + units.Schmeckle
	)
	if cryftBalance < neededBalance {
		log.Printf("skipping P-chain tx issuance due to insufficient balance: %d < %d", cryftBalance, neededBalance)
		return
	}

	var (
		xChainID        = xContext.BlockchainID
		owner           = w.makeOwner()
		exportStartTime = time.Now()
	)
	exportTx, err := pWallet.IssueExportTx(
		xChainID,
		[]*cryft.TransferableOutput{{
			Asset: cryft.Asset{
				ID: cryftAssetID,
			},
			Out: &secp256k1fx.TransferOutput{
				Amt: units.Schmeckle,
			},
		}},
	)
	if err != nil {
		log.Printf("failed to issue P-chain export transaction: %s", err)
		return
	}
	log.Printf("created P-chain export transaction %s in %s", exportTx.ID(), time.Since(exportStartTime))

	importStartTime := time.Now()
	importTx, err := xWallet.IssueImportTx(
		constants.PlatformChainID,
		&owner,
	)
	if err != nil {
		log.Printf("failed to issue X-chain import transaction: %s", err)
		return
	}
	log.Printf("created X-chain import transaction %s in %s", importTx.ID(), time.Since(importStartTime))

	w.confirmPChainTx(ctx, exportTx)
	w.verifyPChainTxConsumedUTXOs(ctx, exportTx)
	w.confirmXChainTx(ctx, importTx)
	w.verifyXChainTxConsumedUTXOs(ctx, importTx)
}

func (w *workload) makeOwner() secp256k1fx.OutputOwners {
	addr, _ := w.addrs.Peek()
	return secp256k1fx.OutputOwners{
		Threshold: 1,
		Addrs: []ids.ShortID{
			addr,
		},
	}
}

func (w *workload) confirmXChainTx(ctx context.Context, tx *xtxs.Tx) {
	txID := tx.ID()
	for _, uri := range w.uris {
		client := avm.NewClient(uri, "X")
		status, err := client.ConfirmTx(ctx, txID, 100*time.Millisecond)
		if err != nil {
			log.Printf("failed to confirm X-chain transaction %s on %s: %s", txID, uri, err)
			return
		}
		if status != choices.Accepted {
			log.Printf("failed to confirm X-chain transaction %s on %s: status == %s", txID, uri, status)
			return
		}
		log.Printf("confirmed X-chain transaction %s on %s", txID, uri)
	}
	log.Printf("confirmed X-chain transaction %s on all nodes", txID)
}

func (w *workload) confirmPChainTx(ctx context.Context, tx *ptxs.Tx) {
	txID := tx.ID()
	for _, uri := range w.uris {
		client := platformvm.NewClient(uri)
		s, err := client.AwaitTxDecided(ctx, txID, 100*time.Millisecond)
		if err != nil {
			log.Printf("failed to confirm P-chain transaction %s on %s: %s", txID, uri, err)
			return
		}
		if s.Status != status.Committed {
			log.Printf("failed to confirm P-chain transaction %s on %s: status == %s", txID, uri, s.Status)
			return
		}
		log.Printf("confirmed P-chain transaction %s on %s", txID, uri)
	}
	log.Printf("confirmed P-chain transaction %s on all nodes", txID)
}

func (w *workload) verifyXChainTxConsumedUTXOs(ctx context.Context, tx *xtxs.Tx) {
	txID := tx.ID()
	chainID := w.wallet.X().Builder().Context().BlockchainID
	for _, uri := range w.uris {
		client := avm.NewClient(uri, "X")

		utxos := common.NewUTXOs()
		err := primary.AddAllUTXOs(
			ctx,
			utxos,
			client,
			xbuilder.Parser.Codec(),
			chainID,
			chainID,
			w.addrs.List(),
		)
		if err != nil {
			log.Printf("failed to fetch X-chain UTXOs on %s: %s", uri, err)
			return
		}

		inputs := tx.Unsigned.InputIDs()
		for input := range inputs {
			_, err := utxos.GetUTXO(ctx, chainID, chainID, input)
			if err != database.ErrNotFound {
				log.Printf("failed to verify that X-chain UTXO %s was deleted on %s after %s", input, uri, txID)
				return
			}
		}
		log.Printf("confirmed all X-chain UTXOs consumed by %s are not present on %s", txID, uri)
	}
	log.Printf("confirmed all X-chain UTXOs consumed by %s are not present on all nodes", txID)
}

func (w *workload) verifyPChainTxConsumedUTXOs(ctx context.Context, tx *ptxs.Tx) {
	txID := tx.ID()
	for _, uri := range w.uris {
		client := platformvm.NewClient(uri)

		utxos := common.NewUTXOs()
		err := primary.AddAllUTXOs(
			ctx,
			utxos,
			client,
			ptxs.Codec,
			constants.PlatformChainID,
			constants.PlatformChainID,
			w.addrs.List(),
		)
		if err != nil {
			log.Printf("failed to fetch P-chain UTXOs on %s: %s", uri, err)
			return
		}

		inputs := tx.Unsigned.InputIDs()
		for input := range inputs {
			_, err := utxos.GetUTXO(ctx, constants.PlatformChainID, constants.PlatformChainID, input)
			if err != database.ErrNotFound {
				log.Printf("failed to verify that P-chain UTXO %s was deleted on %s after %s", input, uri, txID)
				return
			}
		}
		log.Printf("confirmed all P-chain UTXOs consumed by %s are not present on %s", txID, uri)
	}
	log.Printf("confirmed all P-chain UTXOs consumed by %s are not present on all nodes", txID)
}

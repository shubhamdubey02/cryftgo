// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package primary

import (
	"context"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/utils/constants"
	"github.com/shubhamdubey02/cryftgoftgo/utils/crypto/keychain"
	"github.com/shubhamdubey02/cryftgoftgo/utils/set"
	"github.com/shubhamdubey02/cryftgoftgo/vms/platformvm/txs"
	"github.com/shubhamdubey02/cryftgoftgo/wallet/chain/c"
	"github.com/shubhamdubey02/cryftgoftgo/wallet/chain/p"
	"github.com/shubhamdubey02/cryftgoftgo/wallet/chain/x"
	"github.com/shubhamdubey02/cryftgoftgo/wallet/subnet/primary/common"

	pbuilder "github.com/shubhamdubey02/cryftgoftgo/wallet/chain/p/builder"
	psigner "github.com/shubhamdubey02/cryftgoftgo/wallet/chain/p/signer"
	xbuilder "github.com/shubhamdubey02/cryftgoftgo/wallet/chain/x/builder"
	xsigner "github.com/shubhamdubey02/cryftgoftgo/wallet/chain/x/signer"
)

var _ Wallet = (*wallet)(nil)

// Wallet provides chain wallets for the primary network.
type Wallet interface {
	P() p.Wallet
	X() x.Wallet
	C() c.Wallet
}

type wallet struct {
	p p.Wallet
	x x.Wallet
	c c.Wallet
}

func (w *wallet) P() p.Wallet {
	return w.p
}

func (w *wallet) X() x.Wallet {
	return w.x
}

func (w *wallet) C() c.Wallet {
	return w.c
}

// Creates a new default wallet
func NewWallet(p p.Wallet, x x.Wallet, c c.Wallet) Wallet {
	return &wallet{
		p: p,
		x: x,
		c: c,
	}
}

// Creates a Wallet with the given set of options
func NewWalletWithOptions(w Wallet, options ...common.Option) Wallet {
	return NewWallet(
		p.NewWalletWithOptions(w.P(), options...),
		x.NewWalletWithOptions(w.X(), options...),
		c.NewWalletWithOptions(w.C(), options...),
	)
}

type WalletConfig struct {
	// Base URI to use for all node requests.
	URI string // required
	// Keys to use for signing all transactions.
	CRYFTKeychain keychain.Keychain // required
	EthKeychain   c.EthKeychain     // required
	// Set of P-chain transactions that the wallet should know about to be able
	// to generate transactions.
	PChainTxs map[ids.ID]*txs.Tx // optional
	// Set of P-chain transactions that the wallet should fetch to be able to
	// generate transactions.
	PChainTxsToFetch set.Set[ids.ID] // optional
}

// MakeWallet returns a wallet that supports issuing transactions to the chains
// living in the primary network.
//
// On creation, the wallet attaches to the provided uri and fetches all UTXOs
// that reference any of the provided keys. If the UTXOs are modified through an
// external issuance process, such as another instance of the wallet, the UTXOs
// may become out of sync. The wallet will also fetch all requested P-chain
// transactions.
//
// The wallet manages all state locally, and performs all tx signing locally.
func MakeWallet(ctx context.Context, config *WalletConfig) (Wallet, error) {
	cryftAddrs := config.CRYFTKeychain.Addresses()
	cryftState, err := FetchState(ctx, config.URI, cryftAddrs)
	if err != nil {
		return nil, err
	}

	ethAddrs := config.EthKeychain.EthAddresses()
	ethState, err := FetchEthState(ctx, config.URI, ethAddrs)
	if err != nil {
		return nil, err
	}

	pChainTxs := config.PChainTxs
	if pChainTxs == nil {
		pChainTxs = make(map[ids.ID]*txs.Tx)
	}

	for txID := range config.PChainTxsToFetch {
		txBytes, err := cryftState.PClient.GetTx(ctx, txID)
		if err != nil {
			return nil, err
		}
		tx, err := txs.Parse(txs.Codec, txBytes)
		if err != nil {
			return nil, err
		}
		pChainTxs[txID] = tx
	}

	pUTXOs := common.NewChainUTXOs(constants.PlatformChainID, cryftState.UTXOs)
	pBackend := p.NewBackend(cryftState.PCTX, pUTXOs, pChainTxs)
	pBuilder := pbuilder.New(cryftAddrs, cryftState.PCTX, pBackend)
	pSigner := psigner.New(config.CRYFTKeychain, pBackend)

	xChainID := cryftState.XCTX.BlockchainID
	xUTXOs := common.NewChainUTXOs(xChainID, cryftState.UTXOs)
	xBackend := x.NewBackend(cryftState.XCTX, xUTXOs)
	xBuilder := xbuilder.New(cryftAddrs, cryftState.XCTX, xBackend)
	xSigner := xsigner.New(config.CRYFTKeychain, xBackend)

	cChainID := cryftState.CCTX.BlockchainID()
	cUTXOs := common.NewChainUTXOs(cChainID, cryftState.UTXOs)
	cBackend := c.NewBackend(cryftState.CCTX, cUTXOs, ethState.Accounts)
	cBuilder := c.NewBuilder(cryftAddrs, ethAddrs, cBackend)
	cSigner := c.NewSigner(config.CRYFTKeychain, config.EthKeychain, cBackend)

	return NewWallet(
		p.NewWallet(pBuilder, pSigner, cryftState.PClient, pBackend),
		x.NewWallet(xBuilder, xSigner, cryftState.XClient, xBackend),
		c.NewWallet(cBuilder, cSigner, cryftState.CClient, ethState.Client, cBackend),
	), nil
}

// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package signer

import (
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/utils/crypto/keychain"
	"github.com/shubhamdubey02/cryftgoftgo/vms/components/cryft"
	"github.com/shubhamdubey02/cryftgoftgo/vms/platformvm/fx"
	"github.com/shubhamdubey02/cryftgoftgo/vms/platformvm/txs"

	stdcontext "context"
)

var _ Signer = (*txSigner)(nil)

type Signer interface {
	// Sign adds as many missing signatures as possible to the provided
	// transaction.
	//
	// If there are already some signatures on the transaction, those signatures
	// will not be removed.
	//
	// If the signer doesn't have the ability to provide a required signature,
	// the signature slot will be skipped without reporting an error.
	Sign(ctx stdcontext.Context, tx *txs.Tx) error
}

type Backend interface {
	GetUTXO(ctx stdcontext.Context, chainID, utxoID ids.ID) (*cryft.UTXO, error)
	GetSubnetOwner(ctx stdcontext.Context, subnetID ids.ID) (fx.Owner, error)
}

type txSigner struct {
	kc      keychain.Keychain
	backend Backend
}

func New(kc keychain.Keychain, backend Backend) Signer {
	return &txSigner{
		kc:      kc,
		backend: backend,
	}
}

func (s *txSigner) Sign(ctx stdcontext.Context, tx *txs.Tx) error {
	return tx.Unsigned.Visit(&visitor{
		kc:      s.kc,
		backend: s.backend,
		ctx:     ctx,
		tx:      tx,
	})
}

func SignUnsigned(
	ctx stdcontext.Context,
	signer Signer,
	utx txs.UnsignedTx,
) (*txs.Tx, error) {
	tx := &txs.Tx{Unsigned: utx}
	return tx, signer.Sign(ctx, tx)
}

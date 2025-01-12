// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package txs

import (
	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/snow"
	"github.com/shubhamdubey02/cryftgoftgo/vms/components/cryft"
	"github.com/shubhamdubey02/cryftgoftgo/vms/secp256k1fx"
)

var (
	_ UnsignedTx             = (*ExportTx)(nil)
	_ secp256k1fx.UnsignedTx = (*ExportTx)(nil)
)

// ExportTx is a transaction that exports an asset to another blockchain.
type ExportTx struct {
	BaseTx `serialize:"true"`

	// Which chain to send the funds to
	DestinationChain ids.ID `serialize:"true" json:"destinationChain"`

	// The outputs this transaction is sending to the other chain
	ExportedOuts []*cryft.TransferableOutput `serialize:"true" json:"exportedOutputs"`
}

func (t *ExportTx) InitCtx(ctx *snow.Context) {
	for _, out := range t.ExportedOuts {
		out.InitCtx(ctx)
	}
	t.BaseTx.InitCtx(ctx)
}

func (t *ExportTx) Visit(v Visitor) error {
	return v.ExportTx(t)
}

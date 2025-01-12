// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package c

import (
	"errors"
	"math/big"

	"github.com/cryft-labs/coreth/plugin/evm"

	"github.com/shubhamdubey02/cryftgo/ids"
	"github.com/shubhamdubey02/cryftgoftgo/utils"
	"github.com/shubhamdubey02/cryftgoftgo/utils/math"
	"github.com/shubhamdubey02/cryftgoftgo/utils/set"
	"github.com/shubhamdubey02/cryftgoftgo/vms/components/cryft"
	"github.com/shubhamdubey02/cryftgoftgo/vms/secp256k1fx"
	"github.com/shubhamdubey02/cryftgoftgo/wallet/subnet/primary/common"

	stdcontext "context"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

const cryftConversionRateInt = 1_000_000_000

var (
	_ Builder = (*builder)(nil)

	errInsufficientFunds = errors.New("insufficient funds")

	// cryftConversionRate is the conversion rate between the smallest
	// denomination on the X-Chain and P-chain, 1 nCRYFT, and the smallest
	// denomination on the C-Chain 1 wei. Where 1 nCRYFT = 1 gWei.
	//
	// This is only required for CRYFT because the denomination of 1 CRYFT is 9
	// decimal places on the X and P chains, but is 18 decimal places within the
	// EVM.
	cryftConversionRate = big.NewInt(cryftConversionRateInt)
)

// Builder provides a convenient interface for building unsigned C-chain
// transactions.
type Builder interface {
	// GetBalance calculates the amount of CRYFT that this builder has control
	// over.
	GetBalance(
		options ...common.Option,
	) (*big.Int, error)

	// GetImportableBalance calculates the amount of CRYFT that this builder
	// could import from the provided chain.
	//
	// - [chainID] specifies the chain the funds are from.
	GetImportableBalance(
		chainID ids.ID,
		options ...common.Option,
	) (uint64, error)

	// NewImportTx creates an import transaction that attempts to consume all
	// the available UTXOs and import the funds to [to].
	//
	// - [chainID] specifies the chain to be importing funds from.
	// - [to] specifies where to send the imported funds to.
	// - [baseFee] specifies the fee price willing to be paid by this tx.
	NewImportTx(
		chainID ids.ID,
		to ethcommon.Address,
		baseFee *big.Int,
		options ...common.Option,
	) (*evm.UnsignedImportTx, error)

	// NewExportTx creates an export transaction that attempts to send all the
	// provided [outputs] to the requested [chainID].
	//
	// - [chainID] specifies the chain to be exporting the funds to.
	// - [outputs] specifies the outputs to send to the [chainID].
	// - [baseFee] specifies the fee price willing to be paid by this tx.
	NewExportTx(
		chainID ids.ID,
		outputs []*secp256k1fx.TransferOutput,
		baseFee *big.Int,
		options ...common.Option,
	) (*evm.UnsignedExportTx, error)
}

// BuilderBackend specifies the required information needed to build unsigned
// C-chain transactions.
type BuilderBackend interface {
	Context

	UTXOs(ctx stdcontext.Context, sourceChainID ids.ID) ([]*cryft.UTXO, error)
	Balance(ctx stdcontext.Context, addr ethcommon.Address) (*big.Int, error)
	Nonce(ctx stdcontext.Context, addr ethcommon.Address) (uint64, error)
}

type builder struct {
	cryftAddrs set.Set[ids.ShortID]
	ethAddrs   set.Set[ethcommon.Address]
	backend    BuilderBackend
}

// NewBuilder returns a new transaction builder.
//
//   - [cryftAddrs] is the set of addresses in the CRYFT format that the builder
//     assumes can be used when signing the transactions in the future.
//   - [ethAddrs] is the set of addresses in the Eth format that the builder
//     assumes can be used when signing the transactions in the future.
//   - [backend] provides the required access to the chain's context and state
//     to build out the transactions.
func NewBuilder(
	cryftAddrs set.Set[ids.ShortID],
	ethAddrs set.Set[ethcommon.Address],
	backend BuilderBackend,
) Builder {
	return &builder{
		cryftAddrs: cryftAddrs,
		ethAddrs:   ethAddrs,
		backend:    backend,
	}
}

func (b *builder) GetBalance(
	options ...common.Option,
) (*big.Int, error) {
	var (
		ops          = common.NewOptions(options)
		ctx          = ops.Context()
		addrs        = ops.EthAddresses(b.ethAddrs)
		totalBalance = new(big.Int)
	)
	for addr := range addrs {
		balance, err := b.backend.Balance(ctx, addr)
		if err != nil {
			return nil, err
		}
		totalBalance.Add(totalBalance, balance)
	}

	return totalBalance, nil
}

func (b *builder) GetImportableBalance(
	chainID ids.ID,
	options ...common.Option,
) (uint64, error) {
	ops := common.NewOptions(options)
	utxos, err := b.backend.UTXOs(ops.Context(), chainID)
	if err != nil {
		return 0, err
	}

	var (
		addrs           = ops.Addresses(b.cryftAddrs)
		minIssuanceTime = ops.MinIssuanceTime()
		cryftAssetID    = b.backend.CRYFTAssetID()
		balance         uint64
	)
	for _, utxo := range utxos {
		amount, _, ok := getSpendableAmount(utxo, addrs, minIssuanceTime, cryftAssetID)
		if !ok {
			continue
		}

		newBalance, err := math.Add64(balance, amount)
		if err != nil {
			return 0, err
		}
		balance = newBalance
	}

	return balance, nil
}

func (b *builder) NewImportTx(
	chainID ids.ID,
	to ethcommon.Address,
	baseFee *big.Int,
	options ...common.Option,
) (*evm.UnsignedImportTx, error) {
	ops := common.NewOptions(options)
	utxos, err := b.backend.UTXOs(ops.Context(), chainID)
	if err != nil {
		return nil, err
	}

	var (
		addrs           = ops.Addresses(b.cryftAddrs)
		minIssuanceTime = ops.MinIssuanceTime()
		cryftAssetID    = b.backend.CRYFTAssetID()

		importedInputs = make([]*cryft.TransferableInput, 0, len(utxos))
		importedAmount uint64
	)
	for _, utxo := range utxos {
		amount, inputSigIndices, ok := getSpendableAmount(utxo, addrs, minIssuanceTime, cryftAssetID)
		if !ok {
			continue
		}

		importedInputs = append(importedInputs, &cryft.TransferableInput{
			UTXOID: utxo.UTXOID,
			Asset:  utxo.Asset,
			FxID:   secp256k1fx.ID,
			In: &secp256k1fx.TransferInput{
				Amt: amount,
				Input: secp256k1fx.Input{
					SigIndices: inputSigIndices,
				},
			},
		})

		newImportedAmount, err := math.Add64(importedAmount, amount)
		if err != nil {
			return nil, err
		}
		importedAmount = newImportedAmount
	}

	utils.Sort(importedInputs)
	tx := &evm.UnsignedImportTx{
		NetworkID:      b.backend.NetworkID(),
		BlockchainID:   b.backend.BlockchainID(),
		SourceChain:    chainID,
		ImportedInputs: importedInputs,
	}

	// We must initialize the bytes of the tx to calculate the initial cost
	wrappedTx := &evm.Tx{UnsignedAtomicTx: tx}
	if err := wrappedTx.Sign(evm.Codec, nil); err != nil {
		return nil, err
	}

	gasUsedWithoutOutput, err := tx.GasUsed(true /*=IsApricotPhase5*/)
	if err != nil {
		return nil, err
	}
	gasUsedWithOutput := gasUsedWithoutOutput + evm.EVMOutputGas

	txFee, err := evm.CalculateDynamicFee(gasUsedWithOutput, baseFee)
	if err != nil {
		return nil, err
	}

	if importedAmount <= txFee {
		return nil, errInsufficientFunds
	}

	tx.Outs = []evm.EVMOutput{{
		Address: to,
		Amount:  importedAmount - txFee,
		AssetID: cryftAssetID,
	}}
	return tx, nil
}

func (b *builder) NewExportTx(
	chainID ids.ID,
	outputs []*secp256k1fx.TransferOutput,
	baseFee *big.Int,
	options ...common.Option,
) (*evm.UnsignedExportTx, error) {
	var (
		cryftAssetID    = b.backend.CRYFTAssetID()
		exportedOutputs = make([]*cryft.TransferableOutput, len(outputs))
		exportedAmount  uint64
	)
	for i, output := range outputs {
		exportedOutputs[i] = &cryft.TransferableOutput{
			Asset: cryft.Asset{ID: cryftAssetID},
			FxID:  secp256k1fx.ID,
			Out:   output,
		}

		newExportedAmount, err := math.Add64(exportedAmount, output.Amt)
		if err != nil {
			return nil, err
		}
		exportedAmount = newExportedAmount
	}

	cryft.SortTransferableOutputs(exportedOutputs, evm.Codec)
	tx := &evm.UnsignedExportTx{
		NetworkID:        b.backend.NetworkID(),
		BlockchainID:     b.backend.BlockchainID(),
		DestinationChain: chainID,
		ExportedOutputs:  exportedOutputs,
	}

	// We must initialize the bytes of the tx to calculate the initial cost
	wrappedTx := &evm.Tx{UnsignedAtomicTx: tx}
	if err := wrappedTx.Sign(evm.Codec, nil); err != nil {
		return nil, err
	}

	cost, err := tx.GasUsed(true /*=IsApricotPhase5*/)
	if err != nil {
		return nil, err
	}

	initialFee, err := evm.CalculateDynamicFee(cost, baseFee)
	if err != nil {
		return nil, err
	}

	amountToConsume, err := math.Add64(exportedAmount, initialFee)
	if err != nil {
		return nil, err
	}

	var (
		ops    = common.NewOptions(options)
		ctx    = ops.Context()
		addrs  = ops.EthAddresses(b.ethAddrs)
		inputs = make([]evm.EVMInput, 0, addrs.Len())
	)
	for addr := range addrs {
		if amountToConsume == 0 {
			break
		}

		prevFee, err := evm.CalculateDynamicFee(cost, baseFee)
		if err != nil {
			return nil, err
		}

		newCost := cost + evm.EVMInputGas
		newFee, err := evm.CalculateDynamicFee(newCost, baseFee)
		if err != nil {
			return nil, err
		}

		additionalFee := newFee - prevFee

		balance, err := b.backend.Balance(ctx, addr)
		if err != nil {
			return nil, err
		}

		// Since the asset is CRYFT, we divide by the cryftConversionRate to
		// convert back to the correct denomination of CRYFT that can be
		// exported.
		cryftBalance := new(big.Int).Div(balance, cryftConversionRate).Uint64()

		// If the balance for [addr] is insufficient to cover the additional
		// cost of adding an input to the transaction, skip adding the input
		// altogether.
		if cryftBalance <= additionalFee {
			continue
		}

		// Update the cost for the next iteration
		cost = newCost

		amountToConsume, err = math.Add64(amountToConsume, additionalFee)
		if err != nil {
			return nil, err
		}

		nonce, err := b.backend.Nonce(ctx, addr)
		if err != nil {
			return nil, err
		}

		inputAmount := min(amountToConsume, cryftBalance)
		inputs = append(inputs, evm.EVMInput{
			Address: addr,
			Amount:  inputAmount,
			AssetID: cryftAssetID,
			Nonce:   nonce,
		})
		amountToConsume -= inputAmount
	}

	if amountToConsume > 0 {
		return nil, errInsufficientFunds
	}

	utils.Sort(inputs)
	tx.Ins = inputs

	snowCtx, err := newSnowContext(b.backend)
	if err != nil {
		return nil, err
	}
	for _, out := range tx.ExportedOutputs {
		out.InitCtx(snowCtx)
	}
	return tx, nil
}

func getSpendableAmount(
	utxo *cryft.UTXO,
	addrs set.Set[ids.ShortID],
	minIssuanceTime uint64,
	cryftAssetID ids.ID,
) (uint64, []uint32, bool) {
	if utxo.Asset.ID != cryftAssetID {
		// Only CRYFT can be imported
		return 0, nil, false
	}

	out, ok := utxo.Out.(*secp256k1fx.TransferOutput)
	if !ok {
		// Can't import an unknown transfer output type
		return 0, nil, false
	}

	inputSigIndices, ok := common.MatchOwners(&out.OutputOwners, addrs, minIssuanceTime)
	return out.Amt, inputSigIndices, ok
}

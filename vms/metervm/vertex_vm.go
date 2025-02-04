// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package metervm

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/shubhamdubey02/cryftgo/api/metrics"
	"github.com/shubhamdubey02/cryftgo/database"
	"github.com/shubhamdubey02/cryftgo/snow"
	"github.com/shubhamdubey02/cryftgo/snow/consensus/snowstorm"
	"github.com/shubhamdubey02/cryftgo/snow/engine/avalanche/vertex"
	"github.com/shubhamdubey02/cryftgo/snow/engine/common"
	"github.com/shubhamdubey02/cryftgo/utils/timer/mockable"
)

var (
	_ vertex.LinearizableVMWithEngine = (*vertexVM)(nil)
	_ snowstorm.Tx                    = (*meterTx)(nil)
)

func NewVertexVM(vm vertex.LinearizableVMWithEngine) vertex.LinearizableVMWithEngine {
	return &vertexVM{
		LinearizableVMWithEngine: vm,
	}
}

type vertexVM struct {
	vertex.LinearizableVMWithEngine
	vertexMetrics
	clock mockable.Clock
}

func (vm *vertexVM) Initialize(
	ctx context.Context,
	chainCtx *snow.Context,
	db database.Database,
	genesisBytes,
	upgradeBytes,
	configBytes []byte,
	toEngine chan<- common.Message,
	fxs []*common.Fx,
	appSender common.AppSender,
) error {
	registerer := prometheus.NewRegistry()
	if err := vm.vertexMetrics.Initialize("", registerer); err != nil {
		return err
	}

	optionalGatherer := metrics.NewOptionalGatherer()
	multiGatherer := metrics.NewMultiGatherer()
	if err := multiGatherer.Register("metervm", registerer); err != nil {
		return err
	}
	if err := multiGatherer.Register("", optionalGatherer); err != nil {
		return err
	}
	if err := chainCtx.Metrics.Register(multiGatherer); err != nil {
		return err
	}
	chainCtx.Metrics = optionalGatherer

	return vm.LinearizableVMWithEngine.Initialize(
		ctx,
		chainCtx,
		db,
		genesisBytes,
		upgradeBytes,
		configBytes,
		toEngine,
		fxs,
		appSender,
	)
}

func (vm *vertexVM) ParseTx(ctx context.Context, b []byte) (snowstorm.Tx, error) {
	start := vm.clock.Time()
	tx, err := vm.LinearizableVMWithEngine.ParseTx(ctx, b)
	end := vm.clock.Time()
	duration := float64(end.Sub(start))
	if err != nil {
		vm.vertexMetrics.parseErr.Observe(duration)
		return nil, err
	}
	vm.vertexMetrics.parse.Observe(duration)
	return &meterTx{
		Tx: tx,
		vm: vm,
	}, nil
}

type meterTx struct {
	snowstorm.Tx

	vm *vertexVM
}

func (mtx *meterTx) Verify(ctx context.Context) error {
	start := mtx.vm.clock.Time()
	err := mtx.Tx.Verify(ctx)
	end := mtx.vm.clock.Time()
	duration := float64(end.Sub(start))
	if err != nil {
		mtx.vm.vertexMetrics.verifyErr.Observe(duration)
	} else {
		mtx.vm.vertexMetrics.verify.Observe(duration)
	}
	return err
}

func (mtx *meterTx) Accept(ctx context.Context) error {
	start := mtx.vm.clock.Time()
	err := mtx.Tx.Accept(ctx)
	end := mtx.vm.clock.Time()
	mtx.vm.vertexMetrics.accept.Observe(float64(end.Sub(start)))
	return err
}

func (mtx *meterTx) Reject(ctx context.Context) error {
	start := mtx.vm.clock.Time()
	err := mtx.Tx.Reject(ctx)
	end := mtx.vm.clock.Time()
	mtx.vm.vertexMetrics.reject.Observe(float64(end.Sub(start)))
	return err
}

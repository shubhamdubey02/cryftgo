// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package metervm

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/shubhamdubey02/cryftgo/utils/metric"
	"github.com/shubhamdubey02/cryftgo/utils/wrappers"
)

type vertexMetrics struct {
	parse,
	parseErr,
	verify,
	verifyErr,
	accept,
	reject metric.Averager
}

func (m *vertexMetrics) Initialize(
	namespace string,
	reg prometheus.Registerer,
) error {
	errs := wrappers.Errs{}
	m.parse = newAverager(namespace, "parse_tx", reg, &errs)
	m.parseErr = newAverager(namespace, "parse_tx_err", reg, &errs)
	m.verify = newAverager(namespace, "verify_tx", reg, &errs)
	m.verifyErr = newAverager(namespace, "verify_tx_err", reg, &errs)
	m.accept = newAverager(namespace, "accept", reg, &errs)
	m.reject = newAverager(namespace, "reject", reg, &errs)
	return errs.Err
}

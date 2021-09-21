// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package rsm

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-rsm/pkg/config"
	"github.com/onosproject/onos-rsm/pkg/store"
	"time"
)

var logManager = logging.GetLogger("rsm", "manager")

type Manager struct {
	UeStore      store.Store
	CellStore    store.Store
	MetricStore  store.Store
}

func NewManager(cfg *config.AppConfig, ueStore store.Store, cellStore store.Store, metricStore store.Store) Manager {
	return Manager{
		UeStore:      ueStore,
		CellStore:    cellStore,
		MetricStore:  metricStore,
	}
}

func (m *Manager) Run(ctx context.Context) {
	// ToDo: add
	logManager.Info("running rsm manager")
	for {
		// ToDo: add
		time.Sleep(1000 * time.Second)
	}
}
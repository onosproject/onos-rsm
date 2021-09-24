// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package rsm

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-rsm/pkg/config"
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	"github.com/onosproject/onos-rsm/pkg/store"
	"time"
)

var logManager = logging.GetLogger("rsm", "manager")

type Manager struct {
	ueStore store.Store
	sliceStore store.Store
	sliceAssocStore store.Store
	rnibClient rnib.Client
}

func NewManager(cfg *config.AppConfig, rnibClient rnib.Client, ueStore store.Store, sliceStore store.Store, sliceAssocStore store.Store) Manager {
	return Manager{
		ueStore: ueStore,
		rnibClient: rnibClient,
		sliceStore: sliceStore,
		sliceAssocStore: sliceAssocStore,
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
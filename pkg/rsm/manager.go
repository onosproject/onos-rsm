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
	"github.com/onosproject/onos-rsm/pkg/uenib"
	"time"
)

var logManager = logging.GetLogger("rsm", "manager")

// Manager is the RSM manager struct
type Manager struct {
	ueStore         store.Store
	sliceStore      store.Store
	sliceAssocStore store.Store
	rnibClient      rnib.TopoClient
	uenibClient     uenib.UenibClient
}

// NewManager creates the RSM manager
func NewManager(cfg *config.AppConfig, rnibClient rnib.TopoClient, uenibClient uenib.UenibClient, ueStore store.Store, sliceStore store.Store, sliceAssocStore store.Store) Manager {
	return Manager{
		ueStore:         ueStore,
		rnibClient:      rnibClient,
		uenibClient:     uenibClient,
		sliceStore:      sliceStore,
		sliceAssocStore: sliceAssocStore,
	}
}

// Run runs the RSM manager
func (m *Manager) Run(ctx context.Context) {
	// ToDo: add
	logManager.Info("running rsm manager")
	for {
		// ToDo: add
		time.Sleep(1000 * time.Second)
	}
}

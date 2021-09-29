// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package manager

import (
	"context"
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-lib-go/pkg/northbound"
	app "github.com/onosproject/onos-ric-sdk-go/pkg/config/app/default"
	"github.com/onosproject/onos-rsm/pkg/broker"
	appConfig "github.com/onosproject/onos-rsm/pkg/config"
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	nbi "github.com/onosproject/onos-rsm/pkg/northbound"
	"github.com/onosproject/onos-rsm/pkg/rsm"
	"github.com/onosproject/onos-rsm/pkg/southbound/e2"
	"github.com/onosproject/onos-rsm/pkg/store"
	"github.com/onosproject/onos-rsm/pkg/uenib"
)

var log = logging.GetLogger("manager")

// Config is a manager configuration
type Config struct {
	CAPath      string
	KeyPath     string
	CertPath    string
	ConfigPath  string
	E2tEndpoint string
	GRPCPort    int
	AppConfig   *app.Config
	SMName      string
	SMVersion   string
}

// NewManager creates a new manager
func NewManager(config Config) *Manager {
	appCfg, err := appConfig.NewConfig(config.ConfigPath)
	if err != nil {
		log.Warn(err)
	}
	subscriptionBroker := broker.NewBroker()

	ueStore := store.NewStore()
	sliceStore := store.NewStore()
	sliceAssocStore := store.NewStore()
	rnibClient, err := rnib.NewClient()
	if err != nil {
		log.Warn(err)
	}
	uenibClient, err := uenib.NewClient(context.Background(), config.CertPath, config.KeyPath)
	if err != nil {
		log.Warn(err)
	}
	ctrlReqChs := make(map[string]chan *e2api.ControlMessage)

	e2Manager, err := e2.NewManager(
		e2.WithE2TAddress("onos-e2t", 5150),
		e2.WithServiceModel(e2.ServiceModelName(config.SMName), e2.ServiceModelVersion(config.SMVersion)),
		e2.WithAppConfig(appCfg),
		e2.WithAppID("onos-rsm"),
		e2.WithBroker(subscriptionBroker),
		e2.WithRnibClient(rnibClient),
		e2.WithUenibClient(uenibClient),
		e2.WithUEStore(ueStore),
		e2.WithSliceStore(sliceStore),
		e2.WithSliceAssocStore(sliceAssocStore),
		e2.WithCtrlReqChs(ctrlReqChs),
	)

	if err != nil {
		log.Warn(err)
	}

	manager := &Manager{
		appConfig:       appCfg,
		config:          config,
		e2Manager:       e2Manager,
		slicingManager:  rsm.NewManager(appCfg, rnibClient, uenibClient, ueStore, sliceStore, sliceAssocStore),
		ueStore:         ueStore,
		rnibClient:      rnibClient,
		uenibClient:     uenibClient,
		sliceStore:      sliceStore,
		sliceAssocStore: sliceAssocStore,
		ctrlReqChs:      ctrlReqChs,
	}
	return manager
}

// Manager is a manager for the MHO xAPP service
type Manager struct {
	appConfig       appConfig.Config
	config          Config
	e2Manager       e2.Manager
	rnibClient      rnib.TopoClient
	uenibClient		uenib.UenibClient
	slicingManager  rsm.Manager
	ueStore         store.Store
	sliceStore      store.Store
	sliceAssocStore store.Store
	ctrlReqChs      map[string]chan *e2api.ControlMessage
}

// Run starts the manager and the associated services
func (m *Manager) Run() {
	log.Info("Running Manager")
	if err := m.start(); err != nil {
		log.Fatal("Unable to run Manager", err)
	}
}

func (m *Manager) start() error {
	// Start Northbound server
	err := m.startNorthboundServer()
	if err != nil {
		return err
	}

	err = m.e2Manager.Start()
	if err != nil {
		log.Warn(err)
		return err
	}

	go m.slicingManager.Run(context.Background())

	return nil
}

// Close kills the channels and manager related objects
func (m *Manager) Close() {
	log.Info("Closing Manager")
}

func (m *Manager) startNorthboundServer() error {
	s := northbound.NewServer(northbound.NewServerCfg(
		m.config.CAPath,
		m.config.KeyPath,
		m.config.CertPath,
		int16(m.config.GRPCPort),
		true,
		northbound.SecurityConfig{}))

	//TODO - RSM northbound service
	s.AddService(nbi.NewService(m.ctrlReqChs, m.rnibClient, m.uenibClient, m.ueStore, m.sliceStore, m.sliceAssocStore, e2.NewControlMessageHandler()))

	doneCh := make(chan error)
	go func() {
		err := s.Serve(func(started string) {
			log.Info("Started NBI on ", started)
			close(doneCh)
		})
		if err != nil {
			doneCh <- err
		}
	}()
	return <-doneCh
}

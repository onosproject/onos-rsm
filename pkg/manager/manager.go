// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package manager

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-lib-go/pkg/northbound"
	app "github.com/onosproject/onos-ric-sdk-go/pkg/config/app/default"
	"github.com/onosproject/onos-rsm/pkg/broker"
	appConfig "github.com/onosproject/onos-rsm/pkg/config"
	"github.com/onosproject/onos-rsm/pkg/rsm"
	"github.com/onosproject/onos-rsm/pkg/southbound/e2"
	"github.com/onosproject/onos-rsm/pkg/store"
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
	cellStore := store.NewStore()
	metricStore := store.NewStore()

	e2Manager, err := e2.NewManager(
		e2.WithE2TAddress("onos-e2t", 5150),
		e2.WithServiceModel(e2.ServiceModelName(config.SMName), e2.ServiceModelVersion(config.SMVersion)),
		e2.WithAppConfig(appCfg),
		e2.WithAppID("onos-rsm"),
		e2.WithBroker(subscriptionBroker),
		e2.WithUEStore(ueStore),
		e2.WithCellStore(cellStore),
		e2.WithMetricStore(metricStore),
	)

	if err != nil {
		log.Warn(err)
	}

	manager := &Manager{
		appConfig: appCfg,
		config:    config,
		e2Manager: e2Manager,
		slicingManager: rsm.NewManager(appCfg, ueStore, cellStore, metricStore),
		ueStore:   ueStore,
		cellStore: cellStore,
		metricStore: metricStore,
	}
	return manager
}


// Manager is a manager for the MHO xAPP service
type Manager struct {
	appConfig appConfig.Config
	config    Config
	e2Manager e2.Manager
	slicingManager rsm.Manager
	ueStore   store.Store
	cellStore store.Store
	metricStore store.Store
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
	//s.AddService(nbi.NewService(m.ueStore, m.cellStore))

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
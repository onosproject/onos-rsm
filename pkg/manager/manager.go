// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-lib-go/pkg/northbound"
	app "github.com/onosproject/onos-ric-sdk-go/pkg/config/app/default"
	"github.com/onosproject/onos-rsm/pkg/broker"
	appConfig "github.com/onosproject/onos-rsm/pkg/config"
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	"github.com/onosproject/onos-rsm/pkg/nib/uenib"
	nbi "github.com/onosproject/onos-rsm/pkg/northbound"
	"github.com/onosproject/onos-rsm/pkg/slicing"
	"github.com/onosproject/onos-rsm/pkg/southbound/e2"
	"strconv"
	"strings"
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
	UenibHost   string
	AppID       string
	AckTimer    int
}

func NewManager(config Config) *Manager {
	appCfg, err := appConfig.NewConfig(config.ConfigPath)
	if err != nil {
		log.Warn(err)
	}
	subscriptionBroker := broker.NewBroker()
	rnibClient, err := rnib.NewClient()
	if err != nil {
		log.Warn(err)
	}
	uenibClient, err := uenib.NewClient(context.Background(), config.CertPath, config.KeyPath, config.UenibHost)
	if err != nil {
		log.Warn(err)
	}
	ctrlReqChsSliceCreate := make(map[string]chan *e2.CtrlMsg)
	ctrlReqChsSliceUpdate := make(map[string]chan *e2.CtrlMsg)
	ctrlReqChsSliceDelete := make(map[string]chan *e2.CtrlMsg)
	ctrlReqChsUeAssociate := make(map[string]chan *e2.CtrlMsg)

	rsmReqCh := make(chan *nbi.RsmMsg)

	slicingManager := slicing.NewManager(
		slicing.WithRnibClient(rnibClient),
		slicing.WithUenibClient(uenibClient),
		slicing.WithCtrlReqChs(ctrlReqChsSliceCreate, ctrlReqChsSliceUpdate, ctrlReqChsSliceDelete, ctrlReqChsUeAssociate),
		slicing.WithNbiReqChs(rsmReqCh),
		slicing.WithAckTimer(config.AckTimer),
	)

	e2tHostAddr := strings.Split(config.E2tEndpoint, ":")[0]
	e2tPort, err := strconv.Atoi(strings.Split(config.E2tEndpoint, ":")[1])
	if err != nil {
		log.Warn(err)
	}

	e2Manager, err := e2.NewManager(
		e2.WithE2TAddress(e2tHostAddr, e2tPort),
		e2.WithServiceModel(e2.ServiceModelName(config.SMName), e2.ServiceModelVersion(config.SMVersion)),
		e2.WithAppConfig(appCfg),
		e2.WithAppID(config.AppID),
		e2.WithBroker(subscriptionBroker),
		e2.WithRnibClient(rnibClient),
		e2.WithUenibClient(uenibClient),
		e2.WithCtrlReqChs(ctrlReqChsSliceCreate, ctrlReqChsSliceUpdate, ctrlReqChsSliceDelete, ctrlReqChsUeAssociate),
	)
	if err != nil {
		log.Warn(err)
	}

	return &Manager{
		appConfig:             appCfg,
		config:                config,
		e2Manager:             e2Manager,
		rnibClient:            rnibClient,
		uenibClient:           uenibClient,
		slicingManager:        slicingManager,
		ctrlReqChsSliceCreate: ctrlReqChsSliceCreate,
		ctrlReqChsSliceUpdate: ctrlReqChsSliceUpdate,
		ctrlReqChsSliceDelete: ctrlReqChsSliceDelete,
		ctrlReqChsUeAssociate: ctrlReqChsUeAssociate,
		rsmReqCh:              rsmReqCh,
	}
}

// Manager is a manager for the RSM xAPP service
type Manager struct {
	appConfig             appConfig.Config
	config                Config
	e2Manager             e2.Manager
	rnibClient            rnib.TopoClient
	uenibClient           uenib.Client
	slicingManager        slicing.Manager
	ctrlReqChsSliceCreate map[string]chan *e2.CtrlMsg
	ctrlReqChsSliceUpdate map[string]chan *e2.CtrlMsg
	ctrlReqChsSliceDelete map[string]chan *e2.CtrlMsg
	ctrlReqChsUeAssociate map[string]chan *e2.CtrlMsg
	rsmReqCh              chan *nbi.RsmMsg
}

// Run starts the manager and the associated services
func (m *Manager) Run() {
	log.Info("Running Manager")
	if err := m.start(); err != nil {
		log.Fatal("Unable to run Manager", err)
	}
}

func (m *Manager) start() error {
	err := m.startNorthboundServer()
	if err != nil {
		log.Warn(err)
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

	s.AddService(nbi.NewService(m.rnibClient, m.uenibClient, m.rsmReqCh))

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

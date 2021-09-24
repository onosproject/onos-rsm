// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package e2

import (
	"context"
	"fmt"
	prototypes "github.com/gogo/protobuf/types"
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/pdubuilder"
	e2sm_rsm "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-rsm-ies"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	e2client "github.com/onosproject/onos-ric-sdk-go/pkg/e2/v1beta1"
	"github.com/onosproject/onos-rsm/pkg/broker"
	appConfig "github.com/onosproject/onos-rsm/pkg/config"
	"github.com/onosproject/onos-rsm/pkg/monitoring"
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	"github.com/onosproject/onos-rsm/pkg/store"
	"google.golang.org/protobuf/proto"
	"strings"
)

var log = logging.GetLogger("e2", "subscription", "manager")

const (
	oid = "1.3.6.1.4.1.53148.1.1.2.102"
)

// Node e2 manager interface
type Node interface {
	Start() error
	Stop() error
}

// Manager is a E2 session manager
type Manager struct {
	e2client        e2client.Client
	rnibClient      rnib.Client
	serviceModel    ServiceModelOptions
	appConfig       *appConfig.AppConfig
	streams         broker.Broker
	ueStore         store.Store
	sliceStore      store.Store
	sliceAssocStore store.Store
	CtrlReqChs      map[string]chan *e2api.ControlMessage
}

// NewManager creates a new subscription manager
func NewManager(opts ...Option) (Manager, error) {
	log.Info("Init E2Manager")
	options := Options{}

	for _, opt := range opts {
		opt.apply(&options)
	}

	serviceModelName := e2client.ServiceModelName(options.ServiceModel.Name)
	serviceModelVersion := e2client.ServiceModelVersion(options.ServiceModel.Version)
	appID := e2client.AppID(options.App.AppID)
	e2Client := e2client.NewClient(
		e2client.WithServiceModel(serviceModelName, serviceModelVersion),
		e2client.WithAppID(appID),
		e2client.WithE2TAddress(options.E2TService.Host, options.E2TService.Port))

	return Manager{
		e2client:   e2Client,
		rnibClient: options.App.RnibClient,
		serviceModel: ServiceModelOptions{
			Name:    options.ServiceModel.Name,
			Version: options.ServiceModel.Version,
		},
		appConfig:       options.App.AppConfig,
		streams:         options.App.Broker,
		ueStore:         options.App.UEStore,
		sliceStore:      options.App.SliceStore,
		sliceAssocStore: options.App.SliceAssocStore,
		CtrlReqChs:      options.App.CtrlReqChs,
	}, nil

}

// Start starts subscription manager
func (m *Manager) Start() error {
	log.Info("Start E2Manager")
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		err := m.watchE2Connections(ctx)
		if err != nil {
			return
		}
	}()

	return nil
}

func (m *Manager) sendIndicationOnStream(streamID broker.StreamID, ch chan e2api.Indication) {
	streamWriter, err := m.streams.GetWriter(streamID)
	if err != nil {
		log.Error(err)
		return
	}

	for msg := range ch {
		err := streamWriter.Send(msg)
		if err != nil {
			log.Warn(err)
			return
		}
	}
}

func (m *Manager) getRanFunction(serviceModelsInfo map[string]*topoapi.ServiceModelInfo) (*topoapi.RSMRanFunction, error) {
	for _, sm := range serviceModelsInfo {
		smName := strings.ToLower(sm.Name)
		if smName == string(m.serviceModel.Name) && sm.OID == oid {
			rsmRanFunction := &topoapi.RSMRanFunction{}
			for _, ranFunction := range sm.RanFunctions {
				if ranFunction.TypeUrl == ranFunction.GetTypeUrl() {
					err := prototypes.UnmarshalAny(ranFunction, rsmRanFunction)
					if err != nil {
						return nil, err
					}
					return rsmRanFunction, nil
				}
			}
		}
	}
	return nil, errors.New(errors.NotFound, "cannot retrieve ran functions")

}

func (m *Manager) createSubscription(ctx context.Context, e2nodeID topoapi.ID, eventTrigger e2sm_rsm.RsmRicindicationTriggerType) error {
	log.Info("Creating subscription for E2 node with ID:", e2nodeID)
	eventTriggerData, err := m.createEventTrigger(eventTrigger)
	if err != nil {
		return err
	}

	aspects, err := m.rnibClient.GetE2NodeAspects(ctx, e2nodeID)
	if err != nil {
		log.Warn(err)
		return err
	}

	_, err = m.getRanFunction(aspects.ServiceModels)
	if err != nil {
		log.Warn(err)
		return err
	}

	ch := make(chan e2api.Indication)
	node := m.e2client.Node(e2client.NodeID(e2nodeID))
	subName := fmt.Sprintf("onos-rsm-subscription-%s", eventTrigger)
	subSpec := e2api.SubscriptionSpec{
		EventTrigger: e2api.EventTrigger{
			Payload: eventTriggerData,
		},
		Actions: m.createSubscriptionActions(),
	}
	log.Debugf("subSpec: %v", subSpec)

	channelID, err := node.Subscribe(ctx, subName, subSpec, ch)
	if err != nil {
		log.Warn(err)
		return err
	}
	streamReader, err := m.streams.OpenReader(ctx, node, subName, channelID, subSpec)
	if err != nil {
		log.Warn(err)
		return err
	}

	go m.sendIndicationOnStream(streamReader.StreamID(), ch)
	monitor := monitoring.NewMonitor(monitoring.WithAppConfig(m.appConfig),
		monitoring.WithNode(node),
		monitoring.WithNodeID(e2nodeID),
		monitoring.WithStreamReader(streamReader),
		monitoring.WithRNIBClient(m.rnibClient),
		monitoring.WithRicIndicationTriggerType(eventTrigger))

	err = monitor.Start(ctx)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

func (m *Manager) watchE2Connections(ctx context.Context) error {
	ch := make(chan topoapi.Event)
	err := m.rnibClient.WatchE2Connections(ctx, ch)
	if err != nil {
		log.Warn(err)
		return err
	}

	for topoEvent := range ch {
		if topoEvent.Type == topoapi.EventType_ADDED || topoEvent.Type == topoapi.EventType_NONE {
			log.Infof("New E2 connection detected")
			relation := topoEvent.Object.Obj.(*topoapi.Object_Relation)
			e2NodeID := relation.Relation.TgtEntityID
			log.Debugf("E2NodeID %v connected", e2NodeID)
			rsmSupportedCfgs, err := m.rnibClient.GetSupportedSlicingConfigTypes(ctx, e2NodeID)
			if err != nil {
				return err
			}

			log.Debugf("RSM Supported Cfgs - %v", rsmSupportedCfgs)

			for _, cfg := range rsmSupportedCfgs {
				if cfg.GetSlicingConfigType() == topoapi.E2SmRsmCommand_E2_SM_RSM_COMMAND_EVENT_TRIGGERS {
					go func() {
						err := m.createSubscription(ctx, e2NodeID, e2sm_rsm.RsmRicindicationTriggerType_RSM_RICINDICATION_TRIGGER_TYPE_UPON_EMM_EVENT)
						if err != nil {
							log.Warn(err)
						}
					}()
					break
				}
				if cfg.GetSlicingConfigType() == topoapi.E2SmRsmCommand_E2_SM_RSM_COMMAND_UE_ASSOCIATE ||
					cfg.GetSlicingConfigType() == topoapi.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_CREATE ||
					cfg.GetSlicingConfigType() == topoapi.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_UPDATE ||
					cfg.GetSlicingConfigType() == topoapi.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_DELETE {
					m.CtrlReqChs[string(e2NodeID)] = make(chan *e2api.ControlMessage)
					go m.watchCtrlChan(ctx, e2NodeID)
				}
			}
			go m.watchSliceChange(ctx, e2NodeID)
			go m.watchSliceUEAssociation(ctx, e2NodeID)

			//go func() {
			//	err := m.createSubscription(ctx, e2NodeID, e2sm_rsm.RsmRicindicationTriggerType_RSM_RICINDICATION_TRIGGER_TYPE_PERIODIC_METRICS)
			//	if err != nil {
			//		log.Warn(err)
			//	}
			//}()
			//go m.watchRSMChanges(ctx, e2NodeID)
		}
	}

	return nil
}

func (m *Manager) watchCtrlChan(ctx context.Context, e2NodeID topoapi.ID) {
	for ctrlReqMsg := range m.CtrlReqChs[string(e2NodeID)] {
		go func(ctrlReqMsg *e2api.ControlMessage) {
			node := m.e2client.Node(e2client.NodeID(e2NodeID))
			ctrlRespMsg, err := node.Control(ctx, ctrlReqMsg)
			if err != nil {
				log.Warn("Error sending control message - %v", err)
			} else if ctrlRespMsg == nil {
				log.Warn("Ctrl Resp message is nil")
			}
		}(ctrlReqMsg)
	}
}

func (m *Manager) watchSliceChange(ctx context.Context, e2NodeID topoapi.ID) {
	//ch := make(chan store.Event)
	//err := m.sliceAssocStore.Watch(ctx, ch)
	//if err != nil {
	//	log.Errorf("Failed to watch slice change event")
	//	return
	//}
	//
	//for e := range ch {
	//	switch e.Type {
	//	case store.Created:
	//	case store.Updated:
	//	case store.Deleted:
	//	default:
	//	}
	//}
}

func (m *Manager) watchSliceUEAssociation(ctx context.Context, e2NodeID topoapi.ID) {

}

func (m *Manager) createEventTrigger(triggerType e2sm_rsm.RsmRicindicationTriggerType) ([]byte, error) {
	// period is not defined yet
	//reportPeriodMs := int32(0)
	//period, err := m.appConfig.GetReportPeriod()
	//if err != nil {
	//	return nil, err
	//}
	//
	//if triggerType == e2sm_rsm.RsmRicindicationTriggerType_RSM_RICINDICATION_TRIGGER_TYPE_PERIODIC_METRICS {
	//	reportPeriodMs = int32(period)
	//}

	eventTriggerDef, err := pdubuilder.CreateE2SmRsmEventTriggerDefinitionFormat1(triggerType)
	if err != nil {
		log.Warn(err)
		return nil, err
	}

	protoBytes, err := proto.Marshal(eventTriggerDef)
	if err != nil {
		log.Warn(err)
		return nil, err
	}

	return protoBytes, nil

}

func (m *Manager) createSubscriptionActions() []e2api.Action {
	actions := make([]e2api.Action, 0)
	action := &e2api.Action{
		ID:   int32(0),
		Type: e2api.ActionType_ACTION_TYPE_REPORT,

		SubsequentAction: &e2api.SubsequentAction{
			Type:       e2api.SubsequentActionType_SUBSEQUENT_ACTION_TYPE_WAIT,
			TimeToWait: e2api.TimeToWait_TIME_TO_WAIT_W1MS,
		},
	}
	actions = append(actions, *action)
	return actions

}

// Stop stops the subscription manager
func (m *Manager) Stop() error {
	panic("implement me")
}

var _ Node = &Manager{}

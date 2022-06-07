// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package e2

import (
	"context"
	"fmt"
	"strings"

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
	"github.com/onosproject/onos-rsm/pkg/nib/uenib"
	"google.golang.org/protobuf/proto"
)

var log = logging.GetLogger()

const (
	oid = "1.3.6.1.4.1.53148.1.1.2.102"
)

type Node interface {
	Start() error
	Stop() error
}

type Manager struct {
	appID                 string
	e2Client              e2client.Client
	rnibClient            rnib.TopoClient
	uenibClient           uenib.Client
	serviceModel          ServiceModelOptions
	appConfig             *appConfig.AppConfig
	streams               broker.Broker
	ctrlReqChsSliceCreate map[string]chan *CtrlMsg
	ctrlReqChsSliceUpdate map[string]chan *CtrlMsg
	ctrlReqChsSliceDelete map[string]chan *CtrlMsg
	ctrlReqChsUeAssociate map[string]chan *CtrlMsg
}

func NewManager(opts ...Option) (Manager, error) {
	log.Info("Init E2 Manager")
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
		appID:       options.App.AppID,
		e2Client:    e2Client,
		rnibClient:  options.App.RnibClient,
		uenibClient: options.App.UenibClient,
		serviceModel: ServiceModelOptions{
			Name:    options.ServiceModel.Name,
			Version: options.ServiceModel.Version,
		},
		appConfig:             options.App.AppConfig,
		streams:               options.App.Broker,
		ctrlReqChsSliceCreate: options.App.CtrlReqChsSliceCreate,
		ctrlReqChsSliceUpdate: options.App.CtrlReqChsSliceUpdate,
		ctrlReqChsSliceDelete: options.App.CtrlReqChsSliceDelete,
		ctrlReqChsUeAssociate: options.App.CtrlReqChsUeAssociate,
	}, nil
}

func (m *Manager) Start() error {
	log.Info("Start E2 Manager")
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

func (m *Manager) watchE2Connections(ctx context.Context) error {
	ch := make(chan topoapi.Event)
	err := m.rnibClient.WatchE2Connections(ctx, ch)
	if err != nil {
		log.Warn(err)
		return err
	}

	for topoEvent := range ch {
		log.Debugf("Received topo event: type %v, message %v", topoEvent.Type, topoEvent)
		switch topoEvent.Type {
		case topoapi.EventType_ADDED, topoapi.EventType_NONE:
			relation := topoEvent.Object.Obj.(*topoapi.Object_Relation)
			e2NodeID := relation.Relation.TgtEntityID
			if !m.rnibClient.HasRSMRANFunction(ctx, e2NodeID, oid) {
				log.Debugf("Received topo event does not have RSM RAN function - %v", topoEvent)
				continue
			}

			log.Debugf("New E2NodeID %v connected", e2NodeID)
			rsmSupportedCfgs, err := m.rnibClient.GetSupportedSlicingConfigTypes(ctx, e2NodeID)
			if err != nil {
				log.Warn(err)
				return err
			}

			log.Debugf("RSM supported configs: %v", rsmSupportedCfgs)
			for _, cfg := range rsmSupportedCfgs {
				switch cfg.SlicingConfigType {
				case topoapi.E2SmRsmCommand_E2_SM_RSM_COMMAND_EVENT_TRIGGERS:
					go func() {
						err := m.createSubscription(ctx, e2NodeID, e2sm_rsm.RsmRicindicationTriggerType_RSM_RICINDICATION_TRIGGER_TYPE_UPON_EMM_EVENT)
						if err != nil {
							log.Warn(err)
						}
					}()
				case topoapi.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_CREATE:
					m.ctrlReqChsSliceCreate[string(e2NodeID)] = make(chan *CtrlMsg)
					go m.watchCtrlSliceCreated(ctx, e2NodeID)
				case topoapi.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_UPDATE:
					m.ctrlReqChsSliceUpdate[string(e2NodeID)] = make(chan *CtrlMsg)
					go m.watchCtrlSliceUpdated(ctx, e2NodeID)
				case topoapi.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_DELETE:
					m.ctrlReqChsSliceDelete[string(e2NodeID)] = make(chan *CtrlMsg)
					go m.watchCtrlSliceDeleted(ctx, e2NodeID)
				case topoapi.E2SmRsmCommand_E2_SM_RSM_COMMAND_UE_ASSOCIATE:
					m.ctrlReqChsUeAssociate[string(e2NodeID)] = make(chan *CtrlMsg)
					go m.watchCtrlUEAssociate(ctx, e2NodeID)
				}
			}
		case topoapi.EventType_REMOVED:
			relation := topoEvent.Object.Obj.(*topoapi.Object_Relation)
			e2NodeID := relation.Relation.TgtEntityID
			if !m.rnibClient.HasRSMRANFunction(ctx, e2NodeID, oid) {
				log.Debugf("Received topo event does not have RSM RAN function - %v", topoEvent)
				continue
			}

			log.Infof("E2 node %v is disconnected", e2NodeID)
			// Clean up slice information from onos-topo
			duE2NodeID, err := m.rnibClient.GetTargetDUE2NodeID(ctx, e2NodeID)
			hasDU := true
			if err != nil {
				log.Debugf("e2Node %v was not connected to DU - maybe e2Node %v is DU", e2NodeID, e2NodeID)
				hasDU = false
			}

			if hasDU {
				err = m.rnibClient.DeleteRsmSliceList(ctx, duE2NodeID)
				if err != nil {
					log.Warn(err)
				}
			} else {
				err = m.rnibClient.DeleteRsmSliceList(ctx, e2NodeID)
				if err != nil {
					log.Warn(err)
				}
			}

			// Clean up UE information from uenib
			err = m.uenibClient.DeleteUEWithE2NodeID(ctx, string(e2NodeID))
			if err != nil {
				log.Warn(err)
			}
		}
	}

	return nil
}

func (m *Manager) createSubscription(ctx context.Context, e2nodeID topoapi.ID, eventTrigger e2sm_rsm.RsmRicindicationTriggerType) error {
	log.Info("Creating subscription for E2 node ID with: ", e2nodeID)
	eventTriggerData, err := m.createEventTrigger(eventTrigger)
	if err != nil {
		log.Warn(err)
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
	node := m.e2Client.Node(e2client.NodeID(e2nodeID))
	subName := fmt.Sprintf("%s-subscription-%s-%s", m.appID, e2nodeID, eventTrigger)
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
		monitoring.WithUENIBClient(m.uenibClient),
		monitoring.WithRicIndicationTriggerType(eventTrigger))

	err = monitor.Start(ctx)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
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
			Type:       e2api.SubsequentActionType_SUBSEQUENT_ACTION_TYPE_CONTINUE,
			TimeToWait: e2api.TimeToWait_TIME_TO_WAIT_ZERO,
		},
	}
	actions = append(actions, *action)
	return actions
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

func (m *Manager) watchCtrlSliceCreated(ctx context.Context, e2NodeID topoapi.ID) {
	for ctrlReqMsg := range m.ctrlReqChsSliceCreate[string(e2NodeID)] {
		log.Debugf("ctrlReqMsg: %v", ctrlReqMsg)
		node := m.e2Client.Node(e2client.NodeID(e2NodeID))
		ctrlRespMsg, err := node.Control(ctx, ctrlReqMsg.CtrlMsg)
		if err != nil {
			log.Warnf("Error sending control message - %v", err)
			ack := Ack{
				Success: false,
				Reason:  err.Error(),
			}
			ctrlReqMsg.AckCh <- ack
			continue
		} else if ctrlRespMsg == nil {
			log.Warn(" Ctrl Resp message is nil")
			ack := Ack{
				Success: false,
				Reason:  "Ctrl Resp message is nil",
			}
			ctrlReqMsg.AckCh <- ack
			continue
		}
		ack := Ack{
			Success: true,
		}
		ctrlReqMsg.AckCh <- ack
	}
}

func (m *Manager) watchCtrlSliceUpdated(ctx context.Context, e2NodeID topoapi.ID) {
	for ctrlReqMsg := range m.ctrlReqChsSliceUpdate[string(e2NodeID)] {
		log.Debugf("ctrlReqMsg: %v", ctrlReqMsg)
		node := m.e2Client.Node(e2client.NodeID(e2NodeID))
		ctrlRespMsg, err := node.Control(ctx, ctrlReqMsg.CtrlMsg)
		log.Debugf("ctrlRespMsg: %v", ctrlRespMsg)
		if err != nil {
			log.Warnf("Error sending control message - %v", err)
			ack := Ack{
				Success: false,
				Reason:  err.Error(),
			}
			ctrlReqMsg.AckCh <- ack
			continue
		} else if ctrlRespMsg == nil {
			log.Warn(" Ctrl Resp message is nil")
			ack := Ack{
				Success: false,
				Reason:  "Ctrl Resp message is nil",
			}
			ctrlReqMsg.AckCh <- ack
			continue
		}
		ack := Ack{
			Success: true,
		}
		ctrlReqMsg.AckCh <- ack
	}
}

func (m *Manager) watchCtrlSliceDeleted(ctx context.Context, e2NodeID topoapi.ID) {
	for ctrlReqMsg := range m.ctrlReqChsSliceDelete[string(e2NodeID)] {
		log.Debugf("ctrlReqMsg: %v", ctrlReqMsg)
		node := m.e2Client.Node(e2client.NodeID(e2NodeID))
		ctrlRespMsg, err := node.Control(ctx, ctrlReqMsg.CtrlMsg)
		log.Debugf("ctrlRespMsg: %v", ctrlRespMsg)
		if err != nil {
			log.Warnf("Error sending control message - %v", err)
			ack := Ack{
				Success: false,
				Reason:  err.Error(),
			}
			ctrlReqMsg.AckCh <- ack
			continue
		} else if ctrlRespMsg == nil {
			log.Warn(" Ctrl Resp message is nil")
			ack := Ack{
				Success: false,
				Reason:  "Ctrl Resp message is nil",
			}
			ctrlReqMsg.AckCh <- ack
			continue
		}
		ack := Ack{
			Success: true,
		}
		ctrlReqMsg.AckCh <- ack
	}
}

func (m *Manager) watchCtrlUEAssociate(ctx context.Context, e2NodeID topoapi.ID) {
	for ctrlReqMsg := range m.ctrlReqChsUeAssociate[string(e2NodeID)] {
		log.Debugf("ctrlReqMsg: %v", ctrlReqMsg)
		node := m.e2Client.Node(e2client.NodeID(e2NodeID))
		ctrlRespMsg, err := node.Control(ctx, ctrlReqMsg.CtrlMsg)
		log.Debugf("ctrlRespMsg: %v", ctrlRespMsg)
		if err != nil {
			log.Warnf("Error sending control message - %v", err)
			ack := Ack{
				Success: false,
				Reason:  err.Error(),
			}
			ctrlReqMsg.AckCh <- ack
			continue
		} else if ctrlRespMsg == nil {
			log.Warn(" Ctrl Resp message is nil")
			ack := Ack{
				Success: false,
				Reason:  "Ctrl Resp message is nil",
			}
			ctrlReqMsg.AckCh <- ack
			continue
		}
		ack := Ack{
			Success: true,
		}
		ctrlReqMsg.AckCh <- ack
	}
}

func (m *Manager) Stop() error {
	panic("implement me")
}

var _ Node = &Manager{}

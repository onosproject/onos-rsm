// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package monitoring

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	uenib_api "github.com/onosproject/onos-api/go/onos/uenib"
	e2sm_rsm "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-rsm-ies"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-rsm/pkg/broker"
	appConfig "github.com/onosproject/onos-rsm/pkg/config"
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	"github.com/onosproject/onos-rsm/pkg/nib/uenib"
	"google.golang.org/protobuf/proto"
)

var log = logging.GetLogger()

func NewMonitor(opts ...Option) *Monitor {
	options := Options{}
	for _, opt := range opts {
		opt.apply(&options)
	}

	return &Monitor{
		streamReader:           options.Monitor.StreamReader,
		appConfig:              options.App.AppConfig,
		nodeID:                 options.Monitor.NodeID,
		rnibClient:             options.App.RnibClient,
		uenibClient:            options.App.UenibClient,
		ricIndEventTriggerType: options.App.EventTriggerType,
	}
}

type Monitor struct {
	streamReader           broker.StreamReader
	appConfig              *appConfig.AppConfig
	nodeID                 topoapi.ID
	rnibClient             rnib.TopoClient
	uenibClient            uenib.Client
	ricIndEventTriggerType e2sm_rsm.RsmRicindicationTriggerType
}

func (m *Monitor) Start(ctx context.Context) error {
	errCh := make(chan error)
	go func() {
		for {
			indMsg, err := m.streamReader.Recv(ctx)
			if err != nil {
				errCh <- err
			}
			err = m.processIndication(ctx, indMsg, m.nodeID)
			if err != nil {
				errCh <- err
			}
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *Monitor) processIndication(ctx context.Context, indMsg e2api.Indication, nodeID topoapi.ID) error {
	indHeader := e2sm_rsm.E2SmRsmIndicationHeader{}
	indPayload := e2sm_rsm.E2SmRsmIndicationMessage{}

	err := proto.Unmarshal(indMsg.Header, &indHeader)
	if err != nil {
		return err
	}

	err = proto.Unmarshal(indMsg.Payload, &indPayload)
	if err != nil {
		return err
	}

	if indPayload.GetIndicationMessageFormat1() != nil {
		err = m.processMetricTypeMessage(ctx, indHeader.GetIndicationHeaderFormat1(), indPayload.GetIndicationMessageFormat1())
		if err != nil {
			return err
		}
	}

	if indPayload.GetIndicationMessageFormat2() != nil {
		err = m.processEmmEventMessage(ctx, indHeader.GetIndicationHeaderFormat1(), indPayload.GetIndicationMessageFormat2(), string(nodeID))
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Monitor) processMetricTypeMessage(ctx context.Context, indHdr *e2sm_rsm.E2SmRsmIndicationHeaderFormat1, indMsg *e2sm_rsm.E2SmRsmIndicationMessageFormat1) error {

	log.Debugf("Received indication message (Metric) hdr: %v / msg: %v", indHdr, indMsg)

	return nil
}

func (m *Monitor) processEmmEventMessage(ctx context.Context, indHdr *e2sm_rsm.E2SmRsmIndicationHeaderFormat1, indMsg *e2sm_rsm.E2SmRsmIndicationMessageFormat2, cuNodeID string) error {
	log.Debugf("Received indication message (EMM) hdr: %v / msg: %v", indHdr, indMsg)

	var CuUeF1apID, DuUeF1apID, RanUeNgapID, AmfUeNgapID int64
	var EnbUeS1apID int32
	bIDList := make([]*uenib_api.BearerId, 0)

	duNodeID, err := m.rnibClient.GetTargetDUE2NodeID(ctx, topoapi.ID(cuNodeID))
	log.Debugf("Cu ID %v - Du ID %v", cuNodeID, duNodeID)
	if err != nil {
		log.Warn(err)
	}

	for _, id := range indMsg.GetUeIdlist() {
		if id.GetCuUeF1ApId() != nil {
			CuUeF1apID = id.GetCuUeF1ApId().Value
		} else if id.GetDuUeF1ApId() != nil {
			DuUeF1apID = id.GetDuUeF1ApId().Value
		} else if id.GetRanUeNgapId() != nil {
			RanUeNgapID = id.GetRanUeNgapId().Value
		} else if id.GetAmfUeNgapId() != nil {
			AmfUeNgapID = id.GetAmfUeNgapId().Value
		} else if id.GetEnbUeS1ApId() != nil {
			EnbUeS1apID = id.GetEnbUeS1ApId().Value
		}
	}

	for _, bID := range indMsg.GetBearerId() {
		if bID.GetDrbId().GetFiveGdrbId() != nil {
			// for 5G
			flowMapToDrb := make([]*uenib_api.QoSflowLevelParameters, 0)

			for _, fItem := range bID.GetDrbId().GetFiveGdrbId().GetFlowsMapToDrb() {
				if fItem.GetNonDynamicFiveQi() != nil {
					flowMapToDrb = append(flowMapToDrb, &uenib_api.QoSflowLevelParameters{
						QosFlowLevelParameters: &uenib_api.QoSflowLevelParameters_NonDynamicFiveQi{
							NonDynamicFiveQi: &uenib_api.NonDynamicFiveQi{
								FiveQi: &uenib_api.FiveQi{
									Value: fItem.GetNonDynamicFiveQi().GetFiveQi().GetValue(),
								},
							},
						},
					})
				} else if fItem.GetDynamicFiveQi() != nil {
					flowMapToDrb = append(flowMapToDrb, &uenib_api.QoSflowLevelParameters{
						QosFlowLevelParameters: &uenib_api.QoSflowLevelParameters_DynamicFiveQi{
							DynamicFiveQi: &uenib_api.DynamicFiveQi{
								PriorityLevel:    fItem.GetDynamicFiveQi().GetPriorityLevel(),
								PacketDelayBudge: fItem.GetDynamicFiveQi().GetPriorityLevel(),
								PacketErrorRate:  fItem.GetDynamicFiveQi().GetPacketErrorRate(),
							},
						},
					})
				}
			}
			uenibBID := &uenib_api.BearerId{
				BearerId: &uenib_api.BearerId_DrbId{
					DrbId: &uenib_api.DrbId{
						DrbId: &uenib_api.DrbId_FiveGdrbId{
							FiveGdrbId: &uenib_api.FiveGDrbId{
								Value: bID.GetDrbId().GetFiveGdrbId().GetValue(),
								Qfi: &uenib_api.Qfi{
									Value: bID.GetDrbId().GetFiveGdrbId().GetQfi().Value,
								},
								FlowsMapToDrb: flowMapToDrb,
							},
						},
					},
				},
			}
			bIDList = append(bIDList, uenibBID)
		} else if bID.GetDrbId().GetFourGdrbId() != nil {
			// for 4G
			uenibBID := &uenib_api.BearerId{
				BearerId: &uenib_api.BearerId_DrbId{
					DrbId: &uenib_api.DrbId{
						DrbId: &uenib_api.DrbId_FourGdrbId{
							FourGdrbId: &uenib_api.FourGDrbId{
								Value: bID.GetDrbId().GetFourGdrbId().GetValue(),
								Qci: &uenib_api.Qci{
									Value: bID.GetDrbId().GetFourGdrbId().GetQci().GetValue(),
								},
							},
						},
					},
				},
			}
			bIDList = append(bIDList, uenibBID)
		}
	}

	switch indMsg.GetTriggerType() {
	case e2sm_rsm.RsmEmmTriggerType_RSM_EMM_TRIGGER_TYPE_UE_ATTACH, e2sm_rsm.RsmEmmTriggerType_RSM_EMM_TRIGGER_TYPE_HAND_IN_UE_ATTACH:
		// ToDo: Add logic to get GlobalUEID here after SMO is integrated - future
		rsmUE := &uenib_api.RsmUeInfo{
			GlobalUeID: uuid.New().String(),
			UeIdList: &uenib_api.UeIdentity{
				CuUeF1apID: &uenib_api.CuUeF1ApID{
					Value: CuUeF1apID,
				},
				DuUeF1apID: &uenib_api.DuUeF1ApID{
					Value: DuUeF1apID,
				},
				RANUeNgapID: &uenib_api.RanUeNgapID{
					Value: RanUeNgapID,
				},
				AMFUeNgapID: &uenib_api.AmfUeNgapID{
					Value: AmfUeNgapID,
				},
				EnbUeS1apID: &uenib_api.EnbUeS1ApID{
					Value: EnbUeS1apID,
				},
			},
			BearerIdList: bIDList,
			CellGlobalId: indHdr.GetCgi().String(),
			CuE2NodeId:   cuNodeID,
			DuE2NodeId:   string(duNodeID),
			SliceList:    make([]*uenib_api.SliceInfo, 0),
		}
		log.Debugf("pushed rsmUE: %v", rsmUE)
		err := m.uenibClient.AddUE(ctx, rsmUE)
		// ToDo: add ue on ue store

		if err != nil {
			return err
		}
	case e2sm_rsm.RsmEmmTriggerType_RSM_EMM_TRIGGER_TYPE_UE_DETACH, e2sm_rsm.RsmEmmTriggerType_RSM_EMM_TRIGGER_TYPE_HAND_OUT_UE_ATTACH:
		// ToDo: delete ue from ue store
		switch indMsg.GetPrefferedUeIdtype() {
		case e2sm_rsm.UeIdType_UE_ID_TYPE_CU_UE_F1_AP_ID:
			err := m.uenibClient.DeleteUEWithPreferredID(ctx, cuNodeID, uenib_api.UeIdType_UE_ID_TYPE_CU_UE_F1_AP_ID, CuUeF1apID)
			if err != nil {
				return err
			}
		case e2sm_rsm.UeIdType_UE_ID_TYPE_DU_UE_F1_AP_ID:
			err := m.uenibClient.DeleteUEWithPreferredID(ctx, cuNodeID, uenib_api.UeIdType_UE_ID_TYPE_DU_UE_F1_AP_ID, DuUeF1apID)
			if err != nil {
				return err
			}
		case e2sm_rsm.UeIdType_UE_ID_TYPE_RAN_UE_NGAP_ID:
			err := m.uenibClient.DeleteUEWithPreferredID(ctx, cuNodeID, uenib_api.UeIdType_UE_ID_TYPE_RAN_UE_NGAP_ID, RanUeNgapID)
			if err != nil {
				return err
			}
		case e2sm_rsm.UeIdType_UE_ID_TYPE_AMF_UE_NGAP_ID:
			err := m.uenibClient.DeleteUEWithPreferredID(ctx, cuNodeID, uenib_api.UeIdType_UE_ID_TYPE_AMF_UE_NGAP_ID, AmfUeNgapID)
			if err != nil {
				return err
			}
		case e2sm_rsm.UeIdType_UE_ID_TYPE_ENB_UE_S1_AP_ID:
			err := m.uenibClient.DeleteUEWithPreferredID(ctx, cuNodeID, uenib_api.UeIdType_UE_ID_TYPE_ENB_UE_S1_AP_ID, int64(EnbUeS1apID))
			if err != nil {
				return err
			}
		default:
			return errors.NewNotSupported(fmt.Sprintf("Unknown preferred ID type: %v", indMsg.GetPrefferedUeIdtype()))
		}
	default:
		return errors.NewNotSupported(fmt.Sprintf("Unknown EMM trigger type: %v", indMsg.GetTriggerType()))
	}

	return nil
}

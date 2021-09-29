// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package northbound

import (
	"context"
	"fmt"
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	rsmapi "github.com/onosproject/onos-api/go/onos/rsm"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	e2sm_rsm "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-rsm-ies"
	e2sm_v2_ies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-v2-ies"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-lib-go/pkg/logging/service"
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	"github.com/onosproject/onos-rsm/pkg/southbound/e2"
	"github.com/onosproject/onos-rsm/pkg/store"
	"google.golang.org/grpc"
	"strconv"
)

var log = logging.GetLogger("northbound")

// NewService returns the new NBI service
func NewService(ctrlReqChs map[string]chan *e2api.ControlMessage,
	rnibClient rnib.Client, ueStore store.Store,
	sliceStore store.Store, sliceAssocStore store.Store, ctrlMsgHandler e2.ControlMessageHandler) service.Service {
	return &Service{
		ctrlReqChs:      ctrlReqChs,
		rnibClient:      rnibClient,
		ueStore:         ueStore,
		sliceStore:      sliceStore,
		sliceAssocStore: sliceAssocStore,
		ctrlMsgHandler:  ctrlMsgHandler,
	}
}

// Service represents the NBI service for RSM xAPP
type Service struct {
	ctrlReqChs      map[string]chan *e2api.ControlMessage
	rnibClient      rnib.Client
	ueStore         store.Store
	sliceStore      store.Store
	sliceAssocStore store.Store
	ctrlMsgHandler  e2.ControlMessageHandler
}

// Register registers the NBI service
func (s Service) Register(r *grpc.Server) {
	server := &Server{
		ctrlReqChs:      s.ctrlReqChs,
		rnibClient:      s.rnibClient,
		ueStore:         s.ueStore,
		sliceStore:      s.sliceStore,
		sliceAssocStore: s.sliceAssocStore,
		ctrlMsgHandler:  s.ctrlMsgHandler,
	}
	rsmapi.RegisterRsmServer(r, server)
}

// Server is a struct for NBI server
type Server struct {
	ctrlReqChs      map[string]chan *e2api.ControlMessage
	rnibClient      rnib.Client
	ueStore         store.Store
	sliceStore      store.Store
	sliceAssocStore store.Store
	ctrlMsgHandler  e2.ControlMessageHandler
}

// GetSlices gets slices
func (s Server) GetSlices(ctx context.Context, request *rsmapi.GetSlicesRequest) (*rsmapi.GetSliceResponse, error) {
	log.Infof("Called GetSlices: %v", request)
	return &rsmapi.GetSliceResponse{
		Ack: &rsmapi.Ack{
			Success: false,
			Cause:   "not implemented yet",
		},
	}, nil
}

// CreateSlice creats a slice
func (s Server) CreateSlice(ctx context.Context, request *rsmapi.CreateSliceRequest) (*rsmapi.CreateSliceResponse, error) {
	log.Infof("Called CreateSlice: %v", request)
	sliceID, err := strconv.Atoi(request.SliceId)
	if err != nil {
		return &rsmapi.CreateSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to convert slice id to int - %v", err),
			},
		}, nil
	}
	weightInt, err := strconv.Atoi(request.Weight)
	if err != nil {
		return &rsmapi.CreateSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to convert weight to int - %v", err),
			},
		}, nil
	}
	weight := int32(weightInt)

	cmdType := e2sm_rsm.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_CREATE
	var sliceSchedulerType e2sm_rsm.SchedulerType
	switch request.SchedulerType {
	case rsmapi.SchedulerType_SCHEDULER_TYPE_ROUND_ROBIN:
		sliceSchedulerType = e2sm_rsm.SchedulerType_SCHEDULER_TYPE_ROUND_ROBIN
	case rsmapi.SchedulerType_SCHEDULER_TYPE_PROPORTIONALLY_FAIR:
		sliceSchedulerType = e2sm_rsm.SchedulerType_SCHEDULER_TYPE_PROPORTIONALLY_FAIR
	case rsmapi.SchedulerType_SCHEDULER_TYPE_QOS_BASED:
		sliceSchedulerType = e2sm_rsm.SchedulerType_SCHEDULER_TYPE_QOS_BASED
	default:
		sliceSchedulerType = e2sm_rsm.SchedulerType_SCHEDULER_TYPE_ROUND_ROBIN
	}

	var sliceType e2sm_rsm.SliceType
	switch request.SliceType {
	case rsmapi.SliceType_SLICE_TYPE_DL_SLICE:
		sliceType = e2sm_rsm.SliceType_SLICE_TYPE_DL_SLICE
	case rsmapi.SliceType_SLICE_TYPE_UL_SLICE:
		sliceType = e2sm_rsm.SliceType_SLICE_TYPE_UL_SLICE
	default:
		sliceType = e2sm_rsm.SliceType_SLICE_TYPE_DL_SLICE
	}

	sliceConfig := &e2sm_rsm.SliceConfig{
		SliceId: &e2sm_rsm.SliceId{
			Value: int64(sliceID),
		},
		SliceConfigParameters: &e2sm_rsm.SliceParameters{
			SchedulerType: sliceSchedulerType,
			Weight:        &weight,
		},
		SliceType: sliceType,
	}
	ctrlMsg, err := s.ctrlMsgHandler.CreateControlRequest(cmdType, sliceConfig, nil)
	if err != nil {
		return &rsmapi.CreateSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to create the control message - %v", err),
			},
		}, nil
	}

	hasSliceItem := s.rnibClient.HasRsmSliceItemAspect(ctx, topoapi.ID(request.E2NodeId), request.SliceId)

	if hasSliceItem {
		return &rsmapi.CreateSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("Slice ID %v is already added", sliceID),
			},
		}, nil
	}

	s.ctrlReqChs[request.E2NodeId] <- ctrlMsg

	w, err := strconv.Atoi(request.Weight)
	if err != nil {
		return &rsmapi.CreateSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to convert weight value - %v", err),
			},
		}, nil
	}

	value := &topoapi.RSMSlicingItem{
		ID: request.SliceId,
		SliceDesc: "Test",
		SliceParameters: &topoapi.RSMSliceParameters{
			SchedulerType: topoapi.RSMSchedulerType(request.SchedulerType),
			Weight: int32(w),
		},
		SliceType: topoapi.RSMSliceType(request.SliceType),
	}

	err = s.rnibClient.AddRsmSliceItemAspect(ctx, topoapi.ID(request.E2NodeId), value)
	if err != nil {
		return &rsmapi.CreateSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to create slice information to onos-topo although control message was sent: %v", err),
			},
		}, nil
	}

	return &rsmapi.CreateSliceResponse{
		Ack: &rsmapi.Ack{
			Success: true,
		},
	}, nil
}

// UpdateSlice updates a slice
func (s Server) UpdateSlice(ctx context.Context, request *rsmapi.UpdateSliceRequest) (*rsmapi.UpdateSliceResponse, error) {
	log.Infof("Called UpdateSlice: %v", request)
	sliceID, err := strconv.Atoi(request.SliceId)
	if err != nil {
		return &rsmapi.UpdateSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to convert slice id to int - %v", err),
			},
		}, nil
	}
	weightInt, err := strconv.Atoi(request.Weight)
	if err != nil {
		return &rsmapi.UpdateSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to convert weight to int - %v", err),
			},
		}, nil
	}
	weight := int32(weightInt)

	cmdType := e2sm_rsm.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_UPDATE
	var sliceSchedulerType e2sm_rsm.SchedulerType
	switch request.SchedulerType {
	case rsmapi.SchedulerType_SCHEDULER_TYPE_ROUND_ROBIN:
		sliceSchedulerType = e2sm_rsm.SchedulerType_SCHEDULER_TYPE_ROUND_ROBIN
	case rsmapi.SchedulerType_SCHEDULER_TYPE_PROPORTIONALLY_FAIR:
		sliceSchedulerType = e2sm_rsm.SchedulerType_SCHEDULER_TYPE_PROPORTIONALLY_FAIR
	case rsmapi.SchedulerType_SCHEDULER_TYPE_QOS_BASED:
		sliceSchedulerType = e2sm_rsm.SchedulerType_SCHEDULER_TYPE_QOS_BASED
	default:
		sliceSchedulerType = e2sm_rsm.SchedulerType_SCHEDULER_TYPE_ROUND_ROBIN
	}

	var sliceType e2sm_rsm.SliceType
	switch request.SliceType {
	case rsmapi.SliceType_SLICE_TYPE_DL_SLICE:
		sliceType = e2sm_rsm.SliceType_SLICE_TYPE_DL_SLICE
	case rsmapi.SliceType_SLICE_TYPE_UL_SLICE:
		sliceType = e2sm_rsm.SliceType_SLICE_TYPE_UL_SLICE
	default:
		sliceType = e2sm_rsm.SliceType_SLICE_TYPE_DL_SLICE
	}

	sliceConfig := &e2sm_rsm.SliceConfig{
		SliceId: &e2sm_rsm.SliceId{
			Value: int64(sliceID),
		},
		SliceConfigParameters: &e2sm_rsm.SliceParameters{
			SchedulerType: sliceSchedulerType,
			Weight:        &weight,
		},
		SliceType: sliceType,
	}
	ctrlMsg, err := s.ctrlMsgHandler.CreateControlRequest(cmdType, sliceConfig, nil)
	if err != nil {
		return &rsmapi.UpdateSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to create the control message - %v", err),
			},
		}, nil
	}

	hasSliceItem := s.rnibClient.HasRsmSliceItemAspect(ctx, topoapi.ID(request.E2NodeId), request.SliceId)

	if !hasSliceItem {
		return &rsmapi.UpdateSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("no slice ID %v in node %v", sliceID, request.E2NodeId),
			},
		}, nil
	}

	s.ctrlReqChs[request.E2NodeId] <- ctrlMsg

	w, err := strconv.Atoi(request.Weight)
	if err != nil {
		return &rsmapi.UpdateSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to convert weight value - %v", err),
			},
		}, nil	}

	value := &topoapi.RSMSlicingItem{
		ID: request.SliceId,
		SliceDesc: "Test",
		SliceParameters: &topoapi.RSMSliceParameters{
			SchedulerType: topoapi.RSMSchedulerType(request.SchedulerType),
			Weight: int32(w),
		},
		SliceType: topoapi.RSMSliceType(request.SliceType),
	}

	err = s.rnibClient.UpdateRsmSliceItemAspect(ctx, topoapi.ID(request.E2NodeId), value)
	if err != nil {
		return &rsmapi.UpdateSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to update slice information to onos-topo although control message was sent: %v", err),
			},
		}, nil
	}

	return &rsmapi.UpdateSliceResponse{
		Ack: &rsmapi.Ack{
			Success: true,
		},
	}, nil
}

// DeleteSlice deletes a slice
func (s Server) DeleteSlice(ctx context.Context, request *rsmapi.DeleteSliceRequest) (*rsmapi.DeleteSliceResponse, error) {
	log.Infof("Called DeleteSlice: %v", request)
	sliceID, err := strconv.Atoi(request.SliceId)
	if err != nil {
		return &rsmapi.DeleteSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to convert slice id to int - %v", err),
			},
		}, nil
	}
	var sliceType e2sm_rsm.SliceType
	switch request.SliceType {
	case rsmapi.SliceType_SLICE_TYPE_DL_SLICE:
		sliceType = e2sm_rsm.SliceType_SLICE_TYPE_DL_SLICE
	case rsmapi.SliceType_SLICE_TYPE_UL_SLICE:
		sliceType = e2sm_rsm.SliceType_SLICE_TYPE_UL_SLICE
	default:
		sliceType = e2sm_rsm.SliceType_SLICE_TYPE_DL_SLICE
	}
	cmdType := e2sm_rsm.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_DELETE
	sliceConfig := &e2sm_rsm.SliceConfig{
		SliceId: &e2sm_rsm.SliceId{
			Value: int64(sliceID),
		},
		SliceType: sliceType,
	}
	ctrlMsg, err := s.ctrlMsgHandler.CreateControlRequest(cmdType, sliceConfig, nil)
	if err != nil {
		return &rsmapi.DeleteSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to create the control message - %v", err),
			},
		}, nil
	}
	hasSliceItem := s.rnibClient.HasRsmSliceItemAspect(ctx, topoapi.ID(request.E2NodeId), request.SliceId)

	if !hasSliceItem {
		return &rsmapi.DeleteSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("no slice ID %v in node %v", sliceID, request.E2NodeId),
			},
		}, nil
	}

	s.ctrlReqChs[request.E2NodeId] <- ctrlMsg

	err = s.rnibClient.DeleteRsmSliceItemAspect(ctx, topoapi.ID(request.GetE2NodeId()), request.SliceId)
	if err != nil {
		return &rsmapi.DeleteSliceResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to delete slice information from onos-topo although control message was sent: %v", err),
			},
		}, nil
	}

	return &rsmapi.DeleteSliceResponse{
		Ack: &rsmapi.Ack{
			Success: true,
		},
	}, nil
}

// GetUeSliceAssociation gets slice association
func (s Server) GetUeSliceAssociation(ctx context.Context, request *rsmapi.GetUeSliceAssociationRequest) (*rsmapi.GetUeSliceAssociationResponse, error) {
	log.Infof("Called GetUESliceAssociation: %v", request)
	return &rsmapi.GetUeSliceAssociationResponse{
		Ack: &rsmapi.Ack{
			Success: false,
			Cause:   "not implemented yet",
		},
	}, nil
}

// SetUeSliceAssociation sets a slice association
func (s Server) SetUeSliceAssociation(ctx context.Context, request *rsmapi.SetUeSliceAssociationRequest) (*rsmapi.SetUeSliceAssociationResponse, error) {
	log.Infof("Called SetUeSliceAssociation: %v", request)
	cmdType := e2sm_rsm.E2SmRsmCommand_E2_SM_RSM_COMMAND_UE_ASSOCIATE
	dlSliceID, err := strconv.Atoi(request.DlSliceId)
	if err != nil {
		return &rsmapi.SetUeSliceAssociationResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to convert slice id to int - %v", err),
			},
		}, nil
	}
	hasUlSliceID := false
	ulSliceID := 0
	if request.UlSliceId != "" {
		ulSliceID, err = strconv.Atoi(request.UlSliceId)
		if err != nil {
			return &rsmapi.SetUeSliceAssociationResponse{
				Ack: &rsmapi.Ack{
					Success: false,
					Cause:   fmt.Sprintf("failed to convert slice id to int - %v", err),
				},
			}, nil
		}
		hasUlSliceID = true
	}

	var reqUeID int64
	hasValidUeID := false
	for _, ueid := range request.UeId {
		if ueid.GetType() == rsmapi.UeIdType_UE_ID_TYPE_DU_UE_F1_AP_ID {
			hasValidUeID = true
			id, err := strconv.Atoi(ueid.GetUeId())
			if err != nil {
				return &rsmapi.SetUeSliceAssociationResponse{
					Ack: &rsmapi.Ack{
						Success: false,
						Cause:   fmt.Sprintf("failed to convert ue id to int - %v", err),
					},
				}, nil
			}
			reqUeID = int64(id)
		}
	}

	if !hasValidUeID {
		return &rsmapi.SetUeSliceAssociationResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   "need du-ue-f1ap-id",
			},
		}, nil
	}

	ueID := &e2sm_rsm.UeIdentity{
		UeIdentity: &e2sm_rsm.UeIdentity_DuUeF1ApId{
			DuUeF1ApId: &e2sm_rsm.DuUeF1ApId{
				Value: reqUeID,
			},
		},
	}

	bearerIDs := make([]*e2sm_rsm.BearerId, 0)
	bearerID := &e2sm_rsm.BearerId{
		BearerId: &e2sm_rsm.BearerId_DrbId{
			DrbId: &e2sm_rsm.DrbId{
				DrbId: &e2sm_rsm.DrbId_FourGdrbId{
					FourGdrbId: &e2sm_rsm.FourGDrbId{
						Value: int32(1),
						Qci: &e2sm_v2_ies.Qci{
							Value: int32(1),
						},
					},
				},
			},
		},
	}
	bearerIDs = append(bearerIDs, bearerID)

	sliceAssoc := &e2sm_rsm.SliceAssociate{
		DownLinkSliceId: &e2sm_rsm.SliceIdassoc{
			Value: int64(dlSliceID),
		},
		UeId:     ueID,
		BearerId: bearerIDs,
	}
	if hasUlSliceID {
		sliceAssoc.UplinkSliceId = &e2sm_rsm.SliceIdassoc{
			Value: int64(ulSliceID),
		}
	}

	ctrlMsg, err := s.ctrlMsgHandler.CreateControlRequest(cmdType, nil, sliceAssoc)
	if err != nil {
		return &rsmapi.SetUeSliceAssociationResponse{
			Ack: &rsmapi.Ack{
				Success: false,
				Cause:   fmt.Sprintf("failed to create the control message - %v", err),
			},
		}, nil
	}
	s.ctrlReqChs[request.E2NodeId] <- ctrlMsg
	return &rsmapi.SetUeSliceAssociationResponse{
		Ack: &rsmapi.Ack{
			Success: true,
		},
	}, nil
}

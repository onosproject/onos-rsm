// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"context"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	prototypes "github.com/gogo/protobuf/types"
	"github.com/google/uuid"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	e2sm_rsm_ies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-rsm-ies"
	toposdk "github.com/onosproject/onos-ric-sdk-go/pkg/topo"
)

type CUDUType int

const (
	CUDUTypeCU CUDUType = iota
	CUDUTypeDU
)

func (t CUDUType) String() string {
	return [...]string{"CUDUTypeCU", "CUDUTypeDU"}[t]
}

func NewRnibClient() (toposdk.Client, error) {
	return toposdk.NewClient()
}

func CreateMockE2Node(index int, nodeType CUDUType) error {
	client, err := NewRnibClient()
	if err != nil {
		return err
	}

	var mockNodeID string
	switch nodeType {
	case CUDUTypeCU:
		mockNodeID = GetMockCUE2NodeID(index)
	case CUDUTypeDU:
		mockNodeID = GetMockDUE2NodeID(index)
	}

	obj := &topoapi.Object{
		ID:   topoapi.ID(mockNodeID),
		UUID: topoapi.UUID(uuid.New().String()),
		Type: topoapi.Object_ENTITY,
		Obj: &topoapi.Object_Entity{
			Entity: &topoapi.Entity{
				KindID: topoapi.E2NODE,
			},
		},
		Aspects: make(map[string]*types.Any),
	}

	topoE2NodeAspect := &topoapi.E2Node{
		ServiceModels: make(map[string]*topoapi.ServiceModelInfo),
	}

	rsmSMInfo := &topoapi.ServiceModelInfo{
		OID:          RSMSmOID,
		Name:         RSMSmName,
		RanFunctions: make([]*types.Any, 0),
	}

	rsmSmRanFunc := &topoapi.RSMRanFunction{}
	rsmConfigItems := make([]*topoapi.RSMSupportedSlicingConfigItem, 0)
	switch nodeType {
	case CUDUTypeCU:
		rsmConfig := &topoapi.RSMSupportedSlicingConfigItem{
			SlicingConfigType: topoapi.E2SmRsmCommand(e2sm_rsm_ies.E2SmRsmCommand_E2_SM_RSM_COMMAND_EVENT_TRIGGERS),
		}
		rsmConfigItems = append(rsmConfigItems, rsmConfig)
	case CUDUTypeDU:
		rsmConfig1 := &topoapi.RSMSupportedSlicingConfigItem{
			SlicingConfigType: topoapi.E2SmRsmCommand(e2sm_rsm_ies.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_CREATE),
		}
		rsmConfig2 := &topoapi.RSMSupportedSlicingConfigItem{
			SlicingConfigType: topoapi.E2SmRsmCommand(e2sm_rsm_ies.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_UPDATE),
		}
		rsmConfig3 := &topoapi.RSMSupportedSlicingConfigItem{
			SlicingConfigType: topoapi.E2SmRsmCommand(e2sm_rsm_ies.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_DELETE),
		}
		rsmConfig4 := &topoapi.RSMSupportedSlicingConfigItem{
			SlicingConfigType: topoapi.E2SmRsmCommand(e2sm_rsm_ies.E2SmRsmCommand_E2_SM_RSM_COMMAND_UE_ASSOCIATE),
		}
		rsmConfigItems = append(rsmConfigItems, rsmConfig1)
		rsmConfigItems = append(rsmConfigItems, rsmConfig2)
		rsmConfigItems = append(rsmConfigItems, rsmConfig3)
		rsmConfigItems = append(rsmConfigItems, rsmConfig4)
	}
	rsmNodeSlicingCapItem := &topoapi.RSMNodeSlicingCapabilityItem{
		MaxNumberOfSlicesDl:    MaxNumberOfSlicesDl,
		MaxNumberOfSlicesUl:    MaxNumberOfSlicesUl,
		SlicingType:            topoapi.RSMSlicingType_SLICING_TYPE_STATIC,
		MaxNumberOfUesPerSlice: MaxNumberOfUesPerSlice,
		SupportedConfig:        rsmConfigItems,
	}
	rsmSmRanFunc.RicSlicingNodeCapabilityList = append(rsmSmRanFunc.RicSlicingNodeCapabilityList, rsmNodeSlicingCapItem)

	rsmSmRanFuncAny, err := prototypes.MarshalAny(rsmSmRanFunc)
	if err != nil {
		return err
	}

	rsmSMInfo.RanFunctions = append(rsmSMInfo.RanFunctions, rsmSmRanFuncAny)

	topoE2NodeAspect.ServiceModels[RSMSmOID] = rsmSMInfo

	jm := jsonpb.Marshaler{}
	writer := bytes.Buffer{}
	err = jm.Marshal(&writer, topoE2NodeAspect)
	if err != nil {
		return err
	}

	obj.Aspects[proto.MessageName(topoE2NodeAspect)] = &types.Any{
		TypeUrl: proto.MessageName(topoE2NodeAspect),
		Value:   writer.Bytes(),
	}

	return client.Create(context.Background(), obj)
}

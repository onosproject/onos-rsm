// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package rnib

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/onosproject/onos-api/go/onos/rsm"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"strings"

	"github.com/onosproject/onos-lib-go/pkg/errors"

	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	toposdk "github.com/onosproject/onos-ric-sdk-go/pkg/topo"
)

var log = logging.GetLogger("rnib")

func NewClient() (TopoClient, error) {
	sdkClient, err := toposdk.NewClient()
	if err != nil {
		return nil, err
	}
	return &topoClient{
		client: sdkClient,
	}, nil
}

type TopoClient interface {
	WatchE2Connections(ctx context.Context, ch chan topoapi.Event) error
	GetSupportedSlicingConfigTypes(ctx context.Context, nodeID topoapi.ID) ([]*topoapi.RSMSupportedSlicingConfigItem, error)
	GetE2NodeAspects(ctx context.Context, nodeID topoapi.ID) (*topoapi.E2Node, error)
	GetTargetDUE2NodeID(ctx context.Context, cuE2NodeID topoapi.ID) (topoapi.ID, error)
	GetSourceCUE2NodeID(ctx context.Context, duE2NodeID topoapi.ID) (topoapi.ID, error)
	HasRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, sliceID string, sliceType rsm.SliceType) bool
	AddRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, msg *topoapi.RSMSlicingItem) error
	SetRsmSliceListAspect(ctx context.Context, nodeID topoapi.ID, msg *topoapi.RSMSliceItemList) error
	UpdateRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, msg *topoapi.RSMSlicingItem) error
	DeleteRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, sliceID string) error
	GetRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, sliceID string, sliceType rsm.SliceType) (*topoapi.RSMSlicingItem, error)
	GetRsmSliceItemAspects(ctx context.Context, nodeID topoapi.ID) ([]*topoapi.RSMSlicingItem, error)
	DeleteRsmSliceList(ctx context.Context, nodeID topoapi.ID) error
	GetRSMSliceItemAspectsForAllDUs(ctx context.Context) (map[string][]*topoapi.RSMSlicingItem, error)
}

type topoClient struct {
	client toposdk.Client
}

func (t *topoClient) DeleteRsmSliceList(ctx context.Context, nodeID topoapi.ID) error {
	object, err := t.client.Get(ctx, nodeID)
	if err != nil {
		return err
	}

	aspectKey := proto.MessageName(&topoapi.RSMSliceItemList{})
	delete(object.Aspects, aspectKey)

	err = t.client.Update(ctx, object)
	if err != nil {
		return err
	}

	return nil
}

func (t *topoClient) GetRsmSliceItemAspects(ctx context.Context, nodeID topoapi.ID) ([]*topoapi.RSMSlicingItem, error) {
	rsmSliceList, err := t.GetRsmSliceListAspect(ctx, nodeID)
	if err != nil {
		return nil, errors.NewNotFound("node %v has no slices", nodeID)
	}

	return rsmSliceList.GetRsmSliceList(), nil
}

func (t *topoClient) GetRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, sliceID string, sliceType rsm.SliceType) (*topoapi.RSMSlicingItem, error) {
	rsmSliceList, err := t.GetRsmSliceListAspect(ctx, nodeID)
	if err != nil {
		return nil, errors.NewNotFound("node %v has no slices", nodeID)
	}

	var topoSliceType topoapi.RSMSliceType
	switch sliceType {
	case rsm.SliceType_SLICE_TYPE_DL_SLICE:
		topoSliceType = topoapi.RSMSliceType_SLICE_TYPE_DL_SLICE
	case rsm.SliceType_SLICE_TYPE_UL_SLICE:
		topoSliceType = topoapi.RSMSliceType_SLICE_TYPE_UL_SLICE
	default:
		return nil, errors.NewNotSupported(fmt.Sprintf("slice type %v does not support", sliceType.String()))
	}

	for _, item := range rsmSliceList.GetRsmSliceList() {
		if item.GetID() == sliceID && item.GetSliceType() == topoSliceType {
			return item, nil
		}
	}

	return nil, errors.NewNotFound("node %v does not have slice %v (%v)", nodeID, sliceID, sliceType.String())
}

func (t *topoClient) DeleteRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, sliceID string) error {
	rsmSliceList, err := t.GetRsmSliceListAspect(ctx, nodeID)
	if err != nil {
		return errors.NewNotFound("node %v has no slices", nodeID)
	}

	for i := 0; i < len(rsmSliceList.GetRsmSliceList()); i++ {
		if rsmSliceList.GetRsmSliceList()[i].GetID() == sliceID {
			rsmSliceList.RsmSliceList = append(rsmSliceList.RsmSliceList[:i], rsmSliceList.RsmSliceList[i+1:]...)
			break
		}
	}

	err = t.SetRsmSliceListAspect(ctx, nodeID, rsmSliceList)
	if err != nil {
		return err
	}
	return nil
}

func (t *topoClient) UpdateRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, msg *topoapi.RSMSlicingItem) error {
	err := t.DeleteRsmSliceItemAspect(ctx, nodeID, msg.GetID())
	if err != nil {
		return err
	}
	err = t.AddRsmSliceItemAspect(ctx, nodeID, msg)
	if err != nil {
		return err
	}
	return nil
}

func (t *topoClient) SetRsmSliceListAspect(ctx context.Context, nodeID topoapi.ID, msg *topoapi.RSMSliceItemList) error {
	object, err := t.client.Get(ctx, nodeID)
	if err != nil {
		return err
	}

	err = object.SetAspect(msg)
	if err != nil {
		return err
	}
	err = t.client.Update(ctx, object)
	if err != nil {
		return err
	}

	return nil
}

func (t *topoClient) AddRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, msg *topoapi.RSMSlicingItem) error {
	rsmSliceList, err := t.GetRsmSliceListAspect(ctx, nodeID)
	if err != nil {
		rsmSliceList = &topoapi.RSMSliceItemList{
			RsmSliceList: make([]*topoapi.RSMSlicingItem, 0),
		}
	}

	rsmSliceList.RsmSliceList = append(rsmSliceList.RsmSliceList, msg)
	err = t.SetRsmSliceListAspect(ctx, nodeID, rsmSliceList)
	if err != nil {
		return err
	}
	return nil
}

func (t *topoClient) HasRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, sliceID string, sliceType rsm.SliceType) bool {
	rsmSliceList, err := t.GetRsmSliceListAspect(ctx, nodeID)
	if err != nil {
		return false
	}

	var topoSliceType topoapi.RSMSliceType
	switch sliceType {
	case rsm.SliceType_SLICE_TYPE_DL_SLICE:
		topoSliceType = topoapi.RSMSliceType_SLICE_TYPE_DL_SLICE
	case rsm.SliceType_SLICE_TYPE_UL_SLICE:
		topoSliceType = topoapi.RSMSliceType_SLICE_TYPE_UL_SLICE
	default:
		return false
	}

	for _, item := range rsmSliceList.GetRsmSliceList() {
		if item.GetID() == sliceID && item.GetSliceType() == topoSliceType {
			return true
		}
	}

	return false
}

func (t *topoClient) GetRsmSliceListAspect(ctx context.Context, nodeID topoapi.ID) (*topoapi.RSMSliceItemList, error) {
	object, err := t.client.Get(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	value := &topoapi.RSMSliceItemList{}
	err = object.GetAspect(value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (t *topoClient) GetSupportedSlicingConfigTypes(ctx context.Context, nodeID topoapi.ID) ([]*topoapi.RSMSupportedSlicingConfigItem, error) {
	result := make([]*topoapi.RSMSupportedSlicingConfigItem, 0)
	e2Node, err := t.GetE2NodeAspects(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	for smName, sm := range e2Node.GetServiceModels() {
		for _, ranFunc := range sm.GetRanFunctions() {
			rsmRanFunc := &topoapi.RSMRanFunction{}
			err = proto.Unmarshal(ranFunc.GetValue(), rsmRanFunc)
			if err != nil {
				log.Debugf("RanFunction for SM - %v, URL - %v does not have RSM RAN Function Description:\n%v", smName, ranFunc.GetTypeUrl(), err)
				continue
			}
			for _, cap := range rsmRanFunc.GetRicSlicingNodeCapabilityList() {
				for i := 0; i < len(cap.GetSupportedConfig()); i++ {
					result = append(result, cap.GetSupportedConfig()[i])
				}
			}
		}
	}
	return result, nil
}

func (t *topoClient) GetE2NodeAspects(ctx context.Context, nodeID topoapi.ID) (*topoapi.E2Node, error) {
	object, err := t.client.Get(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	e2Node := &topoapi.E2Node{}
	err = object.GetAspect(e2Node)
	if err != nil {
		return nil, err
	}

	return e2Node, nil
}

func (t *topoClient) WatchE2Connections(ctx context.Context, ch chan topoapi.Event) error {
	err := t.client.Watch(ctx, ch, toposdk.WithWatchFilters(getControlRelationFilter()))
	if err != nil {
		return err
	}
	return nil
}

func getControlRelationFilter() *topoapi.Filters {
	controlRelationFilter := &topoapi.Filters{
		KindFilter: &topoapi.Filter{
			Filter: &topoapi.Filter_Equal_{
				Equal_: &topoapi.EqualFilter{
					Value: topoapi.CONTROLS,
				},
			},
		},
	}
	return controlRelationFilter
}

func (t *topoClient) GetTargetDUE2NodeID(ctx context.Context, cuE2NodeID topoapi.ID) (topoapi.ID, error) {
	// ToDo: When auto-discovery comes in, it should be changed
	objects, err := t.client.List(ctx)
	if err != nil {
		return "", err
	}

	for _, obj := range objects {
		log.Debugf("Relation: %v", obj.GetEntity())
		if obj.GetEntity() != nil && obj.GetEntity().GetKindID() == topoapi.E2NODE {
			if cuE2NodeID != obj.GetID() {
				nodeID := fmt.Sprintf("%s/%s", strings.Split(string(cuE2NodeID), "/")[0], strings.Split(string(cuE2NodeID), "/")[1])
				tgtNodeID := fmt.Sprintf("%s/%s", strings.Split(string(obj.GetID()), "/")[0], strings.Split(string(obj.GetID()), "/")[1])
				if nodeID == tgtNodeID {
					return obj.GetID(), nil
				}
			}
		}
	}

	return "", errors.NewNotFound(fmt.Sprintf("DU-ID not found (CU-ID: %v)", cuE2NodeID))
}

func (t *topoClient) GetSourceCUE2NodeID(ctx context.Context, duE2NodeID topoapi.ID) (topoapi.ID, error) {
	// ToDo: When auto-discovery comes in, it should be changed
	objects, err := t.client.List(ctx)
	if err != nil {
		return "", err
	}

	for _, obj := range objects {
		log.Debugf("Relation: %v", obj.GetEntity())
		if obj.GetEntity() != nil && obj.GetEntity().GetKindID() == topoapi.E2NODE {
			if duE2NodeID != obj.GetID() {
				nodeID := fmt.Sprintf("%s/%s", strings.Split(string(duE2NodeID), "/")[0], strings.Split(string(duE2NodeID), "/")[1])
				tgtNodeID := fmt.Sprintf("%s/%s", strings.Split(string(obj.GetID()), "/")[0], strings.Split(string(obj.GetID()), "/")[1])
				if nodeID == tgtNodeID {
					return obj.GetID(), nil
				}
			}
		}
	}

	return "", errors.NewNotFound(fmt.Sprintf("CU-ID not found (DU-ID: %v)", duE2NodeID))
}

func (t *topoClient) GetRSMSliceItemAspectsForAllDUs(ctx context.Context) (map[string][]*topoapi.RSMSlicingItem, error) {
	results := make(map[string][]*topoapi.RSMSlicingItem)
	objects, err := t.client.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, obj := range objects {
		if obj.GetEntity() != nil && obj.GetEntity().GetKindID() == topoapi.E2NODE {
			if obj.GetEntity() != nil && obj.GetEntity().GetKindID() == topoapi.E2NODE && len(strings.Split(string(obj.GetID()), "/")) == 4 && strings.Split(string(obj.GetID()), "/")[2] == "3" {
				value := &topoapi.RSMSliceItemList{}
				err = obj.GetAspect(value)
				if err != nil {
					return nil, err
				}
				results[string(obj.GetID())] = value.GetRsmSliceList()
			}
		}
	}

	return results, nil
}

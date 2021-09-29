// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package rnib

import (
	"context"
	"github.com/gogo/protobuf/proto"
	"github.com/onosproject/onos-lib-go/pkg/logging"

	"github.com/onosproject/onos-lib-go/pkg/errors"

	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	toposdk "github.com/onosproject/onos-ric-sdk-go/pkg/topo"
)

var log = logging.GetLogger("rnib")

// TopoClient R-NIB client interface
type TopoClient interface {
	WatchE2Connections(ctx context.Context, ch chan topoapi.Event) error
	GetCells(ctx context.Context, nodeID topoapi.ID) ([]*topoapi.E2Cell, error)
	GetE2NodeAspects(ctx context.Context, nodeID topoapi.ID) (*topoapi.E2Node, error)
	E2NodeIDs(ctx context.Context) ([]topoapi.ID, error)
	GetSupportedSlicingConfigTypes(ctx context.Context, nodeID topoapi.ID) ([]*topoapi.RSMSupportedSlicingConfigItem, error)
	AddRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, msg *topoapi.RSMSlicingItem) error
	UpdateRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, msg *topoapi.RSMSlicingItem) error
	DeleteRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, sliceID string) error
	GetRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, sliceID string) (*topoapi.RSMSlicingItem, error)
	HasRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, sliceID string) bool
	SetRsmSliceListAspect(ctx context.Context, nodeID topoapi.ID, msg *topoapi.RSMSliceItemList) error
	GetRsmSliceListAspect(ctx context.Context, nodeID topoapi.ID) (*topoapi.RSMSliceItemList, error)
}

// NewClient creates a new topo SDK client
func NewClient() (TopoClient, error) {
	sdkClient, err := toposdk.NewClient()
	if err != nil {
		return &Client{}, err
	}
	cl := &Client{
		client: sdkClient,
	}

	return cl, nil

}

// Client topo SDK client
type Client struct {
	client toposdk.Client
}

// E2NodeIDs lists all of connected E2 nodes
func (c *Client) E2NodeIDs(ctx context.Context) ([]topoapi.ID, error) {
	objects, err := c.client.List(ctx, toposdk.WithListFilters(getControlRelationFilter()))
	if err != nil {
		return nil, err
	}

	e2NodeIDs := make([]topoapi.ID, len(objects))
	for _, object := range objects {
		relation := object.Obj.(*topoapi.Object_Relation)
		e2NodeID := relation.Relation.TgtEntityID
		e2NodeIDs = append(e2NodeIDs, e2NodeID)
	}

	return e2NodeIDs, nil
}

// GetE2NodeAspects gets E2 node aspects
func (c *Client) GetE2NodeAspects(ctx context.Context, nodeID topoapi.ID) (*topoapi.E2Node, error) {
	object, err := c.client.Get(ctx, nodeID)
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

// GetSupportedSlicingConfigTypes gets supported slicing config types
func (c *Client) GetSupportedSlicingConfigTypes(ctx context.Context, nodeID topoapi.ID) ([]*topoapi.RSMSupportedSlicingConfigItem, error) {
	result := make([]*topoapi.RSMSupportedSlicingConfigItem, 0)
	e2Node, err := c.GetE2NodeAspects(ctx, nodeID)
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

// AddRsmSliceItemAspect adds rsm slice item aspect
func (c *Client) AddRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, msg *topoapi.RSMSlicingItem) error {
	rsmSliceList, err := c.GetRsmSliceListAspect(ctx, nodeID)
	if err != nil {
		rsmSliceList = &topoapi.RSMSliceItemList{
			RsmSliceList: make([]*topoapi.RSMSlicingItem, 0),
		}
	}

	rsmSliceList.RsmSliceList = append(rsmSliceList.RsmSliceList, msg)
	err = c.SetRsmSliceListAspect(ctx, nodeID, rsmSliceList)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) UpdateRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, msg *topoapi.RSMSlicingItem) error {
	err := c.DeleteRsmSliceItemAspect(ctx, nodeID, msg.GetID())
	if err != nil {
		return err
	}
	err = c.AddRsmSliceItemAspect(ctx, nodeID, msg)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, sliceID string) error {
	rsmSliceList, err := c.GetRsmSliceListAspect(ctx, nodeID)
	if err != nil {
		return errors.NewNotFound("node %v has no slices", nodeID)
	}

	for i := 0; i < len(rsmSliceList.GetRsmSliceList()); i++ {
		if rsmSliceList.GetRsmSliceList()[i].GetID() == sliceID {
			rsmSliceList.RsmSliceList = append(rsmSliceList.RsmSliceList[:i], rsmSliceList.RsmSliceList[i+1:]...)
			break
		}
	}

	err = c.SetRsmSliceListAspect(ctx, nodeID, rsmSliceList)
	if err != nil {
		return err
	}
	return nil
}

// GetRsmSliceItemAspect gets rsm slice item aspect
func (c *Client) GetRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, sliceID string) (*topoapi.RSMSlicingItem, error) {
	rsmSliceList, err := c.GetRsmSliceListAspect(ctx, nodeID)
	if err != nil {
		return nil, errors.NewNotFound("node %v has no slices", nodeID)
	}

	for _, item := range rsmSliceList.GetRsmSliceList() {
		if item.GetID() == sliceID {
			return item, nil
		}
	}

	return nil, errors.NewNotFound("node %v does not have slice %v", nodeID, sliceID)
}

func (c *Client) HasRsmSliceItemAspect(ctx context.Context, nodeID topoapi.ID, sliceID string) bool {
	rsmSliceList, err := c.GetRsmSliceListAspect(ctx, nodeID)
	if err != nil {
		return false
	}

	for _, item := range rsmSliceList.GetRsmSliceList() {
		if item.GetID() == sliceID {
			return true
		}
	}

	return false
}

// SetRsmSliceListAspect sets Slice list
func (c *Client) SetRsmSliceListAspect(ctx context.Context, nodeID topoapi.ID, msg *topoapi.RSMSliceItemList) error {
	object, err := c.client.Get(ctx, nodeID)
	if err != nil {
		return err
	}

	err = object.SetAspect(msg)
	if err != nil {
		return err
	}
	err = c.client.Update(ctx, object)
	if err != nil {
		return err
	}

	return nil
}

// GetRsmSliceListAspect gets slice list aspect
func (c *Client) GetRsmSliceListAspect(ctx context.Context, nodeID topoapi.ID) (*topoapi.RSMSliceItemList, error) {
	object, err := c.client.Get(ctx, nodeID)
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

// GetCells get list of cells for each E2 node
func (c *Client) GetCells(ctx context.Context, nodeID topoapi.ID) ([]*topoapi.E2Cell, error) {
	filter := &topoapi.Filters{
		RelationFilter: &topoapi.RelationFilter{SrcId: string(nodeID),
			RelationKind: topoapi.CONTAINS,
			TargetKind:   ""}}

	objects, err := c.client.List(ctx, toposdk.WithListFilters(filter))
	if err != nil {
		return nil, err
	}
	var cells []*topoapi.E2Cell
	for _, obj := range objects {
		targetEntity := obj.GetEntity()
		if targetEntity.GetKindID() == topoapi.E2CELL {
			cellObject := &topoapi.E2Cell{}
			err = obj.GetAspect(cellObject)
			if err != nil {
				return nil, err
			}
			cells = append(cells, cellObject)
		}
	}

	if len(cells) == 0 {
		return nil, errors.New(errors.NotFound, "there is no cell to subscribe for e2 node %s", nodeID)
	}

	return cells, nil
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

// WatchE2Connections watch e2 node connection changes
func (c *Client) WatchE2Connections(ctx context.Context, ch chan topoapi.Event) error {
	err := c.client.Watch(ctx, ch, toposdk.WithWatchFilters(getControlRelationFilter()))
	if err != nil {
		return err
	}
	return nil
}

var _ TopoClient = &Client{}

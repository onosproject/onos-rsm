// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package uenib

import (
	"bytes"
	"context"
	"fmt"
	"github.com/atomix/go-client/pkg/client/errors"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/onosproject/onos-api/go/onos/uenib"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-lib-go/pkg/southbound"
	"io"
)

var log = logging.GetLogger("uenib")

func NewClient(ctx context.Context, certPath string, keyPath string, uenibAddr string) (Client, error) {
	conn, err := southbound.Connect(ctx, uenibAddr, certPath, keyPath)
	if err != nil {
		return nil, err
	}

	return &client{
		client: uenib.NewUEServiceClient(conn),
	}, nil
}

type Client interface {
	HasUE(ctx context.Context, ue *uenib.RsmUeInfo) bool
	AddUE(ctx context.Context, ue *uenib.RsmUeInfo) error
	UpdateUE(ctx context.Context, ue *uenib.RsmUeInfo) error
	DeleteUE(ctx context.Context, id string) error
	DeleteUEWithPreferredID(ctx context.Context, cuNodeID string, preferredType uenib.UeIdType, ueID int64) error
	GetUEWithPreferredID(ctx context.Context, cuNodeID string, preferredType uenib.UeIdType, ueID int64) (*uenib.RsmUeInfo, error)
	GetUEWithGlobalID(ctx context.Context, id string) (*uenib.RsmUeInfo, error)
	GetUEs(ctx context.Context) ([]*uenib.RsmUeInfo, error)
	GetUenibUEWithPreferredID(ctx context.Context, cuNodeID string, preferredType uenib.UeIdType, ueID int64) (uenib.UE, error)
	DeleteUEWithE2NodeID(ctx context.Context, e2NodeID string) error
}

type client struct {
	client uenib.UEServiceClient
}

func (c *client) HasUE(ctx context.Context, ue *uenib.RsmUeInfo) bool {
	list, err := c.GetUEs(ctx)
	if err != nil {
		log.Debug("onos-uenib has no UE")
		return false
	}

	for _, item := range list {
		if item.GetGlobalUeID() == ue.GetGlobalUeID() &&
			item.GetUeIdList().GetDuUeF1apID().Value == ue.GetUeIdList().GetDuUeF1apID().Value &&
			item.GetUeIdList().GetCuUeF1apID().Value == ue.GetUeIdList().GetCuUeF1apID().Value &&
			item.GetUeIdList().GetRANUeNgapID().Value == ue.GetUeIdList().GetRANUeNgapID().Value &&
			item.GetUeIdList().GetEnbUeS1apID().Value == ue.GetUeIdList().GetEnbUeS1apID().Value &&
			item.GetUeIdList().GetPreferredIDType().String() == ue.GetUeIdList().GetPreferredIDType().String() &&
			item.GetCellGlobalId() == ue.GetCellGlobalId() &&
			item.GetCuE2NodeId() == ue.GetCuE2NodeId() && item.GetDuE2NodeId() == ue.GetDuE2NodeId() {
			return true
		}
	}

	log.Debugf("onos-uenib has UE %v", *ue)
	return false
}

func (c *client) AddUE(ctx context.Context, ue *uenib.RsmUeInfo) error {
	log.Debugf("received ue: %v", ue)

	if c.HasUE(ctx, ue) {
		return errors.NewAlreadyExists(fmt.Sprintf("UE already exists - UE: %v", *ue))
	}

	uenibObj := uenib.UE{
		ID:      uenib.ID(ue.GetGlobalUeID()),
		Aspects: make(map[string]*types.Any),
	}

	jm := jsonpb.Marshaler{}
	writer := bytes.Buffer{}
	err := jm.Marshal(&writer, ue)
	if err != nil {
		return err
	}

	uenibObj.Aspects[proto.MessageName(ue)] = &types.Any{
		TypeUrl: proto.MessageName(ue),
		Value:   writer.Bytes(),
	}

	req := &uenib.CreateUERequest{
		UE: uenibObj,
	}

	resp, err := c.client.CreateUE(ctx, req)
	if err != nil {
		log.Warn(err)
	}

	log.Debugf("CreateUE Resp: %v", resp)
	return nil
}

func (c *client) UpdateUE(ctx context.Context, ue *uenib.RsmUeInfo) error {
	if !c.HasUE(ctx, ue) {
		return errors.NewNotFound(fmt.Sprintf("UE not found - UE: %v", *ue))
	}

	uenibObj := uenib.UE{
		ID:      uenib.ID(ue.GetGlobalUeID()),
		Aspects: make(map[string]*types.Any),
	}

	jm := jsonpb.Marshaler{}
	writer := bytes.Buffer{}
	err := jm.Marshal(&writer, ue)
	if err != nil {
		return err
	}

	uenibObj.Aspects[proto.MessageName(ue)] = &types.Any{
		TypeUrl: proto.MessageName(ue),
		Value:   writer.Bytes(),
	}

	req := &uenib.UpdateUERequest{
		UE: uenibObj,
	}

	resp, err := c.client.UpdateUE(ctx, req)
	if err != nil {
		return err
	}

	log.Debugf("Update UE Resp: %v", resp)

	return nil
}

func (c *client) DeleteUE(ctx context.Context, id string) error {
	log.Debugf("received id: %v", id)

	ue, err := c.getUenibUEWithGlobalUeID(ctx, id)
	if err != nil {
		return err
	}
	rsmUE := &uenib.RsmUeInfo{}
	err = ue.GetAspect(rsmUE)
	if err != nil {
		return err
	}

	if !c.HasUE(ctx, rsmUE) {
		return errors.NewNotFound(fmt.Sprintf("UE not found - UE: %v", *rsmUE))
	}

	req := &uenib.DeleteUERequest{
		ID: uenib.ID(rsmUE.GetGlobalUeID()),
	}

	resp, err := c.client.DeleteUE(ctx, req)
	if err != nil {
		return err
	}

	log.Debugf("DeleteUE Resp: %v", resp)
	return nil
}

func (c *client) DeleteUEWithPreferredID(ctx context.Context, cuNodeID string, preferredType uenib.UeIdType, ueID int64) error {
	log.Debugf("received CUID: %v, preferredType: %v, ueID: %v", cuNodeID, preferredType, ueID)
	ue, err := c.GetUenibUEWithPreferredID(ctx, cuNodeID, preferredType, ueID)
	if err != nil {
		return err
	}
	rsmUE := &uenib.RsmUeInfo{}
	err = ue.GetAspect(rsmUE)
	if err != nil {
		return err
	}
	return c.DeleteUE(ctx, rsmUE.GetGlobalUeID())
}

func (c *client) DeleteUEWithE2NodeID(ctx context.Context, e2NodeID string) error {
	ues, err := c.GetUEs(ctx)
	if err != nil {
		return err
	}

	for _, ue := range ues {
		if ue.GetDuE2NodeId() == e2NodeID || ue.GetCuE2NodeId() == e2NodeID {
			err = c.DeleteUE(ctx, ue.GetGlobalUeID())
			if err != nil {
				log.Warn(err)
			}
		}
	}
	return err
}

func (c *client) GetUEWithGlobalID(ctx context.Context, id string) (*uenib.RsmUeInfo, error) {
	list, err := c.GetUEs(ctx)
	if err != nil {
		return nil, err
	}

	for _, item := range list {
		if item.GetGlobalUeID() == id {
			return item, nil
		}
	}

	return nil, errors.NewNotFound(fmt.Sprintf("Global UE ID %v does not exist", id))
}

func (c *client) GetUEWithPreferredID(ctx context.Context, cuNodeID string, preferredType uenib.UeIdType, ueID int64) (*uenib.RsmUeInfo, error) {
	ue, err := c.GetUenibUEWithPreferredID(ctx, cuNodeID, preferredType, ueID)
	if err != nil {
		return &uenib.RsmUeInfo{}, err
	}
	rsmUE := &uenib.RsmUeInfo{}
	err = ue.GetAspect(rsmUE)
	if err != nil {
		return &uenib.RsmUeInfo{}, err
	}
	return rsmUE, err
}

func (c *client) GetUEs(ctx context.Context) ([]*uenib.RsmUeInfo, error) {
	result := make([]*uenib.RsmUeInfo, 0)

	stream, err := c.client.ListUEs(ctx, &uenib.ListUERequest{})
	if err != nil {
		return []*uenib.RsmUeInfo{}, err
	}

	for {
		object, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return []*uenib.RsmUeInfo{}, err
		}
		ue := object.GetUE()
		rsmUE := &uenib.RsmUeInfo{}
		err = ue.GetAspect(rsmUE)
		if err != nil {
			return []*uenib.RsmUeInfo{}, err
		}

		result = append(result, rsmUE)
	}
	return result, nil
}

func (c *client) getUenibUEWithGlobalUeID(ctx context.Context, id string) (uenib.UE, error) {
	req := &uenib.GetUERequest{
		ID: uenib.ID(id),
	}

	resp, err := c.client.GetUE(ctx, req)
	if err != nil {
		return uenib.UE{}, err
	}
	return resp.GetUE(), nil
}

func (c *client) GetUenibUEWithPreferredID(ctx context.Context, cuNodeID string, preferredType uenib.UeIdType, ueID int64) (uenib.UE, error) {
	var result uenib.UE
	hasUE := false
	stream, err := c.client.ListUEs(ctx, &uenib.ListUERequest{})
	if err != nil {
		return uenib.UE{}, err
	}

	for {
		object, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return uenib.UE{}, err
		}

		rsmUE := &uenib.RsmUeInfo{}
		uenibUE := object.GetUE()
		err = uenibUE.GetAspect(rsmUE)
		if err != nil {
			return uenib.UE{}, err
		}
		switch preferredType {
		case uenib.UeIdType_UE_ID_TYPE_CU_UE_F1_AP_ID:
			if rsmUE.GetCuE2NodeId() == cuNodeID && rsmUE.GetUeIdList().GetCuUeF1apID().Value == ueID {
				result = object.GetUE()
				hasUE = true
			}
		case uenib.UeIdType_UE_ID_TYPE_DU_UE_F1_AP_ID:
			if rsmUE.GetCuE2NodeId() == cuNodeID && rsmUE.GetUeIdList().GetDuUeF1apID().Value == ueID {
				result = object.GetUE()
				hasUE = true
			}
		case uenib.UeIdType_UE_ID_TYPE_RAN_UE_NGAP_ID:
			if rsmUE.GetCuE2NodeId() == cuNodeID && rsmUE.GetUeIdList().GetRANUeNgapID().Value == ueID {
				result = object.GetUE()
				hasUE = true
			}
		case uenib.UeIdType_UE_ID_TYPE_AMF_UE_NGAP_ID:
			if rsmUE.GetCuE2NodeId() == cuNodeID && rsmUE.GetUeIdList().GetAMFUeNgapID().Value == ueID {
				result = object.GetUE()
				hasUE = true
			}
		case uenib.UeIdType_UE_ID_TYPE_ENB_UE_S1_AP_ID:
			if rsmUE.GetCuE2NodeId() == cuNodeID && rsmUE.GetUeIdList().GetEnbUeS1apID().Value == int32(ueID) {
				result = object.GetUE()
				hasUE = true
			}
		default:
			return uenib.UE{}, errors.NewNotSupported(fmt.Sprintf("ID type %v is not allowed", preferredType.String()))
		}
	}
	if !hasUE {
		return uenib.UE{}, errors.NewNotFound(fmt.Sprintf("UE ID %v %v does not exist in CU %v", preferredType.String(), ueID, cuNodeID))
	}
	return result, nil
}

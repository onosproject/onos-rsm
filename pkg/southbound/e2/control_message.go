// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package e2

import (
	"fmt"
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	"github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/pdubuilder"
	e2sm_rsm "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-rsm-ies"
	"google.golang.org/protobuf/proto"
)

// NewControlMessageHandler creates the new control message handler
func NewControlMessageHandler() ControlMessageHandler {
	return ControlMessageHandler{}
}

// ControlMessageHandler is a struct to handle control message
type ControlMessageHandler struct {
}

// CreateControlRequest returns the control request message
func (c *ControlMessageHandler) CreateControlRequest(cmdType e2sm_rsm.E2SmRsmCommand, sliceConfig *e2sm_rsm.SliceConfig, sliceAssoc *e2sm_rsm.SliceAssociate) (*e2api.ControlMessage, error) {
	hdr, err := c.CreateControlHeader(cmdType)
	if err != nil {
		return nil, err
	}

	payload, err := c.CreateControlPayload(cmdType, sliceConfig, sliceAssoc)
	if err != nil {
		return nil, err
	}

	return &e2api.ControlMessage{
		Header:  hdr,
		Payload: payload,
	}, nil
}

// CreateControlHeader creates the control message header
func (c *ControlMessageHandler) CreateControlHeader(cmdType e2sm_rsm.E2SmRsmCommand) ([]byte, error) {
	hdr := pdubuilder.CreateE2SmRsmControlHeader(cmdType)
	hdrProtoBytes, err := proto.Marshal(hdr)
	if err != nil {
		return nil, err
	}
	return hdrProtoBytes, nil
}

// CreateControlPayload creates the control message payload
func (c *ControlMessageHandler) CreateControlPayload(cmdType e2sm_rsm.E2SmRsmCommand, sliceConfig *e2sm_rsm.SliceConfig, sliceAssoc *e2sm_rsm.SliceAssociate) ([]byte, error) {
	var err error
	var msg *e2sm_rsm.E2SmRsmControlMessage
	var msgProtoBytes []byte
	switch cmdType {
	case e2sm_rsm.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_CREATE:
		msg = pdubuilder.CreateE2SmRsmControlMessageSliceCreate(sliceConfig)
		msgProtoBytes, err = proto.Marshal(msg)
		if err != nil {
			return nil, err
		}
	case e2sm_rsm.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_UPDATE:
		msg = pdubuilder.CreateE2SmRsmControlMessageSliceUpdate(sliceConfig)
		msgProtoBytes, err = proto.Marshal(msg)
		if err != nil {
			return nil, err
		}
	case e2sm_rsm.E2SmRsmCommand_E2_SM_RSM_COMMAND_SLICE_DELETE:
		msg = pdubuilder.CreateE2SmRsmControlMessageSliceDelete(sliceConfig.GetSliceId().GetValue(), sliceConfig.GetSliceType())
		msgProtoBytes, err = proto.Marshal(msg)
		if err != nil {
			return nil, err
		}
	case e2sm_rsm.E2SmRsmCommand_E2_SM_RSM_COMMAND_UE_ASSOCIATE:
		msg = pdubuilder.CreateE2SmRsmControlMessageSliceAssociate(sliceAssoc)
		msgProtoBytes, err = proto.Marshal(msg)
		if err != nil {
			return nil, err
		}
	case e2sm_rsm.E2SmRsmCommand_E2_SM_RSM_COMMAND_EVENT_TRIGGERS:
		// ToDo: check what it is for
		err := fmt.Errorf("%s (%v)", "Unsupported message type", cmdType)
		log.Error(err)
	default:
		err := fmt.Errorf("%s (%v)", "wrong E2SmRsmCommand type", cmdType)
		log.Error(err)
	}

	return msgProtoBytes, err
}

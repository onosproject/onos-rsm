// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package northbound

import (
	"context"
	rsmapi "github.com/onosproject/onos-api/go/onos/rsm"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-lib-go/pkg/logging/service"
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	"github.com/onosproject/onos-rsm/pkg/nib/uenib"
	"google.golang.org/grpc"
)

func NewService(rnibClient rnib.TopoClient, uenibClient uenib.Client, rsmReqCh chan *RsmMsg) service.Service {
	return &Service{
		rnibClient:  rnibClient,
		uenibClient: uenibClient,
		rsmReqCh:    rsmReqCh,
	}
}

type Service struct {
	rnibClient  rnib.TopoClient
	uenibClient uenib.Client
	rsmReqCh    chan *RsmMsg
}

func (s Service) Register(r *grpc.Server) {
	server := &Server{
		rnibClient:  s.rnibClient,
		uenibClient: s.uenibClient,
		rsmReqCh:    s.rsmReqCh,
	}
	rsmapi.RegisterRsmServer(r, server)
}

type Server struct {
	rnibClient  rnib.TopoClient
	uenibClient uenib.Client
	rsmReqCh    chan *RsmMsg
}

func (s Server) CreateSlice(ctx context.Context, request *rsmapi.CreateSliceRequest) (*rsmapi.CreateSliceResponse, error) {
	ackCh := make(chan Ack)
	msg := &RsmMsg{
		NodeID:  topoapi.ID(request.E2NodeId),
		Message: request,
		AckCh:   ackCh,
	}
	go func(msg *RsmMsg) {
		s.rsmReqCh <- msg
	}(msg)

	ack := <-ackCh
	return &rsmapi.CreateSliceResponse{
		Ack: &rsmapi.Ack{
			Success: ack.Success,
			Cause:   ack.Reason,
		},
	}, nil
}

func (s Server) UpdateSlice(ctx context.Context, request *rsmapi.UpdateSliceRequest) (*rsmapi.UpdateSliceResponse, error) {
	ackCh := make(chan Ack)
	msg := &RsmMsg{
		NodeID:  topoapi.ID(request.E2NodeId),
		Message: request,
		AckCh:   ackCh,
	}
	go func(msg *RsmMsg) {
		s.rsmReqCh <- msg
	}(msg)

	ack := <-ackCh
	return &rsmapi.UpdateSliceResponse{
		Ack: &rsmapi.Ack{
			Success: ack.Success,
			Cause:   ack.Reason,
		},
	}, nil
}

func (s Server) DeleteSlice(ctx context.Context, request *rsmapi.DeleteSliceRequest) (*rsmapi.DeleteSliceResponse, error) {
	ackCh := make(chan Ack)
	msg := &RsmMsg{
		NodeID:  topoapi.ID(request.E2NodeId),
		Message: request,
		AckCh:   ackCh,
	}
	go func(msg *RsmMsg) {
		s.rsmReqCh <- msg
	}(msg)

	ack := <-ackCh
	return &rsmapi.DeleteSliceResponse{
		Ack: &rsmapi.Ack{
			Success: ack.Success,
			Cause:   ack.Reason,
		},
	}, nil
}

func (s Server) SetUeSliceAssociation(ctx context.Context, request *rsmapi.SetUeSliceAssociationRequest) (*rsmapi.SetUeSliceAssociationResponse, error) {
	ackCh := make(chan Ack)
	msg := &RsmMsg{
		NodeID:  topoapi.ID(request.E2NodeId),
		Message: request,
		AckCh:   ackCh,
	}
	go func(msg *RsmMsg) {
		s.rsmReqCh <- msg
	}(msg)

	ack := <-ackCh
	return &rsmapi.SetUeSliceAssociationResponse{
		Ack: &rsmapi.Ack{
			Success: ack.Success,
			Cause:   ack.Reason,
		},
	}, nil
}

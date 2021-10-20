// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package utils

import (
	"bytes"
	"context"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/onosproject/onos-api/go/onos/uenib"
	uenib_api "github.com/onosproject/onos-api/go/onos/uenib"
	"github.com/onosproject/onos-lib-go/pkg/southbound"
)

func NewUENibClient(ctx context.Context, certPath string, keyPath string, uenibAddr string) (Client, error) {
	conn, err := southbound.Connect(ctx, uenibAddr, certPath, keyPath)
	if err != nil {
		return Client{}, err
	}

	return Client{
		client: uenib.NewUEServiceClient(conn),
	}, nil
}

type Client struct {
	client uenib.UEServiceClient
}

func AddMockUE() error {
	client, err := NewUENibClient(context.Background(), TLSCrtPath, TLSKeyPath, "onos-uenib:5150")
	if err != nil {
		return err
	}

	bIDList := make([]*uenib_api.BearerId, 0)
	uenibBID := &uenib_api.BearerId{
		BearerId: &uenib_api.BearerId_DrbId{
			DrbId: &uenib_api.DrbId{
				DrbId: &uenib_api.DrbId_FourGdrbId{
					FourGdrbId: &uenib_api.FourGDrbId{
						Value: Ue1DrbID,
						Qci: &uenib_api.Qci{
							Value: Ue1Qci,
						},
					},
				},
			},
		},
	}
	bIDList = append(bIDList, uenibBID)

	ueID := MockUEID

	rsmUEInfo := &uenib_api.RsmUeInfo{
		GlobalUeID: ueID,
		UeIdList: &uenib_api.UeIdentity{
			CuUeF1apID: &uenib_api.CuUeF1ApID{
				Value: CUUEF1apID,
			},
			DuUeF1apID: &uenib_api.DuUeF1ApID{
				Value: DUUEF1apID,
			},
			RANUeNgapID: &uenib_api.RanUeNgapID{
				Value: 0,
			},
			AMFUeNgapID: &uenib_api.AmfUeNgapID{
				Value: 0,
			},
			EnbUeS1apID: &uenib_api.EnbUeS1ApID{
				Value: int32(0),
			},
		},
		BearerIdList: bIDList,
		CellGlobalId: CellGlobalID,
		CuE2NodeId:   MockCUE2NodeID,
		DuE2NodeId:   MockDUE2NodeID,
		SliceList:    make([]*uenib_api.SliceInfo, 0),
	}

	obj := uenib.UE{
		ID:      uenib.ID(ueID),
		Aspects: make(map[string]*types.Any),
	}

	jm := jsonpb.Marshaler{}
	writer := bytes.Buffer{}
	err = jm.Marshal(&writer, rsmUEInfo)
	if err != nil {
		return err
	}

	obj.Aspects[proto.MessageName(rsmUEInfo)] = &types.Any{
		TypeUrl: proto.MessageName(rsmUEInfo),
		Value:   writer.Bytes(),
	}

	req := &uenib.CreateUERequest{
		UE: obj,
	}

	_, err = client.client.CreateUE(context.Background(), req)
	return err
}

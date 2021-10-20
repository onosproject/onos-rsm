// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	rsmapi "github.com/onosproject/onos-api/go/onos/rsm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewConnection() (*grpc.ClientConn, error) {
	address := ":5150"
	certPath := TLSCrtPath
	keyPath := TLSKeyPath
	var opts []grpc.DialOption
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}
	opts = []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		})),
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func CmdCreateSlice(duIndex int, numDUs int, sliceIndex int, numSlices int) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := rsmapi.NewRsmClient(conn)

	for i := duIndex; i <= numDUs; i++ {
		for j := sliceIndex; j <= numSlices; j++ {
			e2NodeID := GetMockDUE2NodeID(i)
			sliceID := GetSliceID(j)
			schedulerType := SliceSched
			weight := GetSliceWeights(j)
			sliceType := SliceType

			setRequest := rsmapi.CreateSliceRequest{
				E2NodeId:      e2NodeID,
				SliceId:       fmt.Sprintf("%d", sliceID),
				Weight:        fmt.Sprintf("%d", weight),
				SchedulerType: schedulerType,
				SliceType:     sliceType,
			}

			resp, err := client.CreateSlice(context.Background(), &setRequest)
			if err != nil {
				return err
			}
			if !resp.GetAck().GetSuccess() {
				return fmt.Errorf(resp.GetAck().GetCause())
			}
		}
	}

	return nil
}

func CmdUpdateSlice(duIndex int, numDUs int, sliceIndex int, numSlices int) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := rsmapi.NewRsmClient(conn)

	for i := duIndex; i <= numDUs; i++ {
		for j := sliceIndex; j <= numSlices; j++ {
			e2NodeID := GetMockDUE2NodeID(i)
			sliceID := GetSliceID(j)
			schedulerType := SliceSched
			weight := GetSliceUpdatedWeights(j)
			sliceType := SliceType

			setRequest := rsmapi.UpdateSliceRequest{
				E2NodeId:      e2NodeID,
				SliceId:       fmt.Sprintf("%d", sliceID),
				Weight:        fmt.Sprintf("%d", weight),
				SchedulerType: schedulerType,
				SliceType:     sliceType,
			}

			resp, err := client.UpdateSlice(context.Background(), &setRequest)
			if err != nil {
				return err
			}
			if !resp.GetAck().GetSuccess() {
				return fmt.Errorf(resp.GetAck().GetCause())
			}
		}
	}
	return nil
}

func CmdAssociateUEWithSlice(duIndex int, numDUs int, sliceIndex int, numSlices int, numUEs int) error {
	// precheck
	if numDUs*numSlices > numUEs {
		return fmt.Errorf("the number of DUs times the number of slices should be more than the number of UEs")
	}

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := rsmapi.NewRsmClient(conn)

	for i := duIndex; i <= numDUs; i++ {
		for j := sliceIndex; j <= numSlices; j++ {
			tmpUEIndex := (i-1)*numSlices + j
			e2NodeID := GetMockDUE2NodeID(i)
			sliceID := GetSliceID(j)
			drbID := GetDrbID(tmpUEIndex)
			duUEID := GetDUUEF1apID(tmpUEIndex)
			idList := make([]*rsmapi.UeId, 0)
			duUeF1apIDField := &rsmapi.UeId{
				UeId: fmt.Sprintf("%d", duUEID),
				Type: rsmapi.UeIdType_UE_ID_TYPE_DU_UE_F1_AP_ID,
			}
			idList = append(idList, duUeF1apIDField)

			setRequest := rsmapi.SetUeSliceAssociationRequest{
				E2NodeId:  e2NodeID,
				UeId:      idList,
				DlSliceId: fmt.Sprintf("%d", sliceID),
				DrbId:     fmt.Sprintf("%d", drbID),
			}
			resp, err := client.SetUeSliceAssociation(context.Background(), &setRequest)
			if err != nil {
				return err
			}
			if !resp.GetAck().GetSuccess() {
				return fmt.Errorf(resp.GetAck().GetCause())
			}
		}
	}
	return nil
}

func CmdDeleteSlice(duIndex int, numDUs int, sliceIndex int, numSlices int) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := rsmapi.NewRsmClient(conn)

	for i := duIndex; i <= numDUs; i++ {
		for j := sliceIndex; j <= numSlices; j++ {
			e2NodeID := GetMockDUE2NodeID(i)
			sliceID := GetSliceID(j)
			sliceType := SliceType

			setRequest := rsmapi.DeleteSliceRequest{
				E2NodeId:  e2NodeID,
				SliceId:   fmt.Sprintf("%d", sliceID),
				SliceType: sliceType,
			}
			resp, err := client.DeleteSlice(context.Background(), &setRequest)
			if err != nil {
				return err
			}
			if !resp.GetAck().GetSuccess() {
				return fmt.Errorf(resp.GetAck().GetCause())
			}
		}
	}
	return nil
}

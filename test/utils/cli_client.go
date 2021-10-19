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
	certPath := TlsCrtPath
	keyPath := TlsKeyPath
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

func CmdCreateSlice1() error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := rsmapi.NewRsmClient(conn)

	e2NodeID := MockDUE2NodeID
	sliceID := Slice1ID
	schedulerType := Slice1Sched
	weight := Slice1Weight1
	sliceType := Slice1Type

	setRequest := rsmapi.CreateSliceRequest{
		E2NodeId:      e2NodeID,
		SliceId:       sliceID,
		Weight:        weight,
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
	return nil
}

func CmdUpdateSlice1() error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := rsmapi.NewRsmClient(conn)

	e2NodeID := MockDUE2NodeID
	sliceID := Slice1ID
	schedulerType := Slice1Sched
	weight := Slice1Weight2
	sliceType := Slice1Type

	setRequest := rsmapi.UpdateSliceRequest{
		E2NodeId: e2NodeID,
		SliceId: sliceID,
		Weight: weight,
		SchedulerType: schedulerType,
		SliceType: sliceType,
	}

	resp, err := client.UpdateSlice(context.Background(), &setRequest)
	if err != nil {
		return err
	}
	if !resp.GetAck().GetSuccess() {
		return fmt.Errorf(resp.GetAck().GetCause())
	}
	return nil
}

func CmdAssociateUE1WithSlice1() error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := rsmapi.NewRsmClient(conn)

	e2NodeID := MockDUE2NodeID
	sliceID := Slice1ID
	drbID := Ue1DrbID
	duUEID := DUUEF1apID
	idList := make([]*rsmapi.UeId, 0)
	duUeF1apIDField := &rsmapi.UeId{
		UeId: fmt.Sprintf("%d", duUEID),
		Type: rsmapi.UeIdType_UE_ID_TYPE_DU_UE_F1_AP_ID,
	}
	idList = append(idList, duUeF1apIDField)

	setRequest := rsmapi.SetUeSliceAssociationRequest{
		E2NodeId: e2NodeID,
		UeId: idList,
		DlSliceId: sliceID,
		DrbId: fmt.Sprintf("%d", drbID),
	}

	resp, err := client.SetUeSliceAssociation(context.Background(), &setRequest)
	if err != nil {
		return err
	}
	if !resp.GetAck().GetSuccess() {
		return fmt.Errorf(resp.GetAck().GetCause())
	}
	return nil
}

func CmdDeleteSlice1() error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := rsmapi.NewRsmClient(conn)

	e2NodeID := MockDUE2NodeID
	sliceID := Slice1ID
	sliceType := Slice1Type

	setRequest := rsmapi.DeleteSliceRequest{
		E2NodeId: e2NodeID,
		SliceId: sliceID,
		SliceType: sliceType,
	}

	resp, err := client.DeleteSlice(context.Background(), &setRequest)
	if err != nil {
		return err
	}
	if !resp.GetAck().GetSuccess() {
		return fmt.Errorf(resp.GetAck().GetCause())
	}
	return nil
}
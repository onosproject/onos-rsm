// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package utils

import (
	"context"
	"fmt"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	uenib_api "github.com/onosproject/onos-api/go/onos/uenib"
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	"github.com/onosproject/onos-rsm/pkg/nib/uenib"
	"strconv"
)

func VerifyCase1CreatingSlice() error {
	rnibClient, err := rnib.NewClient()
	if err != nil {
		return err
	}

	items, err := rnibClient.GetRsmSliceItemAspects(context.Background(), MockDUE2NodeID)
	if err != nil {
		return err
	}

	if len(items) < 1 {
		return fmt.Errorf("there is no slice created - there should be one slice")
	} else if len(items) > 1 {
		return fmt.Errorf("there are multiple slices created - there should be one slice")
	}

	if items[0].GetID() != Slice1ID {
		return fmt.Errorf("slice ID %s is wrong - it should be %s", items[0].GetID(), Slice1ID)
	}

	if len(items[0].GetUeIdList()) != 0 {
		return fmt.Errorf("UeIdList should be empty - UeIdList: %v", items[0].GetUeIdList())
	}

	if items[0].GetSliceType() != topoapi.RSMSliceType(Slice1Type) {
		return fmt.Errorf("slice type %v is wrong - it should be %v", items[0].GetSliceType(), topoapi.RSMSliceType(Slice1Type))
	}

	slice1Weight1Int32, err := strconv.Atoi(Slice1Weight1)
	if err != nil {
		return err
	}
	if items[0].GetSliceParameters().GetWeight() != int32(slice1Weight1Int32) {
		return fmt.Errorf("weight %v is wrong - it should be %v", items[0].GetSliceParameters().GetWeight(), int32(slice1Weight1Int32))
	}

	if items[0].GetSliceParameters().GetSchedulerType() != topoapi.RSMSchedulerType(Slice1Sched) {
		return fmt.Errorf("scheduler type %v is wrong - it should be %v", items[0].GetSliceParameters().GetSchedulerType(), topoapi.RSMSchedulerType(Slice1Sched))
	}

	return nil
}

func VerifyCase2UpdatingSlice() error {
	rnibClient, err := rnib.NewClient()
	if err != nil {
		return err
	}

	items, err := rnibClient.GetRsmSliceItemAspects(context.Background(), MockDUE2NodeID)
	if err != nil {
		return err
	}

	if len(items) < 1 {
		return fmt.Errorf("there is no slice created - there should be one slice")
	} else if len(items) > 1 {
		return fmt.Errorf("there are multiple slices created - there should be one slice")
	}

	if items[0].GetID() != Slice1ID {
		return fmt.Errorf("slice ID %s is wrong - it should be %s", items[0].GetID(), Slice1ID)
	}

	if len(items[0].GetUeIdList()) != 0 {
		return fmt.Errorf("UeIdList should be empty - UeIdList: %v", items[0].GetUeIdList())
	}

	if items[0].GetSliceType() != topoapi.RSMSliceType(Slice1Type) {
		return fmt.Errorf("slice type %v is wrong - it should be %v", items[0].GetSliceType(), topoapi.RSMSliceType(Slice1Type))
	}

	slice1Weight1Int32, err := strconv.Atoi(Slice1Weight2)
	if err != nil {
		return err
	}
	if items[0].GetSliceParameters().GetWeight() != int32(slice1Weight1Int32) {
		return fmt.Errorf("weight %v is wrong - it should be %v", items[0].GetSliceParameters().GetWeight(), int32(slice1Weight1Int32))
	}

	if items[0].GetSliceParameters().GetSchedulerType() != topoapi.RSMSchedulerType(Slice1Sched) {
		return fmt.Errorf("scheduler type %v is wrong - it should be %v", items[0].GetSliceParameters().GetSchedulerType(), topoapi.RSMSchedulerType(Slice1Sched))
	}

	return nil
}

func VerifyCase3AssociatingUEWithSlice() error {
	rnibClient, err := rnib.NewClient()
	if err != nil {
		return err
	}
	uenibClient, err := uenib.NewClient(context.Background(), TLSCrtPath, TLSKeyPath, "onos-uenib:5150")
	if err != nil {
		return err
	}

	items, err := rnibClient.GetRsmSliceItemAspects(context.Background(), MockDUE2NodeID)
	if err != nil {
		return err
	}

	if len(items) < 1 {
		return fmt.Errorf("there is no slice created - there should be one slice")
	} else if len(items) > 1 {
		return fmt.Errorf("there are multiple slices created - there should be one slice")
	}

	if items[0].GetID() != Slice1ID {
		return fmt.Errorf("slice ID %s is wrong - it should be %s", items[0].GetID(), Slice1ID)
	}

	if items[0].GetSliceType() != topoapi.RSMSliceType(Slice1Type) {
		return fmt.Errorf("slice type %v is wrong - it should be %v", items[0].GetSliceType(), topoapi.RSMSliceType(Slice1Type))
	}

	slice1Weight1Int32, err := strconv.Atoi(Slice1Weight2)
	if err != nil {
		return err
	}
	if items[0].GetSliceParameters().GetWeight() != int32(slice1Weight1Int32) {
		return fmt.Errorf("weight %v is wrong - it should be %v", items[0].GetSliceParameters().GetWeight(), int32(slice1Weight1Int32))
	}

	if items[0].GetSliceParameters().GetSchedulerType() != topoapi.RSMSchedulerType(Slice1Sched) {
		return fmt.Errorf("scheduler type %v is wrong - it should be %v", items[0].GetSliceParameters().GetSchedulerType(), topoapi.RSMSchedulerType(Slice1Sched))
	}

	if len(items[0].GetUeIdList()) != 1 {
		return fmt.Errorf("UeIdList should not be empty - UeIdList: %v", items[0].GetUeIdList())
	}

	rsmUEInfo, err := uenibClient.GetUEWithGlobalID(context.Background(), MockUEID)
	if err != nil {
		return err
	}

	if rsmUEInfo.DuE2NodeId != MockDUE2NodeID {
		return fmt.Errorf("DuE2NodeID %v is wrong - it should be %v", rsmUEInfo.DuE2NodeId, MockDUE2NodeID)
	}

	if rsmUEInfo.UeIdList.DuUeF1apID.Value != DUUEF1apID {
		return fmt.Errorf("DuUeF1apID %v is wrong - it should be %v", rsmUEInfo.UeIdList.DuUeF1apID.Value, DUUEF1apID)
	}

	if len(rsmUEInfo.BearerIdList) != 1 {
		return fmt.Errorf("BearerIdList should not be empty - BearerIdList: %v", rsmUEInfo.BearerIdList)
	}

	if rsmUEInfo.BearerIdList[0].GetDrbId().GetFourGdrbId().GetValue() != Ue1DrbID {
		return fmt.Errorf("DrbID %v is wrong - it should be %v", rsmUEInfo.BearerIdList[0].GetDrbId().GetFourGdrbId().GetValue(), Ue1DrbID)
	}

	if rsmUEInfo.BearerIdList[0].GetDrbId().GetFourGdrbId().GetQci().GetValue() != Ue1Qci {
		return fmt.Errorf("QCI %v is wrong - it should be %v", rsmUEInfo.BearerIdList[0].GetDrbId().GetFourGdrbId().GetQci().GetValue(), Ue1Qci)
	}

	if len(rsmUEInfo.GetSliceList()) != 1 {
		return fmt.Errorf("SliceList should not be empty - SlistList: %v", rsmUEInfo.GetSliceList())
	}

	if rsmUEInfo.GetSliceList()[0].GetID() != Slice1ID {
		return fmt.Errorf("slice ID %v is wrong - it should be %v", rsmUEInfo.GetSliceList()[0].GetID(), Slice1ID)
	}

	if rsmUEInfo.GetSliceList()[0].GetDrbId().GetFourGdrbId().GetValue() != Ue1DrbID {
		return fmt.Errorf("DrbID in slice list %v is wrong - it should be %v", rsmUEInfo.GetSliceList()[0].GetDrbId().GetFourGdrbId().GetValue(), Ue1DrbID)
	}

	if rsmUEInfo.GetSliceList()[0].GetDrbId().GetFourGdrbId().GetQci().GetValue() != Ue1Qci {
		return fmt.Errorf("QCI in slice list %v is wrong - it should be %v", rsmUEInfo.GetSliceList()[0].GetDrbId().GetFourGdrbId().GetQci().GetValue(), Ue1Qci)
	}

	if rsmUEInfo.GetSliceList()[0].GetSliceParameters().GetWeight() != int32(slice1Weight1Int32) {
		return fmt.Errorf("weight in slice list %v is wrong - it should be %v", rsmUEInfo.GetSliceList()[0].GetSliceParameters().GetWeight(), int32(slice1Weight1Int32))
	}

	if rsmUEInfo.GetSliceList()[0].GetSliceParameters().GetSchedulerType() != uenib_api.RSMSchedulerType(Slice1Sched) {
		return fmt.Errorf("scheduler type in slice list %v is wrong - it should be %v", items[0].GetSliceParameters().GetSchedulerType(), topoapi.RSMSchedulerType(Slice1Sched))
	}

	return nil
}

func VerifyCase4DeletingSlice() error {
	rnibClient, err := rnib.NewClient()
	if err != nil {
		return err
	}
	uenibClient, err := uenib.NewClient(context.Background(), TLSCrtPath, TLSKeyPath, "onos-uenib:5150")
	if err != nil {
		return err
	}

	items, err := rnibClient.GetRsmSliceItemAspects(context.Background(), MockDUE2NodeID)
	if err != nil {
		return err
	}

	if len(items) != 0 {
		return fmt.Errorf("slice should be empty - %v", items)
	}

	rsmUEInfo, err := uenibClient.GetUEWithGlobalID(context.Background(), MockUEID)
	if err != nil {
		return err
	}

	if rsmUEInfo.DuE2NodeId != MockDUE2NodeID {
		return fmt.Errorf("DuE2NodeID %v is wrong - it should be %v", rsmUEInfo.DuE2NodeId, MockDUE2NodeID)
	}

	if rsmUEInfo.UeIdList.DuUeF1apID.Value != DUUEF1apID {
		return fmt.Errorf("DuUeF1apID %v is wrong - it should be %v", rsmUEInfo.UeIdList.DuUeF1apID.Value, DUUEF1apID)
	}

	if len(rsmUEInfo.BearerIdList) != 1 {
		return fmt.Errorf("BearerIdList should not be empty - BearerIdList: %v", rsmUEInfo.BearerIdList)
	}

	if rsmUEInfo.BearerIdList[0].GetDrbId().GetFourGdrbId().GetValue() != Ue1DrbID {
		return fmt.Errorf("DrbID %v is wrong - it should be %v", rsmUEInfo.BearerIdList[0].GetDrbId().GetFourGdrbId().GetValue(), Ue1DrbID)
	}

	if rsmUEInfo.BearerIdList[0].GetDrbId().GetFourGdrbId().GetQci().GetValue() != Ue1Qci {
		return fmt.Errorf("QCI %v is wrong - it should be %v", rsmUEInfo.BearerIdList[0].GetDrbId().GetFourGdrbId().GetQci().GetValue(), Ue1Qci)
	}

	if len(rsmUEInfo.GetSliceList()) != 0 {
		return fmt.Errorf("SliceList should be empty - SliceList: %v", rsmUEInfo.GetSliceList())
	}

	return nil
}

// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package utils

import (
	"context"
	"fmt"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	"github.com/onosproject/onos-rsm/pkg/nib/uenib"
	"strconv"
	"strings"
)

func VerifySliceInitValuesForAllDUs(numSlices int) error {
	rnibClient, err := rnib.NewClient()
	if err != nil {
		return err
	}

	rsmAspects, err := rnibClient.GetRSMSliceItemAspectsForAllDUs(context.Background())
	if err != nil {
		return err
	}

	for k, v := range rsmAspects {
		if len(v) != numSlices {
			return fmt.Errorf("the number of created slices should be %d by default - currently %v has %v [%v]", numSlices, k, len(v), v)
		}

		for i := 0; i < numSlices; i++ {
			if v[i].GetID() != fmt.Sprintf("%d", GetSliceID(i+1)) {
				return fmt.Errorf("slice ID %v in %v is wrong - it should be %v", v[i].GetID(), k, fmt.Sprintf("%d", GetSliceID(i+1)))
			}

			if len(v[i].GetUeIdList()) != 0 {
				return fmt.Errorf("UeIdList in %v should be empty - UeIdList: %v", k, v[0].GetUeIdList())
			}

			if v[i].GetSliceType() != topoapi.RSMSliceType(SliceType) {
				return fmt.Errorf("slice type %v in %v is wrong - it should be %v", v[0].GetSliceType(), k, topoapi.RSMSliceType(SliceType))
			}

			if v[i].GetSliceParameters().GetWeight() != int32(GetSliceWeights(i+1)) {
				return fmt.Errorf("weight %v in %v is wrong - it should be %v", v[0].GetSliceParameters().GetWeight(), k, int32(GetSliceWeights(i+1)))
			}

			if v[i].GetSliceParameters().GetSchedulerType() != topoapi.RSMSchedulerType(SliceSched) {
				return fmt.Errorf("scheduler type %v in %v is wrong - it should be %v", v[0].GetSliceParameters().GetSchedulerType(), k, topoapi.RSMSchedulerType(SliceSched))
			}
		}
	}

	return nil
}

func VerifySliceUpdatedValuesForAllDUs(numSlices int) error {
	rnibClient, err := rnib.NewClient()
	if err != nil {
		return err
	}

	rsmAspects, err := rnibClient.GetRSMSliceItemAspectsForAllDUs(context.Background())
	if err != nil {
		return err
	}

	for k, v := range rsmAspects {
		if len(v) != numSlices {
			return fmt.Errorf("the number of created slices should be %d by default - currently %v has %v [%v]", numSlices, k, len(v), v)
		}

		for i := 0; i < numSlices; i++ {
			if v[i].GetID() != fmt.Sprintf("%d", GetSliceID(i+1)) {
				return fmt.Errorf("slice ID %v in %v is wrong - it should be %v", v[i].GetID(), k, fmt.Sprintf("%d", GetSliceID(i+1)))
			}

			if len(v[i].GetUeIdList()) != 0 {
				return fmt.Errorf("UeIdList in %v should be empty - UeIdList: %v", k, v[0].GetUeIdList())
			}

			if v[i].GetSliceType() != topoapi.RSMSliceType(SliceType) {
				return fmt.Errorf("slice type %v in %v is wrong - it should be %v", v[0].GetSliceType(), k, topoapi.RSMSliceType(SliceType))
			}

			if v[i].GetSliceParameters().GetWeight() != int32(GetSliceUpdatedWeights(i+1)) {
				return fmt.Errorf("weight %v in %v is wrong - it should be %v", v[0].GetSliceParameters().GetWeight(), k, int32(GetSliceUpdatedWeights(i+1)))
			}

			if v[i].GetSliceParameters().GetSchedulerType() != topoapi.RSMSchedulerType(SliceSched) {
				return fmt.Errorf("scheduler type %v in %v is wrong - it should be %v", v[0].GetSliceParameters().GetSchedulerType(), k, topoapi.RSMSchedulerType(SliceSched))
			}
		}
	}

	return nil
}

func VerifyUESliceAssociationForAllDUsAndUEs(numSlices int) error {
	rnibClient, err := rnib.NewClient()
	if err != nil {
		return err
	}
	uenibClient, err := uenib.NewClient(context.Background(), TLSCrtPath, TLSKeyPath, "onos-uenib:5150")
	if err != nil {
		return err
	}

	rsmAspects, err := rnibClient.GetRSMSliceItemAspectsForAllDUs(context.Background())
	if err != nil {
		return err
	}

	for k, v := range rsmAspects {
		if len(v) != numSlices {
			return fmt.Errorf("the number of created slices should be three by default - currently %v has %v [%v]", k, len(v), v)
		}
		for i := 0; i < numSlices; i++ {
			if v[i].GetID() != fmt.Sprintf("%d", GetSliceID(i+1)) {
				return fmt.Errorf("slice ID %v in %v is wrong - it should be %v", v[i].GetID(), k, fmt.Sprintf("%d", GetSliceID(i+1)))
			}

			if v[i].GetSliceType() != topoapi.RSMSliceType(SliceType) {
				return fmt.Errorf("slice type %v in %v is wrong - it should be %v", v[0].GetSliceType(), k, topoapi.RSMSliceType(SliceType))
			}

			if v[i].GetSliceParameters().GetWeight() != int32(GetSliceUpdatedWeights(i+1)) {
				return fmt.Errorf("weight %v in %v is wrong - it should be %v", v[0].GetSliceParameters().GetWeight(), k, int32(GetSliceUpdatedWeights(i+1)))
			}

			if v[i].GetSliceParameters().GetSchedulerType() != topoapi.RSMSchedulerType(SliceSched) {
				return fmt.Errorf("scheduler type %v in %v is wrong - it should be %v", v[0].GetSliceParameters().GetSchedulerType(), k, topoapi.RSMSchedulerType(SliceSched))
			}
		}
	}

	rsmUEInfoAspects, err := uenibClient.GetUEs(context.Background())
	if err != nil {
		return err
	}

	for _, rsmUEInfoAspect := range rsmUEInfoAspects {
		if _, ok := rsmAspects[rsmUEInfoAspect.GetDuE2NodeId()]; !ok {
			return fmt.Errorf("UE %v does not have DuE2NodeID", rsmUEInfoAspect.GetGlobalUeID())
		}

		parsedUEIDIndex, err := strconv.Atoi(strings.Split(rsmUEInfoAspect.GetGlobalUeID(), "-")[4])
		if err != nil {
			return err
		}

		if rsmUEInfoAspect.GetUeIdList().GetCuUeF1apID().GetValue() != int64(GetCUUEF1apID(parsedUEIDIndex)) {
			return fmt.Errorf("CU UE F1AP ID %v in UENIB %v is wrong. it should be %v", rsmUEInfoAspect.GetGlobalUeID(), rsmUEInfoAspect, GetCUUEF1apID(parsedUEIDIndex))
		}

		if rsmUEInfoAspect.GetUeIdList().GetDuUeF1apID().GetValue() != int64(GetDUUEF1apID(parsedUEIDIndex)) {
			return fmt.Errorf("DU UE F1AP ID %v in UENIB %v is wrong. it should be %v", rsmUEInfoAspect.GetGlobalUeID(), rsmUEInfoAspect, GetCUUEF1apID(parsedUEIDIndex))
		}

		for _, ueSlice := range rsmUEInfoAspect.GetSliceList() {
			// slice ID check
			hasSliceID := false
			uenibDrbID := ueSlice.GetDrbId().GetFourGdrbId().GetValue()
			uenibQci := ueSlice.GetDrbId().GetFourGdrbId().GetQci().GetValue()
			uenibWeight := ueSlice.GetSliceParameters().GetWeight()
			uenibSched := ueSlice.GetSliceParameters().GetSchedulerType()

			for _, topoSlice := range rsmAspects[rsmUEInfoAspect.GetDuE2NodeId()] {
				if ueSlice.GetID() == topoSlice.GetID() {
					hasUEID := false
					var topoDrbID int32
					var topoQci int32
					var topoWeight int32
					var topoSched topoapi.RSMSchedulerType
					for _, ue := range topoSlice.GetUeIdList() {
						if ue.GetDuUeF1apID().GetValue() == rsmUEInfoAspect.GetUeIdList().GetDuUeF1apID().GetValue() {
							hasUEID = true
							topoDrbID = ue.GetDrbId().GetFourGdrbId().GetValue()
							topoQci = ue.GetDrbId().GetFourGdrbId().GetQci().GetValue()
							topoWeight = topoSlice.GetSliceParameters().GetWeight()
							topoSched = topoSlice.GetSliceParameters().GetSchedulerType()
							break
						}
					}
					if !hasUEID {
						return fmt.Errorf("CU/DUUEID in topo %v is not matched with the CU/DUUEID in UENIB %v", topoSlice.GetUeIdList(), rsmUEInfoAspect.GetUeIdList().GetCuUeF1apID().GetValue())
					}

					// bearer ID check
					if uenibDrbID != topoDrbID {
						return fmt.Errorf("DRB-ID in topo %v and DRB-ID in uenib %v are not matched", topoDrbID, uenibDrbID)
					}

					if uenibQci != topoQci {
						return fmt.Errorf("QCI in topo %v and QCI in uenib %v are not matched", topoDrbID, uenibDrbID)
					}

					// scheduler check
					if uenibWeight != topoWeight {
						return fmt.Errorf("weight in topo %v and weight in uenib %v are not matched", topoWeight, uenibWeight)
					}

					if uenibSched.String() != topoSched.String() {
						return fmt.Errorf("scheduler type in topo %v and scheduler type in uenib %v are not matched", topoSched, uenibSched)
					}

					hasSliceID = true
					break
				}
			}

			if !hasSliceID {
				return fmt.Errorf("slice ID in UENIB %v is not matched with the slice ID in topo %v", ueSlice.GetID(), rsmAspects[rsmUEInfoAspect.GetDuE2NodeId()])
			}
		}
	}

	return nil
}

func VerifySliceDeletedForAllDUsAfterUEAssociation() error {
	rnibClient, err := rnib.NewClient()
	if err != nil {
		return err
	}
	uenibClient, err := uenib.NewClient(context.Background(), TLSCrtPath, TLSKeyPath, "onos-uenib:5150")
	if err != nil {
		return err
	}

	rsmAspects, err := rnibClient.GetRSMSliceItemAspectsForAllDUs(context.Background())
	if err != nil {
		return err
	}

	for k, v := range rsmAspects {
		if len(v) != 0 {
			return fmt.Errorf("DU %v slice list should be empty - %v", k, v)
		}
	}

	rsmUEInfoAspects, err := uenibClient.GetUEs(context.Background())
	if err != nil {
		return err
	}

	for _, rsmUEInfoAspect := range rsmUEInfoAspects {
		if len(rsmUEInfoAspect.GetSliceList()) != 0 {
			return fmt.Errorf("SliceList for UE %v should be empty - SliceList: %v", rsmUEInfoAspect.GetGlobalUeID(), rsmUEInfoAspect.GetSliceList())
		}
	}

	return nil
}

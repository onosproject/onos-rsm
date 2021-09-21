// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package utils

import (
	"fmt"
	e2sm_rsm_ies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-rsm-ies"
	e2sm_v2_ies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-v2-ies"
	"github.com/onosproject/onos-rsm/pkg/types"
)

func CreateCgiKey(cgi *e2sm_v2_ies.Cgi) (string, error) {
	if cgi.GetNRCgi() != nil {
		return fmt.Sprintf("nrCGI-%s-%s", cgi.GetNRCgi().GetPLmnidentity(), cgi.GetNRCgi().GetNRcellIdentity()), nil
	} else if cgi.GetEUtraCgi() != nil {
		return fmt.Sprintf("eutraCGI-%s-%s", cgi.GetEUtraCgi().GetPLmnidentity(), cgi.GetEUtraCgi().GetEUtracellIdentity()), nil
	}
	return "", fmt.Errorf("CGI field is empty")
}

func CreateUEKey(ueIDs []*e2sm_rsm_ies.UeIdentity) (string, error) {
	var cuUeF1apID int64
	var duUeF1apID int64
	var ranUeNgapID int64
	var enbUeS1apID int32
	ueType := types.None

	for _, id := range ueIDs {
		if id.GetCuUeF1ApId() != nil {
			cuUeF1apID = id.GetCuUeF1ApId().GetValue()
		}
		if id.GetDuUeF1ApId() != nil {
			duUeF1apID = id.GetDuUeF1ApId().GetValue()
		}
		if id.GetEnbUeS1ApId() != nil {
			enbUeS1apID = id.GetEnbUeS1ApId().GetValue()
			ueType = types.UE
		}
		if id.GetRanUeNgapId() != nil {
			ranUeNgapID = id.GetRanUeNgapId().GetValue()
			ueType = types.NrUE
		}
	}

	switch ueType {
	case types.UE:
		return fmt.Sprintf("%s-%d-%d-%d", ueType.String(), cuUeF1apID, duUeF1apID, enbUeS1apID), nil
	case types.NrUE:
		return fmt.Sprintf("%s-%d-%d-%d", ueType.String(), cuUeF1apID, duUeF1apID, ranUeNgapID), nil
	default:
		return "", fmt.Errorf("UE-ID has to have either UES1AP ID or UENGAP ID")
	}
}
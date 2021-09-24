// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package types

import e2sm_rsm_ies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-rsm-ies"

type RsmUE struct {
	RsmUEID RsmUEID
	BearerIDs []e2sm_rsm_ies.BearerId
	PreferredUeIDType e2sm_rsm_ies.UeIdType
}

type RsmUEID struct {
	CuUeF1ApId e2sm_rsm_ies.CuUeF1ApId
	DuUeF1ApId e2sm_rsm_ies.DuUeF1ApId
	RanUeNgapId e2sm_rsm_ies.RanUeNgapId
	EnbUeS1apId e2sm_rsm_ies.EnbUeS1ApId
}
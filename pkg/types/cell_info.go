// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package types

import (
	e2sm_rsm_ies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-rsm-ies"
	e2sm_v2_ies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-v2-ies"
)

type CellID struct {
	CuNodeID string
	DuNodeID string
	NodeID string
	Cgi e2sm_v2_ies.Cgi
}

type CellInfo struct {
	CellID CellID
	MaxNumSlicesDl int32
	MaxNumSlicesUl int32
	SlicingType e2sm_rsm_ies.SlicingType
	MaxNumUEsPerSlice int32
	SupportedSlices []e2sm_rsm_ies.SupportedSlicingConfigItem
}
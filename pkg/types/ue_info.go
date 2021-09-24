// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package types

import e2sm_rsm_ies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-rsm-ies"

// UeInfo has a UE information
type UeInfo struct {
	UeIDs      []e2sm_rsm_ies.UeIdentity
	BearerIDs  []e2sm_rsm_ies.BearerId
	ServCellID CellID
}

// UEType is a type of UE
type UEType int

const (
	None UEType = iota
	UE
	NrUE
)

// String returns enum as in string format
func (u UEType) String() string {
	return [...]string{"None", "UE", "NrUE"}[u]
}

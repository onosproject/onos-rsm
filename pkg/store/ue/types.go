// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package uestore

import (
	e2sm_rsm_ies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-rsm-ies"
	e2sm_v2_ies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-v2-ies"
)

type Entry struct {
	Key   string
	Value interface{}
}

type EventType int

const (
	None EventType = iota
	Created
	UpdatedUEInfo
	UpdatedSlice
	Deleted
)

func (e EventType) String() string {
	return [...]string{"None", "Created", "UpdatedUEInfo", "UpdatedSlice", "Deleted"}[e]
}

// RsmUE has UE information
type RsmUE struct {
	RsmUEID    RsmUEID
	BearerIDs  []e2sm_rsm_ies.BearerId
	Cgi        e2sm_v2_ies.Cgi
	CuID       int64
	DuID       int64
	DlSliceIDs map[int32]string
	UlSliceIDs map[int32]string
}

// RsmUEID has a set of UE IDs
type RsmUEID struct {
	CuUeF1ApID        e2sm_rsm_ies.CuUeF1ApId
	DuUeF1ApID        e2sm_rsm_ies.DuUeF1ApId
	RanUeNgapID       e2sm_rsm_ies.RanUeNgapId
	EnbUeS1apID       e2sm_rsm_ies.EnbUeS1ApId
	PreferredUeIDType e2sm_rsm_ies.UeIdType
}

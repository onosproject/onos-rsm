// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package types

import (
	"github.com/google/uuid"
	e2sm_rsm_ies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-rsm-ies"
)

type SliceUEAssoc struct {
	SliceUEAssocID uuid.UUID
	RsmUE RsmUE
	RsmE2Node RsmE2Node
	Slice Slice
}

type Slice struct {
	SliceID int32
	SliceDesc string
	SliceConfigParameters e2sm_rsm_ies.SliceParameters
}


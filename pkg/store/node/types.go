// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package nodestore

import "github.com/onosproject/onos-api/go/onos/topo"

type Entry struct {
	Key   string
	Value interface{}
}
type EventType int

const (
	None EventType = iota
	Created
	UpdatedNode
	UpdatedSliceCreated
	UpdatedSliceUpdated
	UpdatedSliceDeleted
	Deleted
)

func (e EventType) String() string {
	return [...]string{"None", "Created", "UpdatedNode", "UpdatedSliceCreated", "UpdatedSliceUpdated", "UpdatedSliceDeleted", "Deleted"}[e]
}

// RsmE2Node has E2 node information
type RsmE2Node struct {
	RsmE2NodeID       string
	RsmNodeCapability []topo.RSMNodeSlicingCapabilityItem
	RsmSliceList      topo.RSMSliceItemList
}

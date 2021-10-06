// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package northbound

import topoapi "github.com/onosproject/onos-api/go/onos/topo"

type Ack struct {
	Success bool
	Reason  string
}

type RsmMsg struct {
	NodeID  topoapi.ID
	Message interface{}
	AckCh   chan Ack
}

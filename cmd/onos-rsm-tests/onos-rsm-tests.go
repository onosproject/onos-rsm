// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/onosproject/helmit/pkg/registry"
	"github.com/onosproject/helmit/pkg/test"
	"github.com/onosproject/onos-rsm/test/scalability"
	"github.com/onosproject/onos-rsm/test/slice"
)

func main() {
	registry.RegisterTestSuite("slice", &slice.TestSuite{})
	registry.RegisterTestSuite("scalability", &scalability.TestSuite{})
	test.Main()
}

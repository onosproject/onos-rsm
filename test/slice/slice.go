// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package slice

import (
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-rsm/pkg/manager"
	"github.com/onosproject/onos-rsm/test/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func (s *TestSuite) TestSlice(t *testing.T) {

	cfg := manager.Config{
		CAPath:      "/tmp/tls.cacrt",
		KeyPath:     "/tmp/tls.key",
		CertPath:    "/tmp/tls.crt",
		ConfigPath:  "/tmp/config.json",
		E2tEndpoint: "onos-e2t:5150",
		GRPCPort:    5150,
		SMName:      "oran-e2sm-rsm",
		SMVersion:   "v1",
		UenibHost:   "onos-uenib:5150",
		AppID:       "onos-rsm",
		AckTimer:    -1, // timer -1 is for integration test or uenib/topo debugging
	}

	_, err := certs.HandleCertPaths(cfg.CAPath, cfg.KeyPath, cfg.CertPath, true)
	assert.NoError(t, err)

	mgr := manager.NewManager(cfg)
	mgr.Run()

	t.Log("Adding Mock CU")

	err = utils.AddMockCUE2Node()
	assert.NoError(t, err)

	t.Log("Adding Mock DU")

	err = utils.AddMockDUE2Node()
	assert.NoError(t, err)

	t.Log("Adding Mock UE")

	err = utils.AddMockUE()
	assert.NoError(t, err)

	t.Log("Case 1: Creating Slice 1")

	err = utils.CmdCreateSlice1()
	assert.NoError(t, err)

	err = utils.VerifyCase1CreatingSlice()
	assert.NoError(t, err)

	t.Log("Case 1 passed")

	t.Log("Case 2: Updating Slice 1")
	err = utils.CmdUpdateSlice1()
	assert.NoError(t, err)

	err = utils.VerifyCase2UpdatingSlice()
	assert.NoError(t, err)

	t.Log("Case 2 passed")

	t.Log("Case 3: Associating UE 1 with Slice 1")
	err = utils.CmdAssociateUE1WithSlice1()
	assert.NoError(t, err)

	err = utils.VerifyCase3AssociatingUEWithSlice()
	assert.NoError(t, err)

	t.Log("Case 3 passed")

	t.Log("Case 4: Deleting Slice 1")
	err = utils.CmdDeleteSlice1()
	assert.NoError(t, err)

	err = utils.VerifyCase4DeletingSlice()
	assert.NoError(t, err)

	t.Log("Case 4 passed")
}

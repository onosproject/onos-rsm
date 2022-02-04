// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package slice

import (
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-rsm/pkg/manager"
	"github.com/onosproject/onos-rsm/test/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	numTestSlices = 1
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

	t.Log("Adding Mock 100 CUs")

	err = utils.CreateMockE2Node(1, utils.CUDUTypeCU)
	assert.NoError(t, err)

	t.Log("Adding Mock DU")

	err = utils.CreateMockE2Node(1, utils.CUDUTypeDU)
	assert.NoError(t, err)

	t.Log("Adding Mock UE")

	err = utils.CreateMockUE(1, 1, 1)
	assert.NoError(t, err)

	t.Log("Case 1: Creating Slice 1")

	err = utils.CmdCreateSlice(1, 1, 1, 1)
	assert.NoError(t, err)

	t.Log("Waiting all slices created")

	err = utils.VerifySliceInitValuesForAllDUs(numTestSlices)
	assert.NoError(t, err)

	if err != nil {
		return
	}
	t.Log("Case 1 passed")

	t.Log("Case 2: Updating Slice 1")
	err = utils.CmdUpdateSlice(1, 1, 1, 1)
	assert.NoError(t, err)

	t.Log("Waiting all slices updated")

	err = utils.VerifySliceUpdatedValuesForAllDUs(numTestSlices)
	assert.NoError(t, err)

	if err != nil {
		return
	}
	t.Log("Case 2 passed")

	t.Log("Case 3: Associating UE 1 with Slice 1")
	err = utils.CmdAssociateUEWithSlice(1, 1, 1, 1, 1)
	assert.NoError(t, err)

	t.Log("Waiting all slices associated")

	err = utils.VerifyUESliceAssociationForAllDUsAndUEs(numTestSlices)
	assert.NoError(t, err)

	if err != nil {
		return
	}
	t.Log("Case 3 passed")

	t.Log("Case 4: Deleting Slice 1")
	err = utils.CmdDeleteSlice(1, 1, 1, 1)
	assert.NoError(t, err)

	t.Log("Waiting all slices deleted")

	err = utils.VerifySliceDeletedForAllDUsAfterUEAssociation()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	t.Log("Case 4 passed")
}

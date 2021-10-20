// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package scalability

import (
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-rsm/pkg/manager"
	"github.com/onosproject/onos-rsm/test/utils"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

const (
	numTestSlices = 3
	numCUs        = 100
	numDUs        = 100
	numUEsPerDU   = 3
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

	t.Log("Adding Mock CUs")
	crtCUErrCh := make(chan error)
	crtCUSuccCh := make(chan struct{})
	crtCUWg := sync.WaitGroup{}
	crtCUWg.Add(numCUs)

	go func() {
		crtCUWg.Wait()
		crtCUSuccCh <- struct{}{}
	}()

	for i := 1; i <= numCUs; i++ {
		go func (i int) {
			defer crtCUWg.Done()
			err := utils.CreateMockE2Node(i, utils.CUDUTypeCU)
			if err != nil {
				crtCUErrCh <- err
			}
		}(i)
	}

	select {
	case e := <- crtCUErrCh:
		assert.NoError(t, e)
		if err != nil {
			return
		}
	case <- crtCUSuccCh:
	}

	t.Log("Adding Mock DUs")
	crtDUErrCh := make(chan error)
	crtDUSuccCh := make(chan struct{})
	crtDUWg := sync.WaitGroup{}
	crtDUWg.Add(numDUs)

	go func() {
		crtDUWg.Wait()
		crtDUSuccCh <- struct{}{}
	}()
	for i := 1; i <= numDUs; i++ {
		go func (i int) {
			err := utils.CreateMockE2Node(i, utils.CUDUTypeDU)
			if err != nil {
				crtDUErrCh <- err
			}
		}(i)
	}

	select {
	case e := <- crtDUErrCh:
		assert.NoError(t, e)
		if err != nil {
			return
		}
	case <- crtDUSuccCh:
	}

	t.Log("Adding Mock UE")
	crtUEErrCh := make(chan error)
	crtUESuccCh := make(chan struct{})
	crtUEWg := sync.WaitGroup{}
	crtUEWg.Add(numDUs * numUEsPerDU)

	go func() {
		crtUEWg.Wait()
		crtUESuccCh <- struct{}{}
	}()
	for i := 1; i <= numDUs; i++ {
		for j := 1; j <= numUEsPerDU; j++ {
			go func(i int, j int) {
				tmpUEID := (i-1)*numUEsPerDU + j
				err := utils.CreateMockUE(i, i, tmpUEID)
				if err != nil {
					crtUEErrCh <- err
				}
			}(i, j)
		}
	}

	select {
	case e := <- crtUEErrCh:
		assert.NoError(t, e)
		if err != nil {
			return
		}
	case <- crtUESuccCh:
	}

	t.Log("Case 1: Creating three slices per DU")
	err = utils.CmdCreateSlice(1, numDUs, 1, numTestSlices)
	assert.NoError(t, err)

	t.Log("Waiting all slices created")

	err = utils.VerifySliceInitValuesForAllDUs(numTestSlices)
	assert.NoError(t, err)

	if err != nil {
		return
	}
	t.Log("Case 1 passed")

	t.Log("Case 2: Updating three slices per DU")
	err = utils.CmdUpdateSlice(1, numDUs, 1, numTestSlices)
	assert.NoError(t, err)

	t.Log("Waiting all slices updated")

	err = utils.VerifySliceUpdatedValuesForAllDUs(numTestSlices)
	assert.NoError(t, err)

	if err != nil {
		return
	}
	t.Log("Case 2 passed")

	t.Log("Case 3: Associating each UE with each slice")
	err = utils.CmdAssociateUEWithSlice(1, numDUs, 1, numTestSlices, numUEsPerDU*numDUs)
	assert.NoError(t, err)

	t.Log("Case 3: Associating UE 1 with Slice 1")

	err = utils.VerifyUESliceAssociationForAllDUsAndUEs(numTestSlices)
	assert.NoError(t, err)

	if err != nil {
		return
	}
	t.Log("Case 3 passed")

	t.Log("Case 4: Deleting all slices")
	err = utils.CmdDeleteSlice(1, numDUs, 1, numTestSlices)
	assert.NoError(t, err)

	t.Log("Waiting all slices deleted")

	err = utils.VerifySliceDeletedForAllDUsAfterUEAssociation()
	assert.NoError(t, err)

	if err != nil {
		return
	}
	t.Log("Case 4 passed")
}

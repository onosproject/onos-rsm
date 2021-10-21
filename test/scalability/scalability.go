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

type mockNodeType int

const (
	CU mockNodeType = iota
	DU
	UE
)

func createMultipleMock(nodeType mockNodeType) error {
	errCh := make(chan error)
	succCh := make(chan struct{})
	wg := sync.WaitGroup{}

	switch nodeType {
	case CU:
		wg.Add(numCUs)
		go func() {
			wg.Wait()
			succCh <- struct{}{}
		}()
		for i := 1; i <= numCUs; i++ {
			go func(i int) {
				defer wg.Done()
				err := utils.CreateMockE2Node(i, utils.CUDUTypeCU)
				if err != nil {
					errCh <- err
				}
			}(i)
		}
	case DU:
		wg.Add(numDUs)
		go func() {
			wg.Wait()
			succCh <- struct{}{}
		}()
		for i := 1; i <= numCUs; i++ {
			go func(i int) {
				defer wg.Done()
				err := utils.CreateMockE2Node(i, utils.CUDUTypeDU)
				if err != nil {
					errCh <- err
				}
			}(i)
		}
	case UE:
		wg.Add(numDUs * numUEsPerDU)
		go func() {
			wg.Wait()
			succCh <- struct{}{}
		}()
		for i := 1; i <= numDUs; i++ {
			for j := 1; j <= numUEsPerDU; j++ {
				go func(i int, j int) {
					defer wg.Done()
					tmpUEID := (i-1)*numUEsPerDU + j
					err := utils.CreateMockUE(i, i, tmpUEID)
					if err != nil {
						errCh <- err
					}
				}(i, j)
			}
		}
	}

	select {
	case e := <-errCh:
		return e
	case <-succCh:
	}
	return nil
}

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
	err = createMultipleMock(CU)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	t.Log("Adding Mock DUs")
	err = createMultipleMock(DU)
	assert.NoError(t, err)
	if err != nil {
		return
	}

	t.Log("Adding Mock UE")
	err = createMultipleMock(UE)
	assert.NoError(t, err)
	if err != nil {
		return
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

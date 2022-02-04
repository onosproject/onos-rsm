// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package slice

import (
	"github.com/onosproject/helmit/pkg/helm"
	"github.com/onosproject/helmit/pkg/input"
	"github.com/onosproject/helmit/pkg/test"
	"github.com/onosproject/onos-rsm/test/utils"
)

type TestSuite struct {
	sdran *helm.HelmRelease
	test.Suite
}

func (s *TestSuite) SetupTestSuite(c *input.Context) error {
	// write files
	err := utils.WriteFile(utils.TLSCaCrtPath, utils.TLSCacrt)
	if err != nil {
		return err
	}
	err = utils.WriteFile(utils.TLSCrtPath, utils.TLSCrt)
	if err != nil {
		return err
	}
	err = utils.WriteFile(utils.TLSKeyPath, utils.TLSKey)
	if err != nil {
		return err
	}
	err = utils.WriteFile("/tmp/config.json", utils.ConfigJSON)
	if err != nil {
		return err
	}

	sdran, err := utils.CreateSdranRelease(c)
	if err != nil {
		return err
	}

	s.sdran = sdran

	return sdran.Install(true)
}

// TearDownTestSuite uninstalls helm chart released
func (s *TestSuite) TearDownTestSuite() error {
	return s.sdran.Uninstall()
}

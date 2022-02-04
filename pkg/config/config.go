// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	app "github.com/onosproject/onos-ric-sdk-go/pkg/config/app/default"
	"github.com/onosproject/onos-ric-sdk-go/pkg/config/event"
	configurable "github.com/onosproject/onos-ric-sdk-go/pkg/config/registry"
	configutils "github.com/onosproject/onos-ric-sdk-go/pkg/config/utils"
)

var log = logging.GetLogger("config")

const (
	// ReportPeriodConfigPath report period config path
	ReportPeriodConfigPath = "/report_period/interval"
)

// Config is an interface for app configuration values
type Config interface {
	// GetReportPeriodWithPath gets report period with a given path
	GetReportPeriodWithPath(path string) (uint64, error)
	// GetReportPeriod gets report period
	GetReportPeriod() (uint64, error)
	// Watch watches config changes
	Watch(context.Context, chan event.Event) error
}

// NewConfig initialize the xApp config
func NewConfig(configPath string) (*AppConfig, error) {
	appConfig, err := configurable.RegisterConfigurable(configPath, &configurable.RegisterRequest{})
	if err != nil {
		return nil, err
	}

	cfg := &AppConfig{
		appConfig: appConfig.Config.(*app.Config),
	}
	return cfg, nil
}

// AppConfig application configuration
type AppConfig struct {
	appConfig *app.Config
}

func (c *AppConfig) Watch(ctx context.Context, ch chan event.Event) error {
	err := c.appConfig.Watch(ctx, ch)
	if err != nil {
		return err
	}
	return nil
}

func (c *AppConfig) GetReportPeriodWithPath(path string) (uint64, error) {
	interval, _ := c.appConfig.Get(path)
	val, err := configutils.ToUint64(interval.Value)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return val, nil
}

func (c *AppConfig) GetReportPeriod() (uint64, error) {
	interval, _ := c.appConfig.Get(ReportPeriodConfigPath)
	val, err := configutils.ToUint64(interval.Value)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return val, nil
}

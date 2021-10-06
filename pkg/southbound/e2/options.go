// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package e2

import (
	"github.com/onosproject/onos-rsm/pkg/broker"
	appConfig "github.com/onosproject/onos-rsm/pkg/config"
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	"github.com/onosproject/onos-rsm/pkg/nib/uenib"
)

type Options struct {
	E2TService ServiceOptions

	ServiceModel ServiceModelOptions

	App AppOptions
}

type AppOptions struct {
	AppID string

	AppConfig *appConfig.AppConfig

	Broker broker.Broker

	RnibClient rnib.TopoClient

	UenibClient uenib.Client

	CtrlReqChsSliceCreate map[string]chan *CtrlMsg

	CtrlReqChsSliceUpdate map[string]chan *CtrlMsg

	CtrlReqChsSliceDelete map[string]chan *CtrlMsg

	CtrlReqChsUeAssociate map[string]chan *CtrlMsg
}

type ServiceOptions struct {
	Host string
	Port int
}

// ServiceModelName is a service model identifier
type ServiceModelName string

// ServiceModelVersion string
type ServiceModelVersion string

// ServiceModelOptions is options for defining a service model
type ServiceModelOptions struct {
	// Name is the service model identifier
	Name ServiceModelName

	// Version is the service model version
	Version ServiceModelVersion
}

type Option interface {
	apply(*Options)
}

type funcOption struct {
	f func(*Options)
}

func (f funcOption) apply(options *Options) {
	f.f(options)
}

func newOption(f func(*Options)) Option {
	return funcOption{
		f: f,
	}
}

func WithE2TAddress(host string, port int) Option {
	return newOption(func(options *Options) {
		options.E2TService.Host = host
		options.E2TService.Port = port
	})
}

func WithServiceModel(name ServiceModelName, version ServiceModelVersion) Option {
	return newOption(func(options *Options) {
		options.ServiceModel = ServiceModelOptions{
			Name:    name,
			Version: version,
		}
	})
}

func WithAppID(appID string) Option {
	return newOption(func(options *Options) {
		options.App.AppID = appID
	})
}

func WithAppConfig(appConfig *appConfig.AppConfig) Option {
	return newOption(func(options *Options) {
		options.App.AppConfig = appConfig
	})
}

func WithBroker(broker broker.Broker) Option {
	return newOption(func(options *Options) {
		options.App.Broker = broker
	})
}

func WithRnibClient(rnibClient rnib.TopoClient) Option {
	return newOption(func(options *Options) {
		options.App.RnibClient = rnibClient
	})
}

func WithUenibClient(uenibClient uenib.Client) Option {
	return newOption(func(options *Options) {
		options.App.UenibClient = uenibClient
	})
}

func WithCtrlReqChs(ctrlReqChsSliceCreate map[string]chan *CtrlMsg,
	ctrlReqChsSliceUpdate map[string]chan *CtrlMsg,
	ctrlReqChsSliceDelete map[string]chan *CtrlMsg,
	ctrlReqChsUeAssociate map[string]chan *CtrlMsg) Option {
	return newOption(func(options *Options) {
		options.App.CtrlReqChsSliceCreate = ctrlReqChsSliceCreate
		options.App.CtrlReqChsSliceUpdate = ctrlReqChsSliceUpdate
		options.App.CtrlReqChsSliceDelete = ctrlReqChsSliceDelete
		options.App.CtrlReqChsUeAssociate = ctrlReqChsUeAssociate
	})
}

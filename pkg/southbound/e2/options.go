// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package e2

import (
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	"github.com/onosproject/onos-rsm/pkg/broker"
	appConfig "github.com/onosproject/onos-rsm/pkg/config"
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	"github.com/onosproject/onos-rsm/pkg/store"
	"github.com/onosproject/onos-rsm/pkg/uenib"
)

// Options E2 client options
type Options struct {
	E2TService ServiceOptions

	E2SubService SubServiceOptions

	ServiceModel ServiceModelOptions

	App AppOptions
}

// AppOptions application options
type AppOptions struct {
	AppID string

	AppConfig *appConfig.AppConfig

	Broker broker.Broker

	RnibClient rnib.TopoClient

	UenibClient uenib.UenibClient

	UEStore store.Store

	SliceStore store.Store

	SliceAssocStore store.Store

	// Ctrl chan - to be removed; now it's just temporal channel map
	CtrlReqChs map[string]chan *e2api.ControlMessage
}

// ServiceOptions are the options for a E2T service
type ServiceOptions struct {
	// Host is the service host
	Host string
	// Port is the service port
	Port int
}

// SubServiceOptions are the options for E2sub service
type SubServiceOptions struct {
	// Host is the service host
	Host string
	// Port is the service port
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

// Option option interface
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

// WithE2TAddress sets the address for the E2T service
func WithE2TAddress(host string, port int) Option {
	return newOption(func(options *Options) {
		options.E2TService.Host = host
		options.E2TService.Port = port
	})
}

// WithE2THost sets the host for the e2t service
func WithE2THost(host string) Option {
	return newOption(func(options *Options) {
		options.E2TService.Host = host
	})
}

// WithE2TPort sets the port for the e2t service
func WithE2TPort(port int) Option {
	return newOption(func(options *Options) {
		options.E2TService.Port = port
	})
}

// WithE2SubAddress sets the address for the E2Sub service
func WithE2SubAddress(host string, port int) Option {
	return newOption(func(options *Options) {
		options.E2SubService.Host = host
		options.E2SubService.Port = port
	})
}

// WithServiceModel sets the client service model
func WithServiceModel(name ServiceModelName, version ServiceModelVersion) Option {
	return newOption(func(options *Options) {
		options.ServiceModel = ServiceModelOptions{
			Name:    name,
			Version: version,
		}
	})
}

// WithAppID sets application ID
func WithAppID(appID string) Option {
	return newOption(func(options *Options) {
		options.App.AppID = appID
	})
}

// WithAppConfig sets the app config interface
func WithAppConfig(appConfig *appConfig.AppConfig) Option {
	return newOption(func(options *Options) {
		options.App.AppConfig = appConfig
	})
}

// WithBroker sets subscription broker
func WithBroker(broker broker.Broker) Option {
	return newOption(func(options *Options) {
		options.App.Broker = broker
	})
}

// WithRnibClient sets rnib client
func WithRnibClient(rnibClient rnib.TopoClient) Option {
	return newOption(func(options *Options) {
		options.App.RnibClient = rnibClient
	})
}

// WithUenibClient sets uenib client
func WithUenibClient(uenibClient uenib.UenibClient) Option {
	return newOption(func(options *Options) {
		options.App.UenibClient = uenibClient
	})
}

// WithUEStore sets ue store
func WithUEStore(s store.Store) Option {
	return newOption(func(options *Options) {
		options.App.UEStore = s
	})
}

// WithSliceStore sets slice store
func WithSliceStore(s store.Store) Option {
	return newOption(func(options *Options) {
		options.App.SliceStore = s
	})
}

// WithSliceAssocStore sets slice assoc store
func WithSliceAssocStore(s store.Store) Option {
	return newOption(func(options *Options) {
		options.App.SliceAssocStore = s
	})
}

// WithCtrlReqChs sets the map of control request message channel
func WithCtrlReqChs(m map[string]chan *e2api.ControlMessage) Option {
	return newOption(func(options *Options) {
		options.App.CtrlReqChs = m
	})
}

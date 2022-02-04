// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package monitoring

import (
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	e2sm_rsm "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-rsm-ies"
	e2client "github.com/onosproject/onos-ric-sdk-go/pkg/e2/v1beta1"
	"github.com/onosproject/onos-rsm/pkg/broker"
	appConfig "github.com/onosproject/onos-rsm/pkg/config"
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	"github.com/onosproject/onos-rsm/pkg/nib/uenib"
)

type Options struct {
	App AppOptions

	Monitor MonitorOptions
}

type AppOptions struct {
	AppConfig *appConfig.AppConfig

	RnibClient rnib.TopoClient

	UenibClient uenib.Client

	EventTriggerType e2sm_rsm.RsmRicindicationTriggerType
}

type MonitorOptions struct {
	Node         e2client.Node
	NodeID       topoapi.ID
	StreamReader broker.StreamReader
}

type Option interface {
	apply(options *Options)
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

// WithNode sets node
func WithNode(node e2client.Node) Option {
	return newOption(func(options *Options) {
		options.Monitor.Node = node
	})
}

// WithNodeID sets node ID
func WithNodeID(nodeID topoapi.ID) Option {
	return newOption(func(options *Options) {
		options.Monitor.NodeID = nodeID
	})
}

// WithStreamReader sets stream reader
func WithStreamReader(streamReader broker.StreamReader) Option {
	return newOption(func(options *Options) {
		options.Monitor.StreamReader = streamReader
	})
}

// WithAppConfig sets app configs
func WithAppConfig(cfg *appConfig.AppConfig) Option {
	return newOption(func(options *Options) {
		options.App.AppConfig = cfg
	})
}

// WithRNIBClient sets rnib client
func WithRNIBClient(client rnib.TopoClient) Option {
	return newOption(func(options *Options) {
		options.App.RnibClient = client
	})
}

// WithUENIBClient sets uenib client
func WithUENIBClient(client uenib.Client) Option {
	return newOption(func(options *Options) {
		options.App.UenibClient = client
	})
}

// WithRicIndicationTriggerType sets ric indication event trigger type
func WithRicIndicationTriggerType(triggerType e2sm_rsm.RsmRicindicationTriggerType) Option {
	return newOption(func(options *Options) {
		options.App.EventTriggerType = triggerType
	})
}

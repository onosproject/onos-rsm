// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package slicing

import (
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	"github.com/onosproject/onos-rsm/pkg/nib/uenib"
	"github.com/onosproject/onos-rsm/pkg/northbound"
	"github.com/onosproject/onos-rsm/pkg/southbound/e2"
)

type Options struct {
	Chans Channels

	App AppOptions
}

type Channels struct {
	RsmMsgCh chan *northbound.RsmMsg

	CtrlReqChsSliceCreate map[string]chan *e2.CtrlMsg

	CtrlReqChsSliceUpdate map[string]chan *e2.CtrlMsg

	CtrlReqChsSliceDelete map[string]chan *e2.CtrlMsg

	CtrlReqChsUeAssociate map[string]chan *e2.CtrlMsg
}

type AppOptions struct {
	RnibClient rnib.TopoClient

	UenibClient uenib.Client

	AckTimer int
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

func WithCtrlReqChs(ctrlReqChsSliceCreate map[string]chan *e2.CtrlMsg,
	ctrlReqChsSliceUpdate map[string]chan *e2.CtrlMsg,
	ctrlReqChsSliceDelete map[string]chan *e2.CtrlMsg,
	ctrlReqChsUeAssociate map[string]chan *e2.CtrlMsg) Option {
	return newOption(func(options *Options) {
		options.Chans.CtrlReqChsSliceCreate = ctrlReqChsSliceCreate
		options.Chans.CtrlReqChsSliceUpdate = ctrlReqChsSliceUpdate
		options.Chans.CtrlReqChsSliceDelete = ctrlReqChsSliceDelete
		options.Chans.CtrlReqChsUeAssociate = ctrlReqChsUeAssociate
	})
}

func WithNbiReqChs(rsmMsgCh chan *northbound.RsmMsg) Option {
	return newOption(func(options *Options) {
		options.Chans.RsmMsgCh = rsmMsgCh
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

func WithAckTimer(ackTimer int) Option {
	return newOption(func(options *Options) {
		options.App.AckTimer = ackTimer
	})
}

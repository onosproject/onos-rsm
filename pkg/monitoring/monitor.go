// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package monitoring

import (
	"context"
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	e2sm_rsm "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm/v1/e2sm-rsm-ies"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-rsm/pkg/broker"
	appConfig "github.com/onosproject/onos-rsm/pkg/config"
	"github.com/onosproject/onos-rsm/pkg/nib/rnib"
	"google.golang.org/protobuf/proto"
)

var log = logging.GetLogger("monitoring")

// NewMonitor returns new Monitor
func NewMonitor(opts ...Option) *Monitor {
	options := Options{}
	for _, opt := range opts {
		opt.apply(&options)
	}

	return &Monitor{
		streamReader:           options.Monitor.StreamReader,
		appConfig:              options.App.AppConfig,
		nodeID:                 options.Monitor.NodeID,
		rnibClient:             options.App.RNIBClient,
		ricIndEventTriggerType: options.App.EventTriggerType,
	}
}

// Monitor is a struct to monitor indication messages
type Monitor struct {
	streamReader           broker.StreamReader
	appConfig              *appConfig.AppConfig
	nodeID                 topoapi.ID
	rnibClient             rnib.Client
	ricIndEventTriggerType e2sm_rsm.RsmRicindicationTriggerType
}

// Start start monitoring of indication messages for a given subscription ID
func (m *Monitor) Start(ctx context.Context) error {
	errCh := make(chan error)
	go func() {
		for {
			indMsg, err := m.streamReader.Recv(ctx)
			if err != nil {
				errCh <- err
			}
			err = m.processIndication(ctx, indMsg, m.nodeID)
			if err != nil {
				errCh <- err
			}
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *Monitor) processIndication(ctx context.Context, indMsg e2api.Indication, nodeID topoapi.ID) error {
	indHeader := e2sm_rsm.E2SmRsmIndicationHeader{}
	indPayload := e2sm_rsm.E2SmRsmIndicationMessage{}

	err := proto.Unmarshal(indMsg.Header, &indHeader)
	if err != nil {
		return err
	}

	err = proto.Unmarshal(indMsg.Payload, &indPayload)
	if err != nil {
		return err
	}

	if indPayload.GetIndicationMessageFormat1() != nil {
		err = m.processMetricTypeMessage(ctx, indHeader.GetIndicationHeaderFormat1(), indPayload.GetIndicationMessageFormat1())
		if err != nil {
			return err
		}
	}

	if indPayload.GetIndicationMessageFormat2() != nil {
		err = m.processEmmEventMessage(ctx, indPayload.GetIndicationMessageFormat2())
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Monitor) processMetricTypeMessage(ctx context.Context, indHdr *e2sm_rsm.E2SmRsmIndicationHeaderFormat1, indMsg *e2sm_rsm.E2SmRsmIndicationMessageFormat1) error {

	log.Infof("Received indication message (Metric): %v", indMsg)

	return nil
}

func (m *Monitor) processEmmEventMessage(ctx context.Context, indMsg *e2sm_rsm.E2SmRsmIndicationMessageFormat2) error {

	log.Infof("Received indication message (EMM): %v", indMsg)

	return nil
}

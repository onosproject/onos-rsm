# onos-rsm
The xApplication for ONOS-SDRAN (µONOS Architecture) to manage RAN slices

## Overview
The `onos-rsm` is the xApplication running over ONOS SD-RAN for `RAN Slice Management (RSM)`.
The RAN slice has definitions related with `Quality of Service (QoS)` for associated UEs, such as `time frame rates` and `scheduling algorithms`.
In order to manage the RAN slice, this xApplication `creates`, `deletes`, and `updates` RAN slices through CLI.
Also, this xApplication associates a specific UE to a RAN slice so that the UE can achieve the QoS that is defined in the associated RAN slice.
The `onos-rsm` xApplication subscribes CU E2 nodes as well as DU E2 nodes.
CU E2 nodes report the `EPC Mobility Management (EMM) event` to the `onos-rsm` xApp.
On the other hands, DU E2 nodes receive control messages for RAN slice management and UE-slice association.
The `onos-rsm` xApplication stores all RAN slice information to `onos-topo` and `onos-uenib`.

## Interaction with other ONOS SD-RAN micro-services
To begin with, `onos-rsm` subscribes `CUs` and `DUs` through `onos-e2t`.
Once `UE` is attached, the `CU` send the indication message to `onos-rsm` to report that the `UE` is connected.
Then, `onos-rsm` stores the attached `UE` information to `onos-uenib`.
When a user creates/deletes/updates a slice through `onos-cli`, `onos-rsm` sends a control message to `DU`.
Likewise, the user associates `UE` with a created slice through `onos-cli`, `onos-rsm` sends a control message to `DU`.
After sending the control message successfully, `onos-rsm` updates `onos-topo` and `onos-uenib`, accordingly.

## Command Line Interface
Go to `onos-cli` and command below for each purpose

* Create slice
  * DU_E2_NODE_ID: target DU's E2 Node ID (e.g., e2:4/e00/3/c8).
  * SCHEDULER_TYPE: scheduler type such as round robin (RR) and proportional fair (PF).
  * SLICE_ID: this slice's ID (e.g., 1).
  * WEIGHT: time frame rates (e.g., 30). It should be less than 80.
  * SLICE_TYPE: downlink (DL) or uplink (UL).

```bash
onos-cli$ kubectl exec -it deployment/onos-cli -n riab -- onos rsm create slice --e2NodeID <DU_E2_NODE_ID> --scheduler <SCHEDULER_TYPE> --sliceID <SLICE_ID> --weight <WEIGHT> --sliceType <SLICE_TYPE>

# example:
onos-cli$ kubectl exec -it deployment/onos-cli -n riab -- onos rsm create slice --e2NodeID e2:4/e00/3/c8 --scheduler RR --sliceID 1 --weight 30 --sliceType DL
```

* Update slice
  * DU_E2_NODE_ID: target DU's E2 Node ID (e.g., e2:4/e00/3/c8).
  * SCHEDULER_TYPE: scheduler type such as round robin (RR) and proportional fair (PF).
  * SLICE_ID: this slice's ID (e.g., 1).
  * WEIGHT: time frame rates (e.g., 30). It should be less than 80.
  * SLICE_TYPE: downlink (DL) or uplink (UL).

```bash
onos-cli$ kubectl exec -it deployment/onos-cli -n riab -- onos rsm update slice --e2NodeID <DU_E2_NODE_ID> --scheduler <SCHEDULER_TYPE> --sliceID <SLICE_ID> --weight <WEIGHT> --sliceType <SLICE_TYPE>

# example:
onos-cli$ kubectl exec -it deployment/onos-cli -n riab -- onos rsm update slice --e2NodeID e2:4/e00/3/c8 --scheduler RR --sliceID 1 --weight 50 --sliceType DL
```

* Delete slice
  * DU_E2_NODE_ID: target DU's E2 Node ID (e.g., e2:4/e00/3/c8).
  * SLICE_ID: this slice's ID (e.g., 1).
  * SLICE_TYPE: downlink (DL) or uplink (UL).

```bash
onos-cli$ kubectl exec -it deployment/onos-cli -n riab -- onos rsm delete slice --e2NodeID <DU_E2_NODE_ID> --sliceID <SLICE_ID> --sliceType <SLICE_TYPE>

# example:
onos-cli$ kubectl exec -it deployment/onos-cli -n riab -- onos rsm delete slice --e2NodeID e2:4/e00/3/c8 --sliceID 1 --sliceType DL
```

* UE-slice association
  * DU_E2_NODE_ID: target DU's E2 Node ID (e.g., e2:4/e00/3/c8).
  * SLICE_ID: this slice's ID (e.g., 1).
  * DRB_ID: DRB-ID for the bearer. It should be in onos-uenib.
  * DU_UE_F1AP_ID: DU_UE_F1AP_ID. It should be in onos-uenib.

```bash
onos-cli$ kubectl exec -it deployment/onos-cli -n riab -- onos rsm set association --dlSliceID <SLICE_ID> --e2NodeID <DU_E2_NODE_ID> --drbID <DRB_ID> --DuUeF1apID <DU_UE_F1AP_ID>

# example:
onos-cli$ kubectl exec -it deployment/onos-cli -n riab -- onos rsm set association --dlSliceID 1 --e2NodeID e2:4/e00/3/c8 --drbID 5 --DuUeF1apID 1240
```
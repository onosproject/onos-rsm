module github.com/onosproject/onos-rsm

go 1.16

require (
	github.com/gogo/protobuf v1.3.2
	github.com/google/uuid v1.1.2
	github.com/onosproject/onos-api/go v0.7.99
	github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm v0.0.0-20210920155345-7d0967cbdcd0
	github.com/onosproject/onos-lib-go v0.7.20
	github.com/onosproject/onos-ric-sdk-go v0.7.26
	google.golang.org/grpc v1.33.2
	google.golang.org/protobuf v1.26.0
)

replace github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rsm => /Users/woojoong/workspace/onf/sdran/onos-e2-sm/servicemodels/e2sm_rsm

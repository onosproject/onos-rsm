ARG ONOS_BUILD_VERSION=undefined

FROM onosproject/golang-build:v0.6.7 as build

ENV GO111MODULE=on
ARG ONOS_MAKE_TARGET=build

COPY Makefile go.mod go.sum /go/src/github.com/onosproject/onos-rsm/
COPY cmd/ /go/src/github.com/onosproject/onos-rsm/cmd/
COPY pkg/ /go/src/github.com/onosproject/onos-rsm/pkg/
COPY vendor/ /go/src/github.com/onosproject/onos-rsm/vendor/
COPY api/ /go/src/github.com/onosproject/onos-rsm/api

RUN cd /go/src/github.com/onosproject/onos-rsm && GOFLAGS=-mod=vendor make ${ONOS_MAKE_TARGET}

FROM alpine:3.11
RUN apk add libc6-compat

USER nobody

COPY --from=build /go/src/github.com/onosproject/onos-rsm/build/_output/onos-rsm /usr/local/bin/onos-rsm

ENTRYPOINT ["onos-rsm"]

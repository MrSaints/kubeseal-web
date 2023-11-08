# syntax=docker/dockerfile:1.2

FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS dev
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

LABEL org.label-schema.vcs-url="https://github.com/MrSaints/kubeseal-web" \
      maintainer="Ian L. <os@fyianlai.com>"

WORKDIR /kubeseal-web/

RUN apk add --no-cache build-base curl

ENV GO111MODULE on
ENV GOPROXY https://proxy.golang.org
ENV PATH="${PATH}:/go/bin"

COPY go.mod go.sum /kubeseal-web/

RUN go mod download \
    && go get github.com/markbates/pkger/cmd/pkger

FROM --platform=$BUILDPLATFORM dev AS build
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

COPY ./ /kubeseal-web/

ENV PATH="${PATH}:/go/bin"

RUN mkdir /build/ \
    && pkger -include /static/ \
    && rm -rf /kubeseal-web/static/ \
    && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v \
       -ldflags "-s -w" -a -installsuffix cgo \
       -o /build/kubeseal-web . \
    && chmod +x /build/kubeseal-web

FROM --platform=$TARGETPLATFORM alpine:3.12 AS prod
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

LABEL org.label-schema.vcs-url="https://github.com/MrSaints/kubeseal-web" \
      maintainer="Ian L. <os@fyianlai.com>"

RUN apk add --no-cache bash ca-certificates curl jq wget nano

COPY --from=build /build/kubeseal-web /kubeseal-web/run

ARG BUILD_VERSION
ENV KSWEB_VERSION $BUILD_VERSION
ENV GIN_MODE release

ENTRYPOINT ["/kubeseal-web/run"]

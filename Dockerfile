FROM golang:1.15-alpine AS dev

LABEL org.label-schema.vcs-url="https://github.com/MrSaints/kubeseal-web" \
      maintainer="Ian L. <os@fyianlai.com>"

WORKDIR /kubeseal-web/

RUN apk add --no-cache build-base curl

ENV GO111MODULE on
ENV GOPROXY https://proxy.golang.org

COPY go.mod go.sum /kubeseal-web/

RUN go mod download \
    && go get github.com/markbates/pkger/cmd/pkger


FROM dev as build

COPY ./ /kubeseal-web/

RUN mkdir /build/

RUN pkger -include /static/ \
    && rm -rf /kubeseal-web/static/

RUN CGO_ENABLED=0 \
    go build -v \
    -ldflags "-s" -a -installsuffix cgo \
    -o /build/kubeseal-web \
    /kubeseal-web/ \
    && chmod +x /build/kubeseal-web


FROM alpine:3.12 AS prod

LABEL org.label-schema.vcs-url="https://github.com/MrSaints/kubeseal-web" \
      maintainer="Ian L. <os@fyianlai.com>"

RUN apk add --no-cache bash ca-certificates curl jq wget nano

COPY --from=build /build/kubeseal-web /kubeseal-web/run

RUN curl -sL https://github.com/bitnami-labs/sealed-secrets/releases/download/v0.13.1/kubeseal-linux-amd64 -o /usr/local/bin/kubeseal
RUN chmod +x /usr/local/bin/kubeseal

ARG BUILD_VERSION
ENV KSWEB_VERSION $BUILD_VERSION
ENV GIN_MODE release

ENTRYPOINT ["/kubeseal-web/run"]

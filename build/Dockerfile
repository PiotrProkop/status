# golang:1.16
ARG BUILDER_IMAGE=docker.io/library/golang@sha256:f254180c5defa2653955e963fb0626e3d4fbbb162f7cff6490e94607d1d867ff

ARG RUNTIME_IMAGE=scratch

FROM ${BUILDER_IMAGE} AS builder
ENV CGO_ENABLED=0

COPY . /status-server
WORKDIR /status-server
RUN go build -ldflags='-w -s -extldflags "-static"' -a -v /status-server/cmd/status-server

FROM ${RUNTIME_IMAGE}
COPY --from=builder /status-server/status-server /
# copy certificates to avoid x509: certificate signed by unknown authority
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT [ "/status-server" ]

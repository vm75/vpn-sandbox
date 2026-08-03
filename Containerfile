# Stage 1: Build dante sockd and vpn-sandbox
FROM golang:alpine AS build

RUN apk --no-cache update && apk --no-cache upgrade
RUN apk add --no-cache build-base

WORKDIR /workdir
ENV CGO_ENABLED=1
ARG VERSION=dev

COPY server/go.mod server/go.sum /workdir/server/
RUN go mod download -C /workdir/server

COPY server /workdir/server

RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux GOARCH=${TARGETARCH} go build -C /workdir/server \
    -ldflags="-s -w -X vpn-sandbox/core.Version=${VERSION}" -o /workdir/vpn-sandbox

# Stage 2: Create the runtime image
FROM alpine:latest AS runtime

RUN apk --no-cache update && apk --no-cache upgrade
RUN apk add --no-cache --no-progress ip6tables iptables bind-tools inotify-tools \
    openvpn wireguard-tools-wg tinyproxy dante-server
RUN ln -s /usr/sbin/sockd /usr/bin/sockd

COPY --from=build /workdir/vpn-sandbox /opt/vpn-sandbox/vpn-sandbox
COPY usr /usr
COPY server/static /opt/vpn-sandbox/static
COPY vpn-sandbox.png /opt/vpn-sandbox/static/assets

RUN addgroup root openvpn && mkdir -p /data

VOLUME ["/data"]
EXPOSE 80/tcp 1080/tcp 3128/tcp

HEALTHCHECK --interval=60s --timeout=15s --start-period=120s \
    CMD netstat -an | grep -c ":::80 "

ENTRYPOINT ["/opt/vpn-sandbox/vpn-sandbox"]

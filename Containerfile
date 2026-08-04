# Stage 1: Build 3proxy
FROM alpine:latest AS proxy-build

ARG THREEPROXY_VERSION=0.9.6
RUN apk add --no-cache build-base linux-headers wget
WORKDIR /workdir/3proxy
RUN wget -qO- "https://github.com/3proxy/3proxy/archive/refs/tags/${THREEPROXY_VERSION}.tar.gz" \
    | tar -xz --strip-components=1 \
    && make -f Makefile.Linux PLUGINS=

# Stage 2: Build vpn-sandbox
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

# Stage 3: Create the runtime image
FROM alpine:latest AS runtime

RUN apk --no-cache update && apk --no-cache upgrade
RUN apk add --no-cache --no-progress ip6tables iptables bind-tools inotify-tools \
    openvpn wireguard-tools-wg

COPY --from=proxy-build /workdir/3proxy/bin/3proxy /usr/bin/3proxy
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

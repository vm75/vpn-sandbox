# Stage 1: Build dante sockd and vpn-sandbox
FROM golang:alpine AS build

# Set the working directory
WORKDIR /workdir

# Install build dependencies
RUN apk add --no-cache build-base upx

# Build dante from source
ARG DANTE_VERSION=1.4.3
RUN wget https://www.inet.no/dante/files/dante-$DANTE_VERSION.tar.gz --output-document - | tar -xz && \
    cd dante-$DANTE_VERSION && \
    ac_cv_func_sched_setscheduler=no ./configure --disable-client && \
    make install

# Download go dependencies
COPY server/go.mod server/go.sum /workdir/server/
RUN go mod download -C /workdir/server

# Copy the server code
COPY server /workdir/server

# Build go server
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -C /workdir/server -ldflags="-s -w" -o /workdir/vpn-sandbox

# Compress the binary
RUN upx /workdir/vpn-sandbox

# Stage 2: Create the final minimal image
FROM alpine

RUN apk --no-cache update
RUN apk --no-cache upgrade
RUN apk --no-cache --no-progress add ip6tables iptables bind-tools tinyproxy inotify-tools openvpn wireguard-tools-wg

ARG IMAGE_VERSION
ARG BUILD_DATE

LABEL source="github.com/vm75/vpn-sandbox"
LABEL version="$IMAGE_VERSION"
LABEL created="$BUILD_DATE"

# Copy binaries from build stage
COPY --from=build /usr/local/sbin/sockd /usr/local/sbin/sockd
COPY --from=build /workdir/vpn-sandbox /opt/vpn-sandbox/vpn-sandbox

# Copy the server code
COPY usr /usr
COPY server/static /opt/vpn-sandbox/static

ENV VPN_LOG_LEVEL=3 \
    KILL_SWITCH=on \
    HTTP_PROXY=on \
    SOCKS_PROXY=on

RUN mkdir -p /data
RUN addgroup root openvpn

VOLUME ["/data"]

# expose ports for http-proxy, socks-proxy and vpn-sandbox
EXPOSE 8080/tcp 1080/tcp 80/tcp

HEALTHCHECK --interval=60s --timeout=15s --start-period=120s \
    CMD netstat -an | grep -c ":::80 "

ENTRYPOINT [ "/opt/vpn-sandbox/vpn-sandbox", "-s" ]

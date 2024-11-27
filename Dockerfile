# Stage 1: Build dante sockd and vpn-sandbox
FROM golang:alpine AS build

# Install required build tools
RUN apk add --no-cache build-base

# Set the working directory
WORKDIR /workdir

# Set environment variables for cross-compilation
ENV CGO_ENABLED=1

# Copy Go modules files first for caching
COPY server/go.mod server/go.sum /workdir/server/
RUN go mod download -C /workdir/server

# Copy the rest of the application source code
COPY server /workdir/server

# Build the application for multiple architectures
# The build command will later be specified by Buildx
RUN --mount=type=cache,target=/root/.cache/go-build \
    GOOS=linux GOARCH=${TARGETARCH} go build -C /workdir/server -ldflags="-s -w" -o /workdir/vpn-sandbox

# Stage 2: Create the final minimal image
FROM alpine:latest AS runtime

# Install required packages
RUN apk --no-cache update
RUN apk --no-cache upgrade
RUN apk --no-cache --no-progress add ip6tables iptables bind-tools inotify-tools \
    openvpn wireguard-tools-wg tinyproxy dante-server
RUN ln -s /usr/sbin/sockd /usr/bin/sockd

# Copy binaries from build stage
COPY --from=build /workdir/vpn-sandbox /opt/vpn-sandbox/vpn-sandbox

# Copy the server code
COPY usr /usr
COPY server/static /opt/vpn-sandbox/static

# Create the openvpn group
RUN addgroup root openvpn

# Create the data directory
RUN mkdir -p /data

# Set the volume
VOLUME ["/data"]

# Expose ports for vpn-sandbox, http-proxy, socks-proxy
EXPOSE 80/tcp 1080/tcp 3128/tcp

# Healthcheck
HEALTHCHECK --interval=60s --timeout=15s --start-period=120s \
    CMD netstat -an | grep -c ":::80 "

# Run vpn-sandbox
ENTRYPOINT [ "/opt/vpn-sandbox/vpn-sandbox" ]

# Stage 1: Build 3proxy
FROM python:3.11-slim-bookworm AS proxy-build

ARG THREEPROXY_VERSION=0.9.7
RUN apt-get update && apt-get install -y --no-install-recommends build-essential wget libssl-dev libpcre2-dev && apt-get clean && rm -rf /var/lib/apt/lists/*
WORKDIR /workdir/3proxy
RUN wget -qO- "https://github.com/3proxy/3proxy/archive/refs/tags/${THREEPROXY_VERSION}.tar.gz" \
    | tar -xz --strip-components=1 \
    && make -f Makefile.Linux PLUGINS=

# Stage 2: Create the runtime image
FROM python:3.11-slim-bookworm AS runtime

RUN apt-get update && apt-get install -y --no-install-recommends \
    iproute2 iptables procps dnsutils inotify-tools openvpn wireguard-tools libpcre2-8-0 libssl3 net-tools \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=proxy-build /workdir/3proxy/bin/3proxy /usr/bin/3proxy

COPY server/requirements.txt /opt/vpn-sandbox/requirements.txt
RUN pip install --no-cache-dir -r /opt/vpn-sandbox/requirements.txt

COPY server /opt/vpn-sandbox
COPY vpn-sandbox.png /opt/vpn-sandbox/static/assets/

RUN addgroup --system openvpn && adduser root openvpn && mkdir -p /data

VOLUME ["/data"]
EXPOSE 80/tcp 1080/tcp 3128/tcp

HEALTHCHECK --interval=60s --timeout=15s --start-period=120s \
    CMD netstat -an | grep -c ":::80 "

WORKDIR /opt/vpn-sandbox
ENTRYPOINT ["python3", "main.py"]

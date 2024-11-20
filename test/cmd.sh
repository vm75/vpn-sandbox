#!/bin/sh

TEST_DIR=/src/test
CONFIG_DIR=${TEST_DIR}/config
VAR_DIR=${TEST_DIR}/var
VPN_UP_SCRIPT=/bin/vpn-up
VPN_DOWN_SCRIPT=/bin/vpn-down
DEFAULT_GW=$(awk '/^nameserver/ {print $2}' /etc/resolv.conf.bak 2>/dev/null)

setup() {
  mkdir -p ${VAR_DIR}
  ln -s ${TEST_DIR}/cmd.sh /bin/cmd
  ln -s ${TEST_DIR}/cmd.sh /bin/vpn-up
  ln -s ${TEST_DIR}/cmd.sh /bin/vpn-down
  apk --no-cache update
  apk --no-cache upgrade
  apk --no-cache --no-progress add ip6tables iptables bind-tools inotify-tools curl build-base openvpn wireguard-tools-wg
  cp -a /etc/resolv.conf /etc/resolv.conf.bak
  server
}

server() {
  while true; do
    {
      echo -e 'HTTP/1.1 200 OK\r\n'
      echo "Port = 80";
      echo "WAN IP = $(curl -s ifconfig.me)";
      echo "Date = $(date)";
    } | nc -l -p 80 &> /dev/null
  done &
  while true; do
    {
      echo -e 'HTTP/1.1 200 OK\r\n'
      echo "Port = 81";
      echo "WAN IP = $(curl -s ifconfig.me)";
      echo "Date = $(date)";
    } | nc -l -p 81 &> /dev/null
  done &
}

vpn_up() {
  env > /src/test/var/vpn-up.log
  echo "$@" >> /src/test/var/vpn-up.log

  echo "nameserver ${vpn_dns}" > /etc/resolv.conf

  # Remove all existing default routes
  ip route del default

  # Default route for all traffic through the VPN
  ip route add default via ${vpn_gw} dev tun0

  # Flush existing rules to start fresh
  iptables -F

  # Allow incoming ESTABLISHED and RELATED connections on the VPN interface
  iptables -A INPUT -i tun0 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

  # Drop all other incoming connections on the VPN interface
  iptables -A INPUT -i tun0 -j DROP
}

vpn_down() {
  env > /src/test/var/vpn-down.log
  echo "$@" >> /src/test/var/vpn-down.log

  # Add nameserver
  cat /etc/resolv.conf.bak > /etc/resolv.conf

  # Remove all existing default routes
  ip route del default

  # Add default gateway
  ip route add default via ${DEFAULT_GW} dev eth0

  # Flush existing rules (to ensure clean slate)
  iptables -F

  # Allow related and established connections (for existing sessions to work)
  iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

  # Allow incoming connections only on port 80
  iptables -A INPUT -p tcp --dport 80 -j ACCEPT
  iptables -A INPUT -p tcp -j DROP
}

run_openvpn() {
  VPN_CONF=${VAR_DIR}/vpn.ovpn
  VPN_AUTH=${VAR_DIR}/vpn.auth
  RETRY_INTERVAL=30
  VPN_SANDBOX_LOG=${VAR_DIR}/vpn-sandbox.log
  openvpn \
    --client \
    --cd ${VAR_DIR} \
    --config ${VPN_CONF} \
    --auth-user-pass ${VPN_AUTH} \
    --auth-nocache \
    --verb ${VPN_LOG_LEVEL:-3} \
    --log ${VAR_DIR}/openvpn.log \
    --status ${VAR_DIR}/openvpn.status ${RETRY_INTERVAL} \
    --ping-restart ${RETRY_INTERVAL} \
    --connect-retry-max 3 \
    --script-security 2 \
    --up /bin/vpn-up --up-delay \
    --down /bin/vpn-down \
    --up-restart \
    --pull-filter ignore route-ipv6 \
    --pull-filter ignore ifconfig-ipv6 \
    --pull-filter ignore block-outside-dns \
    --redirect-gateway def1 \
    --remote-cert-tls server \
    --data-ciphers AES-256-GCM:AES-128-GCM:CHACHA20-POLY1305:AES-256-CBC:AES-128-CBC \
    >> ${VPN_SANDBOX_LOG} 2>&1 &
}

tunnel_up() {
  WG_CONF=${CONFIG_DIR}/wg0.conf
  PRIVATE_KEY=$(awk '/^PrivateKey/ {print $3}' ${WG_CONF})
  PUBLIC_KEY=$(awk '/^PublicKey/ {print $3}' ${WG_CONF})
  ALLOWED_IPS=$(awk '/^AllowedIPs/ {print $3}' ${WG_CONF})
  ADDRESS=$(awk '/^Address/ {print $3}' ${WG_CONF})
  ENDPOINT=$(awk '/^Endpoint/ {print $3}' ${WG_CONF})
  ENDPOINT_IP=$(dig +short ${ENDPOINT/:*})

  ip link add dev wg0 type wireguard
  wg set wg0 \
      private-key <(echo "${PRIVATE_KEY}") \
      peer "${PUBLIC_KEY}" \
      endpoint "${ENDPOINT}" \
      allowed-ips "${ALLOWED_IPS}"
  ip address add "${ADDRESS}" dev wg0
  ip link set up dev wg0

  awk -F ' = ' '/^DNS =/ {gsub(/, /, " ", $2); split($2, dns, " "); for (i in dns) print "nameserver " dns[i]}' ${WG_CONF} > /etc/resolv.conf

  # Remove all existing default routes
  ip route del default

  # Default route for all traffic through the VPN
  ip route add default dev wg0

  # Default route for all traffic through the VPN
  ip route add ${ENDPOINT_IP} via ${DEFAULT_GW}

  # Flush existing rules to start fresh
  iptables -F

  # Allow incoming ESTABLISHED and RELATED connections on the VPN interface
  iptables -A INPUT -i wg0 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

  # Drop all other incoming connections on the VPN interface
  iptables -A INPUT -i wg0 -j DROP
}

tunnel_down() {
  ip link set down dev wg0
  ip link del dev wg0
  vpn_down
}

main() {
  case "$1" in
    run|start)
      if $(command -v docker &> /dev/null) ; then
        docker compose up -d
        docker exec test /src/test/cmd.sh setup
        docker exec -ti test sh
      fi
      ;;
    stop)
      if $(command -v docker &> /dev/null); then
        docker compose down
      fi
      ;;
    setup)
      setup
      ;;
    vpn)
      case "$2" in
        up)
          run_openvpn
          ;;
        down)
          pkill -15 openvpn
          ;;
        reset)
          cmd vpn down ; cmd build ; cmd vpn up
          ;;
      esac
      ;;
    wg)
      case "$2" in
        up)
          tunnel_up
          ;;
        down)
          tunnel_down
          ;;
      esac
      ;;
    build)
      cd /src/server
      CGO_ENABLED=1 go build -o ${VAR_DIR}/vpn-sandbox .
      ;;
    serve)
      server
      ;;
    *)
      echo "Unknown command: $1"
      exit 1
      ;;
  esac
}

script=$(basename "$0")
cd $(dirname $(readlink -f "$0"))
case "$script" in
  cmd|cmd.sh)
    main "$@"
    ;;
  vpn-up)
    vpn_up
    ;;
  vpn-down)
    vpn_down
    ;;
esac
exit 0
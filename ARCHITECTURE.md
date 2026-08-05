# Architecture

VPN Sandbox is one Go process with a browser UI, SQLite configuration storage, VPN process control, and optional proxy processes. The runtime image supplies the network tools and proxy daemons; the Go process owns their configuration and lifecycle.

## Components

- `server/webserver` serves `server/static` and exposes the REST API plus the `/api/status` event stream.
- `server/core` owns the shared data paths, SQLite connection, global configuration, and registered modules.
- `server/modules/openvpn` and `server/modules/wireguard` implement VPN server templates and tunnel lifecycle.
- `server/modules/proxy` manages 3proxy HTTP and SOCKS5 processes and reacts to VPN/global configuration events.
- `server/actions` applies DNS, default routes, firewall rules, the optional user-managed app-script hook, and configured-app lifecycle commands.
- `server/utils` provides command execution, filesystem/network helpers, logging, signals, and the event bus.

## Startup and state flow

1. `server/main.go` parses flags and initializes `/data/config` and `/data/var`.
2. The web-server mode opens SQLite, restores global configuration, writes the PID file, and establishes the VPN-down network policy.
3. Modules register their API routes and the HTTP server starts (port `80` by default).
4. OpenVPN invokes the executable as an up/down script. Script mode records OpenVPN environment data and signals the main process. WireGuard calls the VPN actions directly.
5. `VpnUp` updates DNS, routes, and firewall rules, then publishes `vpn-up` and starts legacy and configured apps; `VpnDown` stops both app types before restoring the disconnected policy and publishing `vpn-down`.
6. Proxy modules start only when enabled and a VPN is active. Status listeners refresh the UI and SSE clients.

## Persistence and boundaries

Global and module configuration is JSON stored in the `configs` SQLite table. OpenVPN and WireGuard server templates use their own SQLite tables. Generated files and logs live under `/data/var`; credentials must not be exposed or committed.

The network policy is intentionally container-scoped. Tunnel operation can replace the default route, rewrite `/etc/resolv.conf`, and flush/rebuild iptables rules. A deployment must provide `NET_ADMIN` and `/dev/net/tun` and must limit access to the unauthenticated management and proxy ports.

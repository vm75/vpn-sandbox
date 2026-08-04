<div align="center">
  <a href="https://github.com/vm75/vpn-sandbox">
    <img src="https://raw.githubusercontent.com/vm75/vpn-sandbox/main/vpn-sandbox.png" alt="Logo" width="24" height="24">
    <img src="https://raw.githubusercontent.com/vm75/vpn-sandbox/main/docs/title.svg" alt="Title">
  </a>
</div>

<div align="center">

[![License]](LICENSE) [![Build]][build_url] [![Version]][tag_url] [![Pulls]][hub_url]

</div>

VPN Sandbox is an open-source Alpine container that routes traffic through a configured OpenVPN or WireGuard tunnel. It provides an optional web UI, HTTP proxy, SOCKS5 proxy, DNS-leak and direct-LAN protections, and an optional `/data/apps.sh` hook for applications that should follow VPN state.

## Requirements

For a real tunnel, the runtime needs access to `/dev/net/tun` and the `NET_ADMIN` capability. The management API has no application authentication, so publish its port only to a trusted network. The same applies to the optional proxy ports.

## Run with Compose

Adjust host paths and ports in [`compose.yml.example`](compose.yml.example), then start it with a Compose-compatible tool:

```sh
podman compose -f compose.yml.example up -d
```

The example includes VPN Sandbox and a FlareSolverr sidecar sharing the VPN service network. Remove the sidecar and its extra ports if they are not needed. The example uses the published image `vm75/vpn-sandbox`; build locally with `make build` when developing.

The management UI is available at `http://<host>:9080` with the example mapping. The container listens on port `80`; HTTP proxy and SOCKS5 proxy ports are `3128` and `1080`.

## Run a local image

```sh
make build
podman run --rm --cap-add NET_ADMIN \
  --device /dev/net/tun \
  -p 8080:80 -p 1080:1080 -p 3128:3128 \
  -v "$PWD/test:/data" vm75/vpn-sandbox
```

`make run`, `make logs`, `make stop`, and `make clean` are shortcuts for the example stack. The project also supports Docker-compatible image and Compose tools; use their equivalent commands with [`Containerfile`](Containerfile) and `compose.yml.example`.

## Configure VPN and proxies

Open the web UI to create an OpenVPN or WireGuard server template, select the VPN type, and enable the desired module. Configuration is stored in SQLite. The 3proxy HTTP and SOCKS5 runtime configs are generated as `/data/var/3proxy-http.cfg` and `/data/var/3proxy-socks.cfg`; other generated runtime files are also kept under `/data/var`. VPN and proxy credentials and generated `.auth` files must remain private.

The persistent volume has this shape:

```text
/data/
├── config/vpn-sandbox.db
├── var/                  # generated configs, logs, PIDs, and credentials
└── apps.sh               # optional executable hook
```

If present, `apps.sh` receives `setup` once on first initialization, then `up` after VPN activation and `down` when it stops. The script runs inside the container and is responsible for installing or starting any custom applications it needs.

## Development and tests

```sh
make build
cd server && go test ./...
cd server && CGO_ENABLED=1 go build .
```

The manual network harness is container-sensitive and requires Docker plus the required privileges: `make test-start` and `make test-stop`. Do not run route, DNS, or firewall test commands directly on the host.

## Project documentation

- [`AGENTS.md`](AGENTS.md) — verified commands, repository rules, and safety boundaries.
- [`ARCHITECTURE.md`](ARCHITECTURE.md) — runtime components and event flow.
- [`DOCKERHUB.md`](DOCKERHUB.md) — published image tags, platforms, and container usage.
- [`CHANGELOG.md`](CHANGELOG.md) — user-visible release history.

## License

VPN Sandbox is licensed under the MIT License. See [`LICENSE`](LICENSE). Licenses for bundled OpenVPN, WireGuard, and 3proxy materials are in [`3rd-party/`](3rd-party/).

[license_url]: https://github.com/vm75/vpn-sandbox/blob/main/LICENSE
[build_url]: https://github.com/vm75/vpn-sandbox/actions
[tag_url]: https://hub.docker.com/r/vm75/vpn-sandbox/tags
[License]: https://img.shields.io/badge/license-MIT-blue.svg
[Build]: https://img.shields.io/github/actions/workflow/status/vm75/vpn-sandbox/.github/workflows/ci.yml?branch=main
[Version]: https://img.shields.io/docker/v/vm75/vpn-sandbox/latest?arch=amd64&sort=semver&color=066da5
[Pulls]: https://img.shields.io/docker/pulls/vm75/vpn-sandbox.svg?style=flat&label=pulls&logo=docker

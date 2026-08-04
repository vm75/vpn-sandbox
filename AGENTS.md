# VPN Sandbox agent guide

VPN Sandbox is a Go daemon and web UI packaged as an Alpine container. It routes container traffic through a configured OpenVPN or WireGuard tunnel and can run 3proxy HTTP and SOCKS5 proxies. Real tunnel operation changes routes, DNS, and iptables rules and requires `NET_ADMIN` plus `/dev/net/tun`.

## Repository map

- `server/main.go` — daemon startup, CLI flags, OpenVPN script mode, and signals.
- `server/core/` — shared paths, SQLite database, global configuration, and module registry.
- `server/actions/` — VPN up/down DNS, route, firewall, and app-script side effects.
- `server/modules/openvpn/`, `wireguard/` — tunnel configuration and lifecycle.
- `server/modules/proxy/` — 3proxy HTTP and SOCKS5 configuration and lifecycle.
- `server/webserver/` — REST/SSE API and static-file serving.
- `server/static/` — Vue 3 components loaded from CDN assets; `component-load.js` defines the SFC loading rules.
- `usr/local/etc/` — source templates for proxy configuration.
- `test/` — privileged/manual network test harness.
- `Containerfile`, `compose.yml.example`, and `Makefile` — image and local runtime workflows.
- `README.md` — installation, configuration, and user operation.
- `ARCHITECTURE.md` — component boundaries and runtime flow.
- `DOCKERHUB.md` — published image usage, tags, and release workflow.

## Verified commands

Run from the repository root unless noted otherwise.

```sh
make build                         # build the image with Podman
make run                           # start the example Compose stack
make logs                          # follow the vpn container logs
make stop                          # stop the stack
make clean                         # remove the stack
cd server && go test ./...         # all Go tests
cd server && go test ./utils -run '^TestArgParse$'
cd server && CGO_ENABLED=1 go build .
```

The Makefile currently uses Podman and Podman Compose. Docker-compatible tools can use the same `Containerfile` and `compose.yml.example` with their equivalent commands. The manual harness uses Docker and requires a privileged container namespace: `make test-start` and `make test-stop`.

The daemon accepts `--data`/`-d` (default `/data`), `--port`/`-p` (default `80`), `--test`, and `--sudo`. Go requires version `1.26.5` as declared in `server/go.mod`; CGO is required by SQLite.

## Boundaries and invariants

- Use `core.DataDir`, `ConfigDir`, `VarDir`, and `AppScript` for daemon paths. Persistent state is `/data/config/vpn-sandbox.db`; generated files, logs, PID files, and credentials are under `/data/var`.
- Initialize SQLite tables with `CREATE TABLE IF NOT EXISTS` and use the shared `core.Db`.
- Register modules through `core.RegisterModule` and implement `core.Module`.
- Preserve the event flow: VPN and global-configuration events are asynchronous; proxies follow VPN state.
- OpenVPN up/down hooks invoke this executable in script mode and signal the main process. WireGuard invokes the same VPN actions directly.
- Do not expose `.auth` files, private keys, tokens, databases, or generated runtime state in commits or API responses.
- The web/API routes have no application authentication. Published management and proxy ports must be restricted to trusted networks.
- Treat route, DNS, firewall, `pkill`, app-script, and tunnel changes as container-only operations. Do not run them against the host.
- Vue SFCs must use `export default { ... }`; keep component CSS in the trailing `<style>` block and follow the custom loader conventions.

## Working rules

Keep changes small and apply KISS/YAGNI. Search before broad reads, preserve existing conventions, and inspect only task-relevant files. Use `gofmt` for changed Go files; there is no frontend package manager or configured JS formatter. Add focused tests for behavior changes where practical, and use the isolated harness for network behavior.

Update only documentation whose truth changed. Review `README.md` for user-facing or deployment changes, `ARCHITECTURE.md` for component/data-flow changes, and this file for agent workflow, commands, or safety constraints. Review `CHANGELOG.md` only for a user-visible release entry.

## Definition of done

Review the final diff for unrelated edits, credentials, generated files, and broken Markdown paths. Run the smallest relevant validation and report skipped checks or remaining container/runtime risks explicitly.

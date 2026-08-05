## [0.5.0]
* Add SQLite configuration database backend.
* Handle SIGTERM signals gracefully and kill apps before tunnel drops.
* Enhance Web UI with app management, runtime file browser improvements, DNS leak display, and IP info refinements.

## [0.4.0]
* Migrate backend daemon architecture from Go to Python 3.11 (`python:3.11-slim-bookworm`).
* Sort runtime files alphabetically and auto-scroll content view on selection in WebUI file explorer.

## [0.3.0]
* Improve WebUI status stream recovery and IP info refresh consistency.

## [0.2.0]
* Upgrade the bundled 3proxy HTTP and SOCKS5 services to 0.9.7.
* Fix container builds failing because of a stale runtime template copy.

## [0.1.2]
* Replace Tinyproxy and Dante with 3proxy for HTTP and SOCKS5 proxy services.
* Refresh cached IP information when a manual force refresh is requested.

## [0.1.1]
* Improve VPN lifecycle handling, proxy integration, and WebUI configuration flows.
* Add container build/version metadata and refresh deployment documentation.

## [0.1.0]
* Updated all libraries and base image dependencies.
* Added Podman Compose shortcuts to Makefile (`run`, `clean`, `start`, `stop`, `logs`, `sh`).
* Documented rootless Podman and Docker support in README and AGENTS guide.

## [0.0.2]
* Fix Host Gateway detection.

## [0.0.1]
* First release. Supports OpenVPN and WireGuard. Includes WebUI and proxies.

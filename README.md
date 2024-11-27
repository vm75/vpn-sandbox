<div align="center">
  <a href="https://github.com/vm75/vpn-sandbox">
    <img src="https://raw.githubusercontent.com/vm75/vpn-sandbox/main/vpn-sandbox.png" alt="Logo" width="24" height="24" >
    <img src="https://raw.githubusercontent.com/vm75/vpn-sandbox/main/docs/title.svg" alt="Title">
  </a>
</div>
<div align="center">

[![License]](LICENSE)
[![Build]][build_url]
[![Version]][tag_url]
[![Size]][tag_url]
[![Pulls]][hub_url]
[![Package]][pkg_url]

</div>


**VPN Sandbox** is an open-source containerized solution for securely tunneling network traffic through a VPN. It supports **OpenVPN** and **WireGuard**, with features like **HTTP Proxy** and **SOCKS Proxy** support, DNS leak prevention, and a web-based interface for easy configuration. The container runs in **rootless mode** and is ideal for secure browsing or running custom applications behind a VPN.

<p align="center">
  <img src="https://raw.githubusercontent.com/vm75/vpn-sandbox/main/docs/screenshots.gif" alt="Screenshot" />
</p>

## Features

- **Supports OpenVPN and WireGuard**: Choose between two popular VPN protocols for your secure connection needs.
- **Rootless Container Support**: Run the container without elevated privileges (Docker/Podman/Kubernetes).
- **HTTP and SOCKS Proxy**: Redirect host network traffic through proxies to browse securely.
- **Web-Based Configuration UI**: Configure VPN servers and manage settings via an intuitive web interface.
- **Template-based Server Configuration**: Create and manage server configurations using templates.
- **Prevention of DNS Leaks and LAN Access**: Ensures that DNS queries do not leak and blocks direct LAN traffic for enhanced privacy.
- **Custom App Support**: Run custom Linux applications in the sandbox.

## Usage  🐳

Via Docker Compose:
```yaml
services:
  vpn-sandbox:
    image: vm75/vpn-sandbox
    container_name: vpn-sandbox
    cap_add:
      - NET_ADMIN
    devices:
      - /dev/net/tun
    ports:
      - "8080:80"   # Web UI
      - "1080:1080" # SOCKS Proxy
      - "3128:3128" # HTTP Proxy
    volumes:
      - /path/to/data:/data
    restart: unless-stopped
```

Via Docker CLI:
```bash
docker pull vm75/vpn-sandbox
docker run -d --name vpn-sandbox \
  --cap-add=NET_ADMIN \
  --device=/dev/net/tun \
  -v /path/to/data:/data \
  -p 8080:80 \
  vm75/vpn-sandbox
```

## Configuration ⚙️

To add a new server, use the web interface to create a new server configuration. It supports **OpenVPN** and **WireGuard** configurations.

The configuration templates can include custom parameters, such as endpoints, IP addresses, and ports. There can be multiple sets of values for each template. The parameters are enclosed in double brackets `{{}}`.

## Volume Structure

The `/data` volume should contain the following structure:
```plaintext
/data
├── config/         # Contains the sqlite3 database
├── var/            # Contains the runtime configuration and logs
├── apps.sh         # Custom apps script (optional)
```
It is recommended to place the `apps.sh` script in the `/data` volume.

### Example `apps.sh` Script (optional)
This script runs custom applications once the VPN connection is established:
```bash
#!/bin/sh

case "$1" in
  setup)
    apk --no-cache --no-progress <packages needed by apps>
    ;;
  up)
    # Run your custom apps here
    ;;
  down)
    # Stop your custom apps here
    ;;
esac
```

Ensure the script is executable:
```bash
chmod +x /data/apps.sh
```

## Web UI Access

The web UI is accessible at `http://<host-ip>:8080` by default. Use it to configure your VPN servers and settings with ease.

## Proxy Usage

Configure your browser or applications to use the container's HTTP (`3128`) or SOCKS (`1080`) proxies to securely route traffic through the VPN.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

### 3rd-Party Components

<table>
  <tr>
    <th>Component</th>
    <th>License</th>
  </tr>
  <tr>
    <td>
      <a href="https://openvpn.net/">OpenVPN</a>
    </td>
    <td>
      <a href="https://raw.githubusercontent.com/vm75/vpn-sandbox/main/3rd-party/openvpn/COPYRIGHT.GPL">COPYRIGHT.GPL</a>
    </td>
  </tr>
  <tr>
    <td>
      <a href="https://www.wireguard.com/">WireGuard</a>
    </td>
    <td>
      <a href="https://raw.githubusercontent.com/vm75/vpn-sandbox/main/3rd-party/wireguard/LICENSE">LICENSE</a>
    </td>
  </tr>
  <tr>
    <td>
      <a href="https://www.inet.no/dante/">Dante (Socks Proxy)</a>
    </td>
    <td>
      <a href="https://raw.githubusercontent.com/vm75/vpn-sandbox/main/3rd-party/dante/LICENSE">LICENSE</a>
    </td>
  </tr>
  <tr>
    <td>
      <a href="https://tinyproxy.github.io/">Tinyproxy (HTTP Proxy)</a>
    </td>
    <td>
      <a href="https://raw.githubusercontent.com/vm75/vpn-sandbox/main/3rd-party/tinyproxy/COPYING">COPYING</a>
    </td>
  </tr>
</table>

---

**VPN Sandbox** provides a simple, secure, and flexible way to manage VPN connections using containerization. Contributions are welcome! 🚀

[license_url]: https://github.com/vm75/vpn-sandbox/blob/main/LICENSE
[build_url]: https://github.com/vm75/vpn-sandbox/actions
[hub_url]: https://hub.docker.com/r/vm75/vpn-sandbox
[tag_url]: https://hub.docker.com/r/vm75/vpn-sandbox/tags
[pkg_url]: https://github.com/vm75/vpn-sandbox/pkgs/container/vpn-sandbox
[screenshot_url]: https://raw.githubusercontent.com/vm75/vpn-sandbox/main/docs/screenshot.gif

[License]: https://img.shields.io/badge/license-MIT-blue.svg
[Build]: https://img.shields.io/github/actions/workflow/status/vm75/vpn-sandbox/.github/workflows/ci.yml?branch=main
[Version]: https://img.shields.io/docker/v/vm75/vpn-sandbox/latest?arch=amd64&sort=semver&color=066da5
[Size]: https://img.shields.io/docker/image-size/vm75/vpn-sandbox/latest?color=066da5&label=size
[Package]: https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fipitio.github.io%2Fbackage%2Fvm75%2Fvpn-sandbox%2Fvpn-sandbox.json&query=%24.downloads&logo=github&style=flat&color=066da5&label=pulls
[Pulls]: https://img.shields.io/docker/pulls/vm75/vpn-sandbox.svg?style=flat&label=pulls&logo=docker
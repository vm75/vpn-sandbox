package wireguard

import (
	"errors"
	"net"
	"os"
	"strings"
	"vpn-sandbox/actions"
	"vpn-sandbox/utils"
)

func isTunnelUp() bool {
	if _, err := os.Stat("/sys/class/net/wg0"); os.IsNotExist(err) {
		return false
	}

	out, err := utils.RunCommand(utils.UseSudo, "/usr/bin/wg", "show", "wg0")

	if strings.Contains(out, "peer: ") && err == nil {
		return true
	}

	return false
}

// Function to find the line that starts with the given key and return the value after the first '='
func findValue(context, key, def string) string {
	lines := strings.Split(context, "\n") // Split the context into lines
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line) // Trim any leading/trailing whitespace
		if strings.HasPrefix(trimmedLine, key) {
			parts := strings.SplitN(trimmedLine, "=", 2) // Split only at the first '='
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return def
}

func getAddress(endpoint string) string {
	ips, err := net.LookupIP(strings.Split(endpoint, ":")[0])

	if err != nil {
		utils.LogLn(err)
		return ""
	}

	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.String()
		}
	}
	return ""
}

func tunnelUp() error {
	if isTunnelUp() || !wireguardConfig.Enabled {
		return nil
	}

	var server = getWireguardServer(wireguardConfig.ServerName)
	if server == nil {
		return errors.New("server not found")
	}
	var endPoint map[string]string = nil

	for _, entry := range server.Endpoints {
		if entry["name"] == wireguardConfig.ServerEndpoint {
			endPoint = entry
			break
		}
	}

	wgConfig := server.Template
	for key, value := range endPoint {
		wgConfig = strings.ReplaceAll(wgConfig, "{{"+key+"}}", value)
	}

	// var hostGateway string
	privateKey := findValue(wgConfig, "PrivateKey", "")
	peerPublicKey := findValue(wgConfig, "PublicKey", "")
	endpoint := findValue(wgConfig, "Endpoint", "")
	DNS := findValue(wgConfig, "DNS", "1.1.1.1, 1.0.0.1")
	vpnAddress := getAddress(endpoint)
	address := findValue(wgConfig, "Address", vpnAddress)
	allowedIps := findValue(wgConfig, "AllowedIPs", "0.0.0.0/0")

	utils.RunCommand(utils.UseSudo, "/sbin/ip", "link", "add", "dev", "wg0", "type", "wireguard")

	// save private key to /tmp/wg0.key
	os.WriteFile("/tmp/wg0.key", []byte(privateKey), 0644)

	utils.RunCommand(utils.UseSudo,
		"/usr/bin/wg", "set", "wg0",
		"private-key", "/tmp/wg0.key",
		"peer", peerPublicKey,
		"endpoint", endpoint,
		"allowed-ips", allowedIps)

	// clean up private key
	os.Remove("/tmp/wg0.key")

	utils.RunCommand(utils.UseSudo, "/sbin/ip", "address", "add", address, "dev", "wg0")
	utils.RunCommand(utils.UseSudo, "/sbin/ip", "link", "set", "up", "dev", "wg0")

	if !isTunnelUp() {
		utils.LogLn("Tunnel up failed")
		return nil
	}

	go actions.VpnUp(&actions.NetSpec{
		Dev:         "wg0",
		Domains:     []string{},
		DNS:         strings.Fields(strings.ReplaceAll(DNS, ",", " ")),
		VPNGateway:  "",
		VpnEndpoint: vpnAddress,
	})

	return nil
}

func tunnelDown() error {
	utils.RunCommand(utils.UseSudo, "/sbin/ip", "link", "set", "down", "dev", "wg0")
	utils.RunCommand(utils.UseSudo, "/sbin/ip", "link", "del", "dev", "wg0")

	if isTunnelUp() {
		utils.LogLn("Tunnel down failed")
	}

	actions.VpnDown()

	return nil
}

func Shutdown() {
	tunnelDown()
}

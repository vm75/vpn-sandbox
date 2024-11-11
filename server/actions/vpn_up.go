package actions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

type NetSpec struct {
	Dev        string   `json:"dev"`
	Domains    []string `json:"domains"`
	DNS        []string `json:"dns"`
	VPNGateway string   `json:"vpn_gateway"`
}

func SaveOpenVPNSpec() {
	specFile := filepath.Join(core.VarDir, "openvpn.spec")

	netSpec := &NetSpec{
		Dev:        os.Getenv("dev"),
		Domains:    []string{},
		DNS:        []string{},
		VPNGateway: os.Getenv("route_vpn_gateway"),
	}

	// iterate over foreign_option_N env vars
	i := 1
	for ; os.Getenv(fmt.Sprintf("foreign_option_%d", i)) != ""; i++ {
		fopt := os.Getenv(fmt.Sprintf("foreign_option_%d", i))
		if fopt == "" {
			break
		}

		if strings.HasPrefix(fopt, "dhcp-option DOMAIN ") {
			netSpec.Domains = append(netSpec.Domains, fopt[len("dhcp-option DOMAIN "):])
			continue
		}
		if strings.HasPrefix(fopt, "dhcp-option DNS ") {
			netSpec.DNS = append(netSpec.DNS, fopt[len("dhcp-option DNS "):])
			continue
		}
	}

	data, _ := json.MarshalIndent(netSpec, "", "  ")
	os.WriteFile(specFile, data, 0644)
}

func RetrieveOpenVPNSpec() (*NetSpec, error) {
	specFile := filepath.Join(core.VarDir, "openvpn.spec")
	data, err := os.ReadFile(specFile)
	if err != nil {
		return nil, err
	}
	netSpec := &NetSpec{}
	json.Unmarshal(data, netSpec)
	return netSpec, nil
}

func VpnUp(netSpec *NetSpec) {
	utils.LogLn("vpn up")

	if netSpec == nil {
		var err error
		netSpec, err = RetrieveOpenVPNSpec()
		if netSpec == nil {
			utils.LogError("No openvpn spec found", err)
			return
		}
	}

	var sb strings.Builder
	if len(netSpec.Domains) == 1 {
		sb.WriteString(fmt.Sprintf("domain %s\n", netSpec.Domains[0]))
	} else if len(netSpec.Domains) > 1 {
		sb.WriteString(fmt.Sprintf("search %s\n", strings.Join(netSpec.Domains, " ")))
	}
	for _, nameserver := range netSpec.DNS {
		sb.WriteString(fmt.Sprintf("nameserver %s\n", nameserver))
	}
	// write resolv.conf
	if err := os.WriteFile("/etc/resolv.conf", []byte(sb.String()), 0644); err != nil {
		utils.LogError("Error updating /etc/resolv.conf", err)
	}

	// Set routes
	// Remove all existing default routes
	utils.RunCommand("/sbin/ip", "route", "del", "default")

	// Default route for all traffic through the VPN tunnel
	utils.RunCommand("/sbin/ip", "route", "add", "default", "via", netSpec.VPNGateway, "dev", netSpec.Dev)

	// Set firewall rules
	// Flush existing rules to start fresh
	utils.RunCommand("/sbin/iptables", "-F")

	// Allow incoming ESTABLISHED and RELATED connections on the VPN interface
	utils.RunCommand("/sbin/iptables", "-A", "INPUT", "-i", netSpec.Dev, "-m", "conntrack", "--ctstate", "ESTABLISHED,RELATED", "-j", "ACCEPT")

	// Drop all other incoming connections on the VPN interface
	utils.RunCommand("/sbin/iptables", "-A", "INPUT", "-i", netSpec.Dev, "-j", "DROP")

	// Trigger vpn-up actions
	utils.RunCommand(core.AppScript, "up")
}

package actions

import (
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

func VpnDown() {
	utils.LogLn("vpn down")

	// restore resolv.conf
	utils.RestoreResolvConf()

	// get host gateway from resolv.conf
	hostGateway := utils.GetHostGateway()
	utils.LogLn("host gateway: " + hostGateway)

	if core.Testing {
		utils.PublishEvent(utils.Event{Name: "vpn-down", Context: map[string]interface{}{}})
		utils.LogLn("Skipping vpn down actions for testing")
		return
	}

	// Set routes
	// Remove all existing default routes
	utils.RunCommand("/sbin/ip", "route", "del", "default")

	// Add default gateway
	utils.RunCommand("/sbin/ip", "route", "add", "default", "via", hostGateway, "dev", "eth0")

	// Set firewall rules
	// Flush existing rules to start fresh
	utils.RunCommand("/sbin/iptables", "-F")

	// Allow related and established connections (for existing sessions to work)
	utils.RunCommand("/sbin/iptables", "-A", "INPUT", "-m", "conntrack", "--ctstate", "RELATED,ESTABLISHED", "-j", "ACCEPT")

	// Allow incoming connections only on port 80
	utils.RunCommand("/sbin/iptables", "-A", "INPUT", "-p", "tcp", "--dport", "80", "-j", "ACCEPT")

	// Drop all other incoming connections
	utils.RunCommand("/sbin/iptables", "-A", "INPUT", "-j", "DROP")

	// Trigger vpn down actions
	utils.PublishEvent(utils.Event{Name: "vpn-down", Context: map[string]interface{}{}})
	utils.RunCommand(core.AppScript, "down")
}

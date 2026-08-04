package proxy

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

func quoteConfigValue(value string) (string, error) {
	if strings.ContainsAny(value, "\r\n") {
		return "", errors.New("3proxy config values cannot contain newlines")
	}
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`, nil
}

func renderProxyConfig(proxyType ProxyType, listenAddr, bindAddr, logFile, username, password string) (string, error) {
	if listenAddr == "" || bindAddr == "" {
		return "", errors.New("listen or bind address not found")
	}

	quotedLogFile, err := quoteConfigValue(logFile)
	if err != nil {
		return "", err
	}

	lines := []string{
		"log " + quotedLogFile,
		"timeouts 1 5 30 60 180 1800 15 60 15 5",
		"maxconn 100",
		"internal " + listenAddr,
		"external " + bindAddr,
	}

	if username != "" && password != "" {
		credentials, err := quoteConfigValue(username + ":CL:" + password)
		if err != nil {
			return "", err
		}
		lines = append(lines, "users "+credentials, "auth strong", "allow *")
	} else {
		lines = append(lines, "auth none")
	}

	switch proxyType {
	case HttpProxy:
		lines = append(lines, "proxy -p3128")
	case SocksProxy:
		lines = append(lines, "socks -p1080")
	default:
		return "", fmt.Errorf("unknown proxy type %d", proxyType)
	}

	return strings.Join(lines, "\n") + "\n", nil
}

func updateProxyConfig(p *ProxyModule) error {
	vpnDev := core.GetVpnDevice()
	if vpnDev == "" {
		return errors.New("VPN device not found")
	}

	content, err := renderProxyConfig(
		p.proxyType,
		utils.GetIpV4Addr("eth0", true),
		utils.GetIpV4Addr(vpnDev, true),
		strings.TrimSuffix(p.configFile, ".cfg")+".log",
		core.GlobalConfig.ProxyUsername,
		core.GlobalConfig.ProxyPassword,
	)
	if err != nil {
		return err
	}

	return os.WriteFile(p.configFile, []byte(content), 0600)
}

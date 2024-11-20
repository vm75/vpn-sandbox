package http_proxy

import (
	"errors"
	"os"
	"regexp"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

func updateRuntimeConfig() error {
	vpnDev := core.GetVpnDevice()
	if vpnDev == "" {
		return errors.New("VPN device not found")
	}

	content, err := os.ReadFile("/usr/local/etc/tinyproxy.conf")
	if err != nil {
		return err
	}

	contentStr := string(content)

	listenRegex := regexp.MustCompile(`Listen.*`)
	bindRegex := regexp.MustCompile(`Bind.*`)

	listenAddr := utils.GetIpV4Addr("eth0", true)
	bindAddr := utils.GetIpV4Addr(vpnDev, true)

	if listenAddr == "" || bindAddr == "" {
		return errors.New("listen or bind not found")
	}

	contentStr = listenRegex.ReplaceAllString(contentStr, "Listen "+listenAddr)
	contentStr = bindRegex.ReplaceAllString(contentStr, "Bind "+bindAddr)

	if core.GlobalConfig.ProxyUsername != "" && core.GlobalConfig.ProxyPassword != "" {
		contentStr = contentStr +
			"\nBasicAuth " + core.GlobalConfig.ProxyUsername +
			" " + core.GlobalConfig.ProxyPassword
	}

	err = os.WriteFile(configFile, []byte(contentStr), 0644)

	if err != nil {
		return err
	}

	return nil
}

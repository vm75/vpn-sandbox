package socks_proxy

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

func updateRuntimeConfig() error {
	vpnDev := core.GetVpnDevice()
	if vpnDev == "" {
		return errors.New("VPN device not found")
	}

	content, err := os.ReadFile("/usr/local/etc/sockd.conf")
	if err != nil {
		return err
	}

	contentStr := string(content)

	externalRegex := regexp.MustCompile(`external: .*`)
	contentStr = externalRegex.ReplaceAllString(contentStr, "external: "+vpnDev)

	if core.GlobalConfig.ProxyUsername != "" && core.GlobalConfig.ProxyPassword != "" {
		utils.CreateUser(core.GlobalConfig.ProxyUsername)
		contentStr = strings.Replace(contentStr, "socksmethod: none", "socksmethod: username", 1)
	}

	err = os.WriteFile(configFile, []byte(contentStr), 0644)
	if err != nil {
		return err
	}

	return nil
}

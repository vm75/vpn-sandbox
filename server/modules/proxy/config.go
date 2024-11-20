package proxy

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

func updateHttpProxyConfig(p *ProxyModule) error {
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

	err = os.WriteFile(p.configFile, []byte(contentStr), 0644)

	if err != nil {
		return err
	}

	return nil
}

func updateSocksProxyConfig(p *ProxyModule) error {
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

	err = os.WriteFile(p.configFile, []byte(contentStr), 0644)
	if err != nil {
		return err
	}

	return nil
}

func updateProxyConfig(p *ProxyModule) error {
	if p.proxyType == HttpProxy {
		return updateHttpProxyConfig(p)
	} else if p.proxyType == SocksProxy {
		return updateSocksProxyConfig(p)
	}
	return nil
}

package http_proxy

import (
	"errors"
	"os"
	"regexp"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

func updateRuntimeConfig() error {
	content, err := os.ReadFile("/usr/local/etc/tinyproxy.conf")

	if err != nil {
		return err
	}

	contentStr := string(content)

	listenRegex := regexp.MustCompile(`Listen.*`)
	bindRegex := regexp.MustCompile(`Bind.*`)

	listenAddr := utils.GetIpV4Addr("eth0", true)
	bindAddr := utils.GetIpV4Addr("tun0", true)

	if listenAddr == "" || bindAddr == "" {
		return errors.New("listen or bind not found")
	}

	contentStr = listenRegex.ReplaceAllString(contentStr, "Listen "+listenAddr)
	contentStr = bindRegex.ReplaceAllString(contentStr, "Bind "+bindAddr)

	if core.GlobalConfig.ProxyUsername != "" && core.GlobalConfig.ProxyPassword != "" {
		// append "\nBasicAuth ${PROXY_USERNAME} ${PROXY_PASSWORD}"
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

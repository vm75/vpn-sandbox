package proxy

import (
	"os"
	"path/filepath"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

type ProxyType int

const (
	HttpProxy ProxyType = iota + 1
	SocksProxy
)

type ProxyModule struct {
	core.DefaultModule
	proxyType   ProxyType
	displayName string
	execName    string
	proxyCmd    []string
	configFile  string
	pidFile     string
}

func InitModule(proxyType ProxyType) {
	var module ProxyModule
	switch proxyType {
	case HttpProxy:
		configFile := filepath.Join(core.VarDir, "tinyproxy.conf")
		module = ProxyModule{
			DefaultModule: core.DefaultModule{
				Name: "http_proxy",
			},
			proxyType:   HttpProxy,
			displayName: "HTTP Proxy",
			execName:    "tinyproxy",
			proxyCmd:    []string{"/usr/bin/tinyproxy", "-d", "-c", configFile},
			configFile:  configFile,
			pidFile:     filepath.Join(core.VarDir, "tinyproxy.pid"),
		}
	case SocksProxy:
		configFile := filepath.Join(core.VarDir, "sockd.conf")
		module = ProxyModule{
			DefaultModule: core.DefaultModule{
				Name: "socks_proxy",
			},
			proxyType:   SocksProxy,
			displayName: "SOCKS Proxy",
			execName:    "sockd",
			proxyCmd:    []string{"/usr/local/sbin/sockd", "-f", configFile},
			configFile:  configFile,
			pidFile:     filepath.Join(core.VarDir, "sockd.pid"),
		}
	}

	module.LoadConfig()

	core.RegisterModule(module.Name, &module)
	utils.RegisterListener("global-config-changed", &module)
	utils.AddSignalHandler([]os.Signal{core.VPN_UP}, func(_ os.Signal) {
		if module.Config["enabled"].(bool) {
			go startProxy(&module)
		}
	})
	utils.AddSignalHandler([]os.Signal{core.VPN_DOWN}, func(_ os.Signal) {
		stopProxy(&module)
	})
}

// IsRunning implements core.Module.
func (p *ProxyModule) IsRunning() bool {
	return utils.IsRunning(proxyCmd)
}

func (p *ProxyModule) Enable(startNow bool) error {
	err := p.DefaultModule.Enable(startNow)

	if err == nil && core.IsVpnUp() && startNow {
		go startProxy(p)
	}

	return err
}

func (p *ProxyModule) Disable(stopNow bool) error {
	err := p.DefaultModule.Disable(stopNow)

	if err == nil && stopNow {
		stopProxy(p)
	}

	return err
}

// HandleEvent implements utils.EventListener.
func (p *ProxyModule) HandleEvent(event utils.Event) {
	switch event.Name {
	case "global-config-changed":
		updateProxyConfig(p)
	}
}

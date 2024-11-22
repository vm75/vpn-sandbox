package proxy

import (
	"os/exec"
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
	cmdObject   *exec.Cmd
	isRunning   bool
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
			cmdObject:   nil,
			isRunning:   false,
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
			proxyCmd:    []string{"/usr/bin/sockd", "-f", configFile},
			configFile:  configFile,
			pidFile:     filepath.Join(core.VarDir, "sockd.pid"),
			cmdObject:   nil,
			isRunning:   false,
		}
	}

	module.LoadConfig()

	core.RegisterModule(module.Name, &module)
	utils.RegisterListener([]string{"global-config-changed", "vpn-up", "vpn-down"}, &module)
}

// IsRunning implements core.Module.
func (p *ProxyModule) IsRunning() bool {
	return utils.IsRunning(p.cmdObject)
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
		stopProxy(p)
		go startProxy(p)
	case "vpn-up":
		stopProxy(p)
		go startProxy(p)
	case "vpn-down":
		stopProxy(p)
	}
}

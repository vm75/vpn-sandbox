package socks_proxy

import (
	"os"
	"path/filepath"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

type HttpProxyModule struct {
	core.DefaultModule
}

var pidFile = ""
var configFile = ""

func InitModule() {
	var module = HttpProxyModule{
		DefaultModule: core.DefaultModule{
			Name: "socks_proxy",
		},
	}

	module.LoadConfig()

	core.RegisterModule(module.Name, &module)
	utils.RegisterListener("global-config-changed", &module)
	utils.AddSignalHandler([]os.Signal{core.VPN_UP}, func(_ os.Signal) {
		if module.Config["enabled"].(bool) {
			go startProxy()
		}
	})
	utils.AddSignalHandler([]os.Signal{core.VPN_DOWN}, func(_ os.Signal) {
		go stopProxy()
	})

	configFile = filepath.Join(core.VarDir, "sockd.conf")
	pidFile = filepath.Join(core.VarDir, "sockd.pid")
}

// GetStatus implements core.Module.
func (h *HttpProxyModule) GetStatus() (core.ModuleStatus, error) {
	return core.ModuleStatus{Running: utils.IsRunning(proxyCmd)}, nil
}

// HandleEvent implements utils.EventListener.
func (h *HttpProxyModule) HandleEvent(event utils.Event) {
	switch event.Name {
	case "global-config-changed":
		updateRuntimeConfig()
	}
}

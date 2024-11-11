package http_proxy

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
			Name: "http_proxy",
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
	utils.RegisterListener("vpn-down", &module)

	configFile = filepath.Join(core.VarDir, "tinyproxy.conf")
	pidFile = filepath.Join(core.VarDir, "tinyproxy.pid")
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

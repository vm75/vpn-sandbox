package socks_proxy

import (
	"os"
	"path/filepath"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

const ModuleName = "socks_proxy"

type HttpProxyModule struct {
	core.DefaultModule
}

var pidFile = ""
var configFile = ""

func InitModule() {
	var module = HttpProxyModule{
		DefaultModule: core.DefaultModule{
			Name: ModuleName,
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
		stopProxy()
	})

	configFile = filepath.Join(core.VarDir, "sockd.conf")
	pidFile = filepath.Join(core.VarDir, "sockd.pid")
}

// IsRunning implements core.Module.
func (h *HttpProxyModule) IsRunning() bool {
	return utils.IsRunning(proxyCmd)
}

func (h *HttpProxyModule) Enable(startNow bool) error {
	err := h.DefaultModule.Enable(startNow)

	if err == nil && core.IsVpnUp() && startNow {
		go startProxy()
	}

	return err
}

func (h *HttpProxyModule) Disable(stopNow bool) error {
	err := h.DefaultModule.Disable(stopNow)

	if err == nil && stopNow {
		stopProxy()
	}

	return err
}

// HandleEvent implements utils.EventListener.
func (h *HttpProxyModule) HandleEvent(event utils.Event) {
	switch event.Name {
	case "global-config-changed":
		updateRuntimeConfig()
	}
}

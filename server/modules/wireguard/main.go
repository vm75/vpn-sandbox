package wireguard

import (
	"vpn-sandbox/core"
	"vpn-sandbox/utils"

	"github.com/gorilla/mux"
)

type WireguardModule struct {
	Enabled    bool   `json:"enabled"`
	ServerName string `json:"serverName"`
}

var wireguardConfig = WireguardModule{
	Enabled:    false,
	ServerName: "",
}

func InitModule() {
	initDb()

	savedConfig, err := core.GetConfig("wireguard")
	if err == nil {
		utils.MapToObject(savedConfig, &wireguardConfig)
	} else {
		utils.ObjectToMap(wireguardConfig, &savedConfig)
		core.SaveConfig("wireguard", savedConfig)
	}

	core.RegisterModule("wireguard", &wireguardConfig)

	if wireguardConfig.Enabled {
		tunnelUp()
	}
}

// RegisterRoutes implements core.Module.
func (w *WireguardModule) RegisterRoutes(r *mux.Router) {
	// Template-related routes
	r.HandleFunc("/api/wireguard/servers", listServers).Methods("GET")
	r.HandleFunc("/api/wireguard/servers/{name}", getServer).Methods("GET")
	r.HandleFunc("/api/wireguard/servers/save", saveServer).Methods("POST")
	r.HandleFunc("/api/wireguard/servers/delete/{name}", deleteServer).Methods("DELETE")
}

// GetStatus implements core.Module.
func (w *WireguardModule) GetStatus() (core.ModuleStatus, error) {
	if isTunnelUp() {
		return core.ModuleStatus{Running: true}, nil
	}
	return core.ModuleStatus{Running: false}, nil
}

// Enable implements core.Module.
func (w *WireguardModule) Enable(startNow bool) error {
	w.Enabled = true
	config := map[string]interface{}{}
	utils.ObjectToMap(w, &config)
	core.SaveConfig("wireguard", config)
	if startNow {
		tunnelUp()
	}
	return nil
}

// Disable implements core.Module.
func (w *WireguardModule) Disable(stopNow bool) error {
	w.Enabled = false
	config := map[string]interface{}{}
	utils.ObjectToMap(w, &config)
	core.SaveConfig("wireguard", config)
	if stopNow {
		tunnelDown()
	}
	return nil
}

// Start implements core.Module.
func (w *WireguardModule) Start() error {
	if w.Enabled {
		tunnelUp()
	} else {
		w.Enable(true)
	}
	return nil
}

// Stop implements core.Module.
func (w *WireguardModule) Stop() error {
	if w.Enabled {
		w.Disable(true)
	} else {
		tunnelDown()
	}
	return nil
}

// Restart implements core.Module.
func (w *WireguardModule) Restart() error {
	tunnelDown()
	tunnelUp()
	return nil
}

// GetConfig implements core.Module.
func (w *WireguardModule) GetConfig(params map[string]string) (map[string]interface{}, error) {
	var config map[string]interface{}
	utils.ObjectToMap(wireguardConfig, &config)
	return config, nil
}

// SaveConfig implements core.Module.
func (w *WireguardModule) SaveConfig(params map[string]string, config map[string]interface{}) error {
	if !utils.HasChanged(w, config) {
		return nil
	}
	utils.MapToObject(config, w)
	err := core.SaveConfig("wireguard", config)
	if err != nil {
		return err
	}

	w.Restart()

	return nil
}

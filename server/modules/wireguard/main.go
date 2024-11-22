package wireguard

import (
	"vpn-sandbox/core"
	"vpn-sandbox/utils"

	"github.com/gorilla/mux"
)

const ModuleName = "wireguard"

type WireguardModule struct {
	Enabled        bool   `json:"enabled"`
	ServerName     string `json:"serverName"`
	ServerEndpoint string `json:"serverEndpoint"`
}

var wireguardConfig = WireguardModule{
	Enabled:        false,
	ServerName:     "",
	ServerEndpoint: "",
}

func InitModule() {
	initDb()

	savedConfig, err := core.GetConfig(ModuleName)
	if err == nil {
		utils.MapToObject(savedConfig, &wireguardConfig)
	} else {
		utils.ObjectToMap(wireguardConfig, &savedConfig)
		core.SaveConfig(ModuleName, savedConfig)
	}

	core.RegisterModule(ModuleName, &wireguardConfig)

	if wireguardConfig.Enabled {
		tunnelUp()
	}
}

// RegisterRoutes implements core.Module.
func (w *WireguardModule) RegisterRoutes(r *mux.Router) {
	// Template-related routes
	r.HandleFunc("/api/wireguard/servers", listServersHandler).Methods("GET")
	r.HandleFunc("/api/wireguard/servers/{name}", getServerHandler).Methods("GET")
	r.HandleFunc("/api/wireguard/servers/save", saveServerHandler).Methods("POST")
	r.HandleFunc("/api/wireguard/servers/delete/{name}", deleteServerHandler).Methods("DELETE")
}

// IsRunning implements core.Module.
func (w *WireguardModule) IsRunning() bool {
	return isTunnelUp()
}

// Enable implements core.Module.
func (w *WireguardModule) Enable(startNow bool) error {
	w.Enabled = true
	config := map[string]interface{}{}
	utils.ObjectToMap(w, &config)
	core.SaveConfig(ModuleName, config)
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
	core.SaveConfig(ModuleName, config)
	if stopNow {
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
	config["servers"] = getWireguardServers()
	return config, nil
}

// SaveConfig implements core.Module.
func (w *WireguardModule) SaveConfig(params map[string]string, config map[string]interface{}) error {
	if !utils.HasChanged(w, config) {
		return nil
	}
	utils.MapToObject(config, w)
	err := core.SaveConfig(ModuleName, config)
	if err != nil {
		return err
	}

	w.Restart()

	return nil
}

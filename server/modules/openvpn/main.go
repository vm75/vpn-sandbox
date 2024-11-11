package openvpn

import (
	"os"
	"path/filepath"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"

	"github.com/gorilla/mux"
)

type OpenVPNModule struct {
	Enabled        bool   `json:"enabled"`
	ServerName     string `json:"serverName"`
	ServerEndpoint string `json:"serverEndpoint"`
	LogLevel       int    `json:"logLevel"`
	RetryInterval  int    `json:"retryInterval"`
}

var openvpnConfig = OpenVPNModule{
	Enabled:        false,
	ServerName:     "",
	ServerEndpoint: "",
	LogLevel:       0,
	RetryInterval:  3600,
}
var configFile = ""
var authFile = ""
var pidFile = ""
var logFile = ""
var statusFile = ""

func InitModule() {
	initDb()

	configFile = filepath.Join(core.VarDir, "vpn.ovpn")
	authFile = filepath.Join(core.VarDir, "vpn.auth")
	pidFile = filepath.Join(core.VarDir, "openvpn.pid")
	logFile = filepath.Join(core.VarDir, "openvpn.log")
	statusFile = filepath.Join(core.VarDir, "openvpn.status")

	savedConfig, err := core.GetConfig("openvpn")
	if err == nil {
		utils.MapToObject(savedConfig, &openvpnConfig)
	} else {
		utils.ObjectToMap(openvpnConfig, &savedConfig)
		core.SaveConfig("openvpn", savedConfig)
	}

	core.RegisterModule("openvpn", &openvpnConfig)

	utils.AddSignalHandler([]os.Signal{core.SHUTDOWN}, func(_ os.Signal) {
		openvpnConfig.Enabled = false
		killOpenVPN()
		if utils.IsRunning(openvpnCmd) {
			openvpnCmd.Wait()
		}
		os.Exit(0)
	})

	if openvpnConfig.Enabled {
		go runOpenVPN()
	}
}

// RegisterRoutes implements core.Module.
func (o *OpenVPNModule) RegisterRoutes(r *mux.Router) {
	// Template-related routes
	r.HandleFunc("/api/openvpn/servers", listServers).Methods("GET")
	r.HandleFunc("/api/openvpn/servers/{name}", getServer).Methods("GET")
	r.HandleFunc("/api/openvpn/servers/save", saveServer).Methods("POST")
	r.HandleFunc("/api/openvpn/servers/delete/{name}", deleteServer).Methods("DELETE")
}

// GetStatus implements core.Module.
func (o *OpenVPNModule) GetStatus() (core.ModuleStatus, error) {
	if utils.IsRunning(openvpnCmd) {
		return core.ModuleStatus{Running: true}, nil
	}
	return core.ModuleStatus{Running: false}, nil
}

// Enable implements core.Module.
func (o *OpenVPNModule) Enable(startNow bool) error {
	o.Enabled = true
	config := map[string]interface{}{}
	utils.ObjectToMap(o, &config)
	core.SaveConfig("openvpn", config)
	if startNow {
		go runOpenVPN()
	}
	return nil
}

// Disable implements core.Module.
func (o *OpenVPNModule) Disable(stopNow bool) error {
	o.Enabled = false
	config := map[string]interface{}{}
	utils.ObjectToMap(o, &config)
	core.SaveConfig("openvpn", config)
	if stopNow {
		killOpenVPN()
	}
	return nil
}

// Start implements core.Module.
func (o *OpenVPNModule) Start() error {
	if o.Enabled {
		go runOpenVPN()
	} else {
		o.Enable(true)
	}
	return nil
}

// Stop implements core.Module.
func (o *OpenVPNModule) Stop() error {
	if o.Enabled {
		o.Disable(true)
	} else {
		killOpenVPN()
	}
	return nil
}

// Restart implements core.Module.
func (o *OpenVPNModule) Restart() error {
	killOpenVPN()
	return nil
}

// GetConfig implements core.Module.
func (o *OpenVPNModule) GetConfig(params map[string]string) (map[string]interface{}, error) {
	var config map[string]interface{}
	utils.ObjectToMap(openvpnConfig, &config)
	return config, nil
}

// SaveConfig implements core.Module.
func (o *OpenVPNModule) SaveConfig(params map[string]string, config map[string]interface{}) error {
	if !utils.HasChanged(o, config) {
		return nil
	}
	utils.MapToObject(config, o)
	err := core.SaveConfig("openvpn", config)
	if err != nil {
		return err
	}
	saveOvpnConfig()

	killOpenVPN()
	go runOpenVPN()

	return nil
}

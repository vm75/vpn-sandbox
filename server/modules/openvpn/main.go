package openvpn

import (
	"os"
	"path/filepath"
	"strings"
	"time"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"

	"github.com/gorilla/mux"
)

const ModuleName = "openvpn"

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
	RetryInterval:  300,
}
var configFile = ""
var authFile = ""
var pidFile = ""
var logFile = ""
var statusFile = ""
var statusUpdateInterval = 60 * time.Second

func InitModule() {
	initDb()

	configFile = filepath.Join(core.VarDir, "vpn.ovpn")
	authFile = filepath.Join(core.VarDir, "vpn.auth")
	pidFile = filepath.Join(core.VarDir, "openvpn.pid")
	logFile = filepath.Join(core.VarDir, "openvpn.log")
	statusFile = filepath.Join(core.VarDir, "openvpn.status")

	savedConfig, err := core.GetConfig(ModuleName)
	if err == nil {
		utils.MapToObject(savedConfig, &openvpnConfig)
	} else {
		utils.ObjectToMap(openvpnConfig, &savedConfig)
		core.SaveConfig(ModuleName, savedConfig)
	}

	core.RegisterModule(ModuleName, &openvpnConfig)

	if openvpnConfig.Enabled {
		go runOpenVPN()
	}
}

// RegisterRoutes implements core.Module.
func (o *OpenVPNModule) RegisterRoutes(r *mux.Router) {
	// Template-related routes
	r.HandleFunc("/api/openvpn/servers", listServersHandler).Methods("GET")
	r.HandleFunc("/api/openvpn/servers/{name}", getServerHandler).Methods("GET")
	r.HandleFunc("/api/openvpn/servers/save", saveServerHandler).Methods("POST")
	r.HandleFunc("/api/openvpn/servers/delete/{name}", deleteServerHandler).Methods("DELETE")
}

// IsRunning implements core.Module.
func (o *OpenVPNModule) IsRunning() bool {
	if _, err := os.Stat("/sys/class/net/tun0"); os.IsNotExist(err) {
		return false
	}

	out, err := utils.RunCommand(utils.UseSudo, "/sbin/ip", "a", "show", "dev", "tun0")
	if err != nil || strings.Contains(out, "state DOWN") {
		return false
	}

	// Read the statusfile
	content, err := os.ReadFile(statusFile)
	if err != nil {
		return false
	}

	// Variable to hold the last updated time
	var updatedTimeStr string
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(line, "Updated,") {
			updatedTimeStr = strings.TrimPrefix(line, "Updated,")
			break
		}
	}

	// Parse the updated time
	updatedTime, err := time.Parse("2006-01-02 15:04:05", updatedTimeStr)
	if err != nil {
		return false
	}

	return time.Since(updatedTime) <= statusUpdateInterval
}

// Enable implements core.Module.
func (o *OpenVPNModule) Enable(startNow bool) error {
	o.Enabled = true
	config := map[string]interface{}{}
	utils.ObjectToMap(o, &config)
	core.SaveConfig(ModuleName, config)
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
	core.SaveConfig(ModuleName, config)
	if stopNow {
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
	config["servers"] = getOpenVPNServers()
	return config, nil
}

// SaveConfig implements core.Module.
func (o *OpenVPNModule) SaveConfig(params map[string]string, config map[string]interface{}) error {
	if !utils.HasChanged(o, config) {
		return nil
	}
	utils.MapToObject(config, o)
	err := core.SaveConfig(ModuleName, config)
	if err != nil {
		return err
	}
	saveOvpnConfig()

	killOpenVPN()
	go runOpenVPN()

	return nil
}

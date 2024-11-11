package core

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"vpn-sandbox/utils"
)

var DataDir string
var ConfigDir string
var VarDir string
var PidFile string
var AppScript string

type GlobalSettings struct {
	VPNTypes      []string `json:"vpnTypes"`
	VPNType       string   `json:"vpnType"`
	Subnets       []string `json:"subnets"`
	ProxyUsername string   `json:"proxyUsername"`
	ProxyPassword string   `json:"proxyPassword"`
}

var GlobalConfig = GlobalSettings{
	VPNTypes:      []string{"openvpn", "wireguard"},
	VPNType:       "openvpn",
	Subnets:       []string{},
	ProxyUsername: "",
	ProxyPassword: "",
}

// enum for app mode (1 = webserver, 2 = vpn-action)
type AppMode int

const (
	WebServer AppMode = iota + 1
	OpenVPNAction
)

var (
	SHUTDOWN = syscall.SIGTERM
	VPN_UP   = utils.RealTimeSignal(1)
	VPN_DOWN = utils.RealTimeSignal(2)
)

func Init(dataDir string, appMode AppMode) error {
	utils.InitSignals([]os.Signal{SHUTDOWN, VPN_UP, VPN_DOWN})

	DataDir = dataDir
	ConfigDir = filepath.Join(dataDir, "config")
	AppScript = filepath.Join(dataDir, "apps.sh")
	VarDir = filepath.Join(dataDir, "var")
	PidFile = filepath.Join(VarDir, "vpn-sandbox.pid")

	err := os.MkdirAll(ConfigDir, 0755)
	if err != nil {
		return err
	}
	err = os.MkdirAll(VarDir, 0755)
	if err != nil {
		return err
	}

	if appMode == OpenVPNAction {
		return nil
	}

	utils.InitLog(filepath.Join(VarDir, "vpn-sandbox.log"))

	// if pid file exists, and process is still running, return
	if utils.SignalRunning(PidFile, syscall.SIGCONT) {
		os.Exit(0)
	}
	err = os.WriteFile(PidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
	if err != nil {
		return err
	}

	// appMode == WebServer
	err = initDb()
	if err != nil {
		return err
	}

	var savedConfig map[string]interface{}
	savedConfig, err = GetConfig("global")
	if err == nil {
		utils.MapToObject(savedConfig, &GlobalConfig)
	} else {
		utils.ObjectToMap(GlobalConfig, &savedConfig)
		SaveConfig("global", savedConfig)
	}

	return nil
}

func GetGlobalConfig() (map[string]interface{}, error) {
	var config map[string]interface{}
	utils.ObjectToMap(GlobalConfig, &config)
	return config, nil
}

func SaveGlobalConfig(config map[string]interface{}) error {
	if !utils.HasChanged(&GlobalConfig, config) {
		return nil
	}
	utils.MapToObject(config, &GlobalConfig)
	err := SaveConfig("global", config)
	if err != nil {
		return err
	}

	utils.PublishEvent(utils.Event{Name: "global-config-changed", Context: config})

	return nil
}

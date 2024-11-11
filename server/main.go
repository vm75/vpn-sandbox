package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"vpn-sandbox/actions"
	"vpn-sandbox/core"
	"vpn-sandbox/modules/http_proxy"
	"vpn-sandbox/modules/openvpn"
	"vpn-sandbox/modules/socks_proxy"
	"vpn-sandbox/utils"
	"vpn-sandbox/webserver"
)

func oneTimeSetup() {
	markerFile := "/.initialized"

	utils.BackupResolvConf()

	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		if _, err := os.Stat(core.AppScript); err == nil {
			utils.LogF("Running one-time setup for apps script %s", core.AppScript)
			cmd := exec.Command(core.AppScript, "setup")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				utils.LogLn(err)
			}
		} else {
			utils.LogLn("No apps script found")
		}
		os.Create(markerFile)
	}
}

func main() {
	ex, err := os.Executable()
	if err != nil {
		utils.LogFatal(err)
	}
	err = os.Chdir(filepath.Dir(ex))
	if err != nil {
		utils.LogFatal(err)
	}

	params, args := utils.SmartArgs("--data|-d=/data:,--port|-p=80:,--sudo", os.Args[1:])
	dataDir := params["--data"].GetValue()
	sudoMode := params["--sudo"].IsSet()

	// detect if this is an openvpn action
	scriptType := os.Getenv("script_type")
	appMode := core.WebServer
	if scriptType != "" && len(args) > 0 && args[0][:3] == "tun" {
		utils.InitLog(filepath.Join(core.VarDir, "vpn-"+scriptType+".log"))
		appMode = core.OpenVPNAction
	}

	err = core.Init(dataDir, appMode)
	if err != nil {
		utils.LogFatal(err)
	}

	if appMode == core.OpenVPNAction {
		switch scriptType {
		case "up":
			utils.LogLn("logfile: " + filepath.Join(core.VarDir, "vpn-"+scriptType+".log"))
			utils.LogLn("core.VarDir: " + core.VarDir)
			utils.LogLn("Running script: " + scriptType)
			actions.SaveOpenVPNSpec()
			utils.SignalRunning(core.PidFile, core.VPN_UP)
		case "down":
			utils.SignalRunning(core.PidFile, core.VPN_DOWN)
		}
		os.Exit(0)
	}

	utils.AddSignalHandler([]os.Signal{core.VPN_UP, core.VPN_DOWN}, func(sig os.Signal) {
		switch sig {
		case core.VPN_UP:
			actions.VpnUp(nil)
		case core.VPN_DOWN:
			actions.VpnDown()
		}
	})

	oneTimeSetup()

	// Disable all connectivity
	actions.VpnDown()

	// Register modules
	openvpn.InitModule(sudoMode)
	http_proxy.InitModule()
	socks_proxy.InitModule()

	// Launch webserver
	webserver.WebServer(params["--port"].GetValue())
}

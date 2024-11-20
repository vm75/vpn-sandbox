package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"vpn-sandbox/actions"
	"vpn-sandbox/core"
	"vpn-sandbox/modules/openvpn"
	"vpn-sandbox/modules/proxy"
	"vpn-sandbox/modules/wireguard"
	"vpn-sandbox/utils"
	"vpn-sandbox/webserver"
)

func oneTimeSetup() {
	markerFile := "/.initialized"

	utils.BackupResolvConf()

	if _, err := os.Stat(markerFile); os.IsNotExist(err) {
		if core.Testing {
			utils.LogLn("Skipping apps setup for testing")
		} else if _, err := os.Stat(core.AppScript); err == nil {
			utils.LogF("Running one-time setup for apps script %s", core.AppScript)
			cmd := exec.Command(core.AppScript, "setup")
			cmd.Stdout = utils.GetLogFile()
			cmd.Stderr = utils.GetLogFile()
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

	params, args := utils.SmartArgs("--data|-d=/data:,--port|-p=80:,--test", os.Args[1:])
	dataDir := params["--data"].GetValue()
	core.Testing = params["--test"].IsSet()

	// detect if this is an openvpn action
	scriptType := os.Getenv("script_type")
	appMode := core.WebServer
	if scriptType != "" && len(args) > 0 && args[0][:3] == "tun" {
		appMode = core.OpenVPNAction
	}

	err = core.Init(dataDir, appMode)
	if err != nil {
		utils.LogFatal(err)
	}

	if appMode == core.OpenVPNAction {
		utils.InitLog(filepath.Join(core.VarDir, "vpn-"+scriptType+".log"))
		utils.LogF("Running openvpn action %s\n", scriptType)
		switch scriptType {
		case "up":
			utils.LogLn("Saving openvpn spec file")
			actions.SaveOpenVPNSpec()
			utils.LogLn("Signaling vpn up to main process")
			utils.SignalRunning(core.ServerPidFile, core.VPN_UP)
		case "down":
			utils.LogLn("Signaling vpn down to main process")
			utils.SignalRunning(core.ServerPidFile, core.VPN_DOWN)
		}
		os.Exit(0)
	}

	utils.AddSignalHandler([]os.Signal{core.VPN_UP, core.VPN_DOWN, core.SHUTDOWN}, func(sig os.Signal) {
		switch sig {
		case core.VPN_UP:
			actions.VpnUp(nil)
		case core.VPN_DOWN:
			actions.VpnDown()
		case core.SHUTDOWN:
			openvpn.Shutdown()
			wireguard.Shutdown()
			os.Exit(0)
		}
	})

	oneTimeSetup()

	// Disable all connectivity
	actions.VpnDown()

	// Register modules
	proxy.InitModule(proxy.HttpProxy)
	proxy.InitModule(proxy.SocksProxy)
	openvpn.InitModule()
	wireguard.InitModule()

	// Launch webserver
	webserver.WebServer(params["--port"].GetValue())
}

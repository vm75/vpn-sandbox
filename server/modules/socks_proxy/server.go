package socks_proxy

import (
	"os"
	"os/exec"
	"strconv"
	"vpn-sandbox/utils"
)

var proxyCmd *exec.Cmd = nil

func startProxy() {
	if utils.IsRunning(proxyCmd) {
		utils.LogLn("Socks Proxy is already running")
		return
	}

	err := updateRuntimeConfig()
	if err != nil {
		utils.LogError("Error updating runtime config", err)
		return
	}

	proxyCmd = exec.Command("/usr/local/sbin/sockd", "-f", configFile)

	proxyCmd.Stdout = utils.GetLogFile()
	proxyCmd.Stderr = utils.GetLogFile()

	err = proxyCmd.Start()
	if err != nil {
		utils.LogError("Error starting Socks Proxy", err)
	} else {
		utils.LogLn("Socks Proxy started with pid", proxyCmd.Process.Pid)
		os.WriteFile(pidFile, []byte(strconv.Itoa(proxyCmd.Process.Pid)), 0644)
		status := proxyCmd.Wait()
		os.Remove(pidFile)
		utils.LogF("Socks Proxy exited with status: %v\n", status)
	}
}

func stopProxy() {
	utils.LogLn("Stopping Socks Proxy")
	utils.RunCommand("/usr/bin/pkill", "-15", "sockd")
	// proxyCmd.Wait()
}

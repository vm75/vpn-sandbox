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
		return
	}

	updateRuntimeConfig()

	proxyCmd = exec.Command("/usr/local/sbin/sockd", "-f", configFile)

	proxyCmd.Stdout = utils.GetLogFile()
	proxyCmd.Stderr = utils.GetLogFile()

	err := proxyCmd.Start()
	if err != nil {
		utils.LogLn(err)
	} else {
		utils.LogLn("Socks Proxy started with pid", proxyCmd.Process.Pid)
		os.WriteFile(pidFile, []byte(strconv.Itoa(proxyCmd.Process.Pid)), 0644)
		status := proxyCmd.Wait()
		os.Remove(pidFile)
		utils.LogF("Socks Proxy exited with status: %v\n", status)
	}
}

func stopProxy() {
	utils.RunCommand("/usr/bin/pkill", "-15", "sockd")
	// proxyCmd.Wait()
}

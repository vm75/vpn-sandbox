package http_proxy

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

	proxyCmd = exec.Command("/usr/bin/tinyproxy", "-d", "-c", configFile)

	proxyCmd.Stdout = os.Stdout
	proxyCmd.Stderr = os.Stderr

	err := proxyCmd.Start()
	if err != nil {
		utils.LogLn(err)
	} else {
		utils.LogLn("Http Proxy started with pid", proxyCmd.Process.Pid)
		os.WriteFile(pidFile, []byte(strconv.Itoa(proxyCmd.Process.Pid)), 0644)
		status := proxyCmd.Wait()
		os.Remove(pidFile)
		utils.LogF("Http Proxy exited with status: %v\n", status)
	}
}

func stopProxy() {
	go utils.RunCommand("/usr/bin/pkill", "-15", "tinyproxy")
	// proxyCmd.Wait()
}

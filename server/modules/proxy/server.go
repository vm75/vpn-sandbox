package proxy

import (
	"os"
	"os/exec"
	"strconv"
	"vpn-sandbox/utils"
)

var proxyCmd *exec.Cmd = nil

func startProxy(p *ProxyModule) {
	if utils.IsRunning(proxyCmd) {
		utils.LogF("%s is already running\n", p.displayName)
		return
	}

	err := updateProxyConfig(p)
	if err != nil {
		utils.LogError("Error updating runtime config", err)
		return
	}

	proxyCmd = exec.Command(p.proxyCmd[0], p.proxyCmd[1:]...)

	proxyCmd.Stdout = utils.GetLogFile()
	proxyCmd.Stderr = utils.GetLogFile()

	err = proxyCmd.Start()
	if err != nil {
		utils.LogError("Error starting "+p.displayName, err)
	} else {
		utils.LogF("%s started with pid %d\n", p.displayName, proxyCmd.Process.Pid)
		os.WriteFile(p.pidFile, []byte(strconv.Itoa(proxyCmd.Process.Pid)), 0644)
		status := proxyCmd.Wait()
		os.Remove(p.pidFile)
		utils.LogF("%s exited with status: %v\n", p.displayName, status)
	}
}

func stopProxy(p *ProxyModule) {
	utils.LogF("Stopping %s\n", p.displayName)
	utils.RunCommand("/usr/bin/pkill", "-15", p.execName)
	// proxyCmd.Wait()
}

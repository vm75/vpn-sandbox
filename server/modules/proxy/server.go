package proxy

import (
	"os"
	"os/exec"
	"strconv"
	"vpn-sandbox/utils"
)

func startProxy(p *ProxyModule) {
	if !p.Config["enabled"].(bool) {
		return
	}
	if utils.IsRunning(p.cmdObject) {
		utils.LogF("%s is already running\n", p.displayName)
		return
	}

	err := updateProxyConfig(p)
	if err != nil {
		utils.LogError("Error updating runtime config", err)
		return
	}

	p.cmdObject = exec.Command(p.proxyCmd[0], p.proxyCmd[1:]...)

	p.cmdObject.Stdout = utils.GetLogFile()
	p.cmdObject.Stderr = utils.GetLogFile()

	err = p.cmdObject.Start()
	if err != nil {
		utils.LogError("Error starting "+p.displayName, err)
	} else {
		utils.LogF("%s started with pid %d\n", p.displayName, p.cmdObject.Process.Pid)
		os.WriteFile(p.pidFile, []byte(strconv.Itoa(p.cmdObject.Process.Pid)), 0644)
		status := p.cmdObject.Wait()
		os.Remove(p.pidFile)
		utils.LogF("%s exited with status: %v\n", p.displayName, status)
	}
}

func stopProxy(p *ProxyModule) {
	utils.LogF("Stopping %s\n", p.displayName)
	utils.RunCommand(utils.UseSudo, "/usr/bin/pkill", "-15", p.execName)
	p.cmdObject = nil
	// proxyCmd.Wait()
}

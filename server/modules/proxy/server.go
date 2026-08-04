package proxy

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

func startProxy(p *ProxyModule) {
	if !p.Config["enabled"].(bool) || !core.IsVpnUp() {
		return
	}
	p.cmdMutex.Lock()
	if utils.IsRunning(p.cmdObject) {
		p.cmdMutex.Unlock()
		utils.LogF("%s is already running\n", p.displayName)
		return
	}

	err := updateProxyConfig(p)
	if err != nil {
		p.cmdMutex.Unlock()
		utils.LogError("Error updating runtime config", err)
		return
	}

	cmd := exec.Command(p.proxyCmd[0], p.proxyCmd[1:]...)

	cmd.Stdout = utils.GetLogFile()
	cmd.Stderr = utils.GetLogFile()

	err = cmd.Start()
	if err != nil {
		p.cmdMutex.Unlock()
		utils.LogError("Error starting "+p.displayName, err)
		return
	}

	p.cmdObject = cmd
	p.done = make(chan struct{})
	done := p.done
	p.cmdMutex.Unlock()

	utils.PublishEvent(utils.Event{Name: "proxy-up", Context: map[string]interface{}{}})
	utils.LogF("%s started with pid %d\n", p.displayName, cmd.Process.Pid)
	if err := os.WriteFile(p.pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		utils.LogError("Error writing "+p.displayName+" PID file", err)
	}

	status := cmd.Wait()
	p.cmdMutex.Lock()
	if p.cmdObject == cmd {
		p.cmdObject = nil
		p.done = nil
	}
	close(done)
	p.cmdMutex.Unlock()

	utils.PublishEvent(utils.Event{Name: "proxy-down", Context: map[string]interface{}{}})
	if err := os.Remove(p.pidFile); err != nil && !os.IsNotExist(err) {
		utils.LogError("Error removing "+p.displayName+" PID file", err)
	}
	utils.LogF("%s exited with status: %v\n", p.displayName, status)
}

func stopProxy(p *ProxyModule) {
	utils.LogF("Stopping %s\n", p.displayName)
	p.cmdMutex.Lock()
	cmd := p.cmdObject
	done := p.done
	if !utils.IsRunning(cmd) {
		p.cmdMutex.Unlock()
		return
	}
	err := cmd.Process.Signal(syscall.SIGTERM)
	p.cmdMutex.Unlock()
	if err != nil {
		utils.LogError("Error stopping "+p.displayName, err)
		return
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		utils.LogF("%s did not stop after 5 seconds; killing it\n", p.displayName)
		if err := cmd.Process.Kill(); err != nil {
			utils.LogError("Error killing "+p.displayName, err)
		}
		<-done
	}
}

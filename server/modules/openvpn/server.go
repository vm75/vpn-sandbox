package openvpn

import (
	"os"
	"os/exec"
	"strconv"
	"time"
	"vpn-sandbox/core"
	"vpn-sandbox/utils"
)

const (
	binDir      = "/usr/local/bin"
	dataCiphers = "AES-256-GCM:AES-128-GCM:CHACHA20-POLY1305:AES-256-CBC:AES-128-CBC"
)

var openvpnCmd *exec.Cmd = nil

func runOpenVPN() {
	if utils.IsRunning(openvpnCmd) || !openvpnConfig.Enabled {
		return
	}

	saveOvpnConfig()

	// check if config files exist
	if !utils.FileExists(configFile) || !utils.FileExists(authFile) {
		utils.LogLn("VPN config/auth file(s) not found")
		return
	}

	execPath, _ := os.Executable()

	for openvpnConfig.Enabled {
		retryInterval := strconv.Itoa(openvpnConfig.RetryInterval)

		utils.LogLn("Starting OpenVPN")
		var cmd []string = []string{
			"openvpn",
			"--client",
			"--cd", core.VarDir,
			"--config", configFile,
			"--auth-user-pass", authFile,
			"--auth-nocache",
			"--verb", strconv.Itoa(openvpnConfig.LogLevel),
			"--log", logFile,
			"--status", statusFile, retryInterval,
			"--ping-restart", retryInterval,
			"--connect-retry-max", "3",
			"--script-security", "2",
			"--up", execPath, "--up-delay",
			"--down", execPath,
			"--up-restart",
			"--pull-filter", "ignore", "route-ipv6",
			"--pull-filter", "ignore", "ifconfig-ipv6",
			"--pull-filter", "ignore", "block-outside-dns",
			"--redirect-gateway", "def1",
			"--remote-cert-tls", "server",
			"--data-ciphers", dataCiphers,
			"--writepid", pidFile,
		}
		if core.Testing {
			cmd = append([]string{"sudo"}, cmd...)
		}
		openvpnCmd = exec.Command(
			cmd[0],
			cmd[1:]...,
		)

		openvpnCmd.Stdout = utils.GetLogFile()
		openvpnCmd.Stderr = utils.GetLogFile()

		err := openvpnCmd.Start()
		if err != nil {
			utils.LogLn(err)
			sleepFor := max(openvpnConfig.RetryInterval, 60)
			time.Sleep(time.Duration(sleepFor) * time.Second)
		} else {
			utils.LogLn("OpenVPN started with pid", openvpnCmd.Process.Pid)
			status := openvpnCmd.Wait()
			utils.LogF("OpenVPN exited with status: %v\n", status)
			sleepFor := max(openvpnConfig.RetryInterval, 60)
			time.Sleep(time.Duration(sleepFor) * time.Second)
		}

		if !openvpnConfig.Enabled {
			break
		}
	}
}

func killOpenVPN() {
	var cmd []string = []string{
		"/usr/bin/pkill", "-15", "-x", "openvpn",
	}
	if core.Testing {
		cmd = append([]string{"sudo"}, cmd...)
	}

	utils.RunCommand(cmd[0], cmd[1:]...)
	// openvpnCmd.Wait()
}

func Shutdown() {
	openvpnConfig.Enabled = false
	killOpenVPN()
	if utils.IsRunning(openvpnCmd) {
		openvpnCmd.Wait()
	}
}

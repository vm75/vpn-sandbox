package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func RunCommand(command string, args ...string) error {
	LogLn(fmt.Sprintf("Executing: %s %s", command, strings.Join(args, " ")))

	cmd := exec.Command(command, args...)
	// cmd.Env = append(os.Environ(), "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin")
	// cmd.Stdout = logFile
	// cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create a new process group
	}
	err := cmd.Run()
	if err != nil {
		LogLn(logFile, err)
	}

	return err
}

func IsRunning(cmd *exec.Cmd) bool {
	return cmd != nil && cmd.Process != nil && cmd.ProcessState != nil && !cmd.ProcessState.Exited()
}

func SignalCmd(cmd *exec.Cmd, signal os.Signal) {
	if IsRunning(cmd) {
		cmd.Process.Signal(signal)
		// openvpnCmd.Wait()
	}
}

func CreateUser(username string) {
	cmd := exec.Command("/usr/sbin/adduser", "-S", "-D", "-H", "-h", "/dev/null", "-G", username, username)
	cmd.CombinedOutput()
}

func SignalRunning(pidFile string, signal os.Signal) bool {
	isRunning := false
	if _, err := os.Stat(pidFile); err != nil {
		return isRunning
	}
	file, err := os.Open(pidFile)
	if err != nil {
		return isRunning
	}
	defer file.Close()
	var pid int
	_, err = fmt.Fscanf(file, "%d", &pid)
	if err != nil {
		return isRunning
	}
	proc, err := os.FindProcess(pid)
	if err == nil {
		err = proc.Signal(signal)
		if err != nil {
			return isRunning
		}
		isRunning = true
	}

	return isRunning
}

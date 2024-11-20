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
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create a new process group
	}
	err := cmd.Run()
	if err != nil {
		LogLn(logFile, err)
	}

	return err
}

func IsRunning(cmdObject *exec.Cmd) bool {
	return cmdObject != nil && cmdObject.Process != nil && (cmdObject.ProcessState == nil || !cmdObject.ProcessState.Exited())
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
	if _, err := os.Stat(pidFile); err != nil {
		LogF("PID file %s not found\n", pidFile)
		return false
	}

	file, err := os.Open(pidFile)
	if err != nil {
		LogF("Error opening PID file %s\n", pidFile)
		return false
	}
	defer file.Close()
	var pid int
	_, err = fmt.Fscanf(file, "%d", &pid)
	if err != nil {
		LogF("Error reading PID from file %s\n", pidFile)
		return false
	}

	err = SignalProcess(pid, syscall.Signal(0))
	if err != nil {
		LogF("Process with PID %d is not running\n", pid)
	}

	LogF("Sending signal %s to process with PID %d\n", signal, pid)
	return SignalProcess(pid, signal) == nil
}

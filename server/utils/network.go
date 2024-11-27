package utils

import (
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const ResolvConfBackup = "/etc/resolv.conf.bak"

func GetDefaultGateway() (string, error) {
	cmd := exec.Command("ip", "r")

	// Capture standard output and standard error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// get line starting with default via and get the following ip in that line
	return strings.Split(strings.Split(string(output), "default via ")[1], " ")[0], nil
}

func BackupResolvConf() {
	if !FileExists(ResolvConfBackup) {
		// copy exising resolv.conf to resolv.conf.bak
		_, err := RunCommand(false, "/bin/cp", "/etc/resolv.conf", ResolvConfBackup)
		if !FileExists(ResolvConfBackup) {
			LogError("Error creating "+ResolvConfBackup, err)
		}
	}
}

func RestoreResolvConf() {
	// copy exising resolv.conf.bak to resolv.conf. don't use cp, read content from resolv.conf.bak
	if _, err := os.Stat(ResolvConfBackup); !os.IsNotExist(err) {
		fileContent, _ := os.ReadFile(ResolvConfBackup)
		if err := os.WriteFile("/etc/resolv.conf", fileContent, 0644); err != nil {
			LogError("Error updating /etc/resolv.conf", err)
		}
	}
}

func GetHostGateway() string {
	out, _ := RunCommand(false, "/sbin/ip", "route", "show", "default")

	// get line starting with default via and get the following ip in that line
	gw := strings.Split(strings.Split(string(out), "default via ")[1], " ")[0]
	if gw != "" {
		return gw
	}

	if _, err := os.Stat(ResolvConfBackup); !os.IsNotExist(err) {
		fileContent, err := os.ReadFile(ResolvConfBackup)
		if err == nil {
			// extract first nameserver as host gateway
			lines := strings.Split(string(fileContent), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "nameserver") {
					return strings.Split(line, " ")[1]
				}
			}
		}
	}

	return ""
}

func GetIpInfo(ipInfo map[string]interface{}) error {
	LogLn("get ip info")
	// https://worldtimeapi.org/api/ip
	cmd := exec.Command("/usr/bin/wget", "-q", "-O", "-", "https://ipinfo.io/json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		LogLn(string(out))
		LogError("Error getting ip info", err)
		return err
	}

	for k := range ipInfo {
		delete(ipInfo, k)
	}

	err = json.Unmarshal(out, &ipInfo)
	if err != nil {
		return err
	}

	return nil
}

func GetIpV4Addr(dev string, stripMask bool) string {
	cmd := exec.Command("/sbin/ip", "a", "s", dev)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	// regex for ip address
	regex, _ := regexp.Compile(`([0-9]{1,3}(\.[0-9]{1,3}){3})\/[0-9]{1,2}`)

	// use regex to get ip address
	addr := regex.FindString(string(out))

	if stripMask {
		addr = strings.Split(addr, "/")[0]
	}
	return addr
}

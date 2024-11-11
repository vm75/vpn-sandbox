package utils

import (
	"testing"
)

func TestArgParse(t *testing.T) {
	argv := []string{
		"openvpn",
		"--config", "/etc/openvpn/openvpn.cnf",
		"-v", "3",
		"--auth-nocache",
	}
	params, args := SmartArgs("--config|-c=/etc/openvpn/openvpn.conf:,--verb|-v=3:,--auth-nocache,--auth-user-pass|-a=/etc/openvpn/openvpn.auth:", argv)

	if len(args) != 1 {
		t.Fatal(args)
	}

	if params["--config"].GetValue() != "/etc/openvpn/openvpn.cnf" {
		t.Fatal("Failed to parse config file")
	}

	if params["--verb"].GetValue() != "3" {
		t.Fatal("Failed to parse verb")
	}

	if !params["--auth-nocache"].IsSet() {
		t.Fatal("Failed to parse auth-nocache")
	}

	if params["--auth-user-pass"].GetValue() != "/etc/openvpn/openvpn.auth" {
		t.Fatal("Failed to parse auth-user-pass")
	}
}

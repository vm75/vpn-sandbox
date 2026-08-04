package proxy

import (
	"strings"
	"testing"
)

func TestRenderProxyConfig(t *testing.T) {
	tests := []struct {
		name      string
		proxyType ProxyType
		service   string
	}{
		{name: "HTTP", proxyType: HttpProxy, service: "proxy -p3128"},
		{name: "SOCKS", proxyType: SocksProxy, service: "socks -p1080"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config, err := renderProxyConfig(test.proxyType, "172.17.0.2", "10.8.0.2", "/data/var/proxy.log", "alice", `p"ass`)
			if err != nil {
				t.Fatal(err)
			}

			for _, expected := range []string{
				"internal 172.17.0.2",
				"external 10.8.0.2",
				`users "alice:CL:p""ass"`,
				"auth strong",
				test.service,
			} {
				if !strings.Contains(config, expected) {
					t.Errorf("config does not contain %q:\n%s", expected, config)
				}
			}
		})
	}
}

func TestRenderProxyConfigWithoutCredentials(t *testing.T) {
	config, err := renderProxyConfig(HttpProxy, "172.17.0.2", "10.8.0.2", "/data/var/proxy.log", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(config, "auth none") || strings.Contains(config, "users ") {
		t.Fatalf("unexpected authentication config:\n%s", config)
	}
}

func TestRenderProxyConfigRejectsNewlines(t *testing.T) {
	_, err := renderProxyConfig(HttpProxy, "172.17.0.2", "10.8.0.2", "/data/var/proxy.log", "alice\nallow *", "password")
	if err == nil {
		t.Fatal("expected newline-containing credentials to be rejected")
	}
}

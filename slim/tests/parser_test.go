package adapter

import "testing"

func TestRestrictedProxyProfile(t *testing.T) {
	for _, typ := range []string{"ss", "ssr", "vmess", "trojan", "hysteria", "hysteria2", "wireguard", "tuic", "ssh", "mieru", "openvpn", "zerotier", "easytier"} {
		t.Run(typ, func(t *testing.T) {
			if _, err := ParseProxy(map[string]any{"name": "disabled", "type": typ}); err == nil {
				t.Fatalf("unexpectedly accepted %s", typ)
			}
		})
	}
	for _, network := range []string{"tcp", "ws", "grpc", "xhttp"} {
		t.Run("vless-"+network, func(t *testing.T) {
			p := map[string]any{"type": "vless", "name": "test", "server": "example.invalid", "port": 443, "uuid": "00000000-0000-4000-8000-000000000001", "tls": true, "udp": true, "network": network, "client-fingerprint": "chrome", "reality-opts": map[string]any{"public-key": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}}
			_, err := ParseProxy(p)
			if network == "tcp" && err != nil {
				t.Fatal(err)
			}
			if network != "tcp" && err == nil {
				t.Fatal("accepted disabled transport")
			}
		})
	}
}

package config

import (
	"strings"
	"testing"
)

func TestKeepAwakeURL(t *testing.T) {
	cases := []struct {
		explicit, render, want string
	}{
		{"", "", ""},
		{"", "https://plimsoll-indexer.onrender.com", "https://plimsoll-indexer.onrender.com/healthz"},
		{"", "https://x.onrender.com/", "https://x.onrender.com/healthz"},
		{"https://example.org/ping", "https://x.onrender.com", "https://example.org/ping"},
		{"none", "https://x.onrender.com", ""},
	}
	for _, c := range cases {
		if got := keepAwakeURL(c.explicit, c.render); got != c.want {
			t.Errorf("keepAwakeURL(%q, %q) = %q, want %q", c.explicit, c.render, got, c.want)
		}
	}
}

func setBase(t *testing.T) {
	t.Helper()
	for _, k := range []string{"NETWORK", "STELLAR_RPC_URL", "HORIZON_URL", "STELLAR_NETWORK_PASSPHRASE"} {
		t.Setenv(k, "")
	}
	t.Setenv("DATABASE_URL", "postgres://x")
	t.Setenv("COVERAGE_LEDGER_ID", "C1")
	t.Setenv("REPORTER_REGISTRY_ID", "C2")
}

func TestNetworkPresets(t *testing.T) {
	cases := []struct {
		network, rpc, horizon, pass string
	}{
		{"", TestnetRPC, TestnetHorizon, TestnetPassphrase},
		{"testnet", TestnetRPC, TestnetHorizon, TestnetPassphrase},
		{"TestNet", TestnetRPC, TestnetHorizon, TestnetPassphrase},
		{"standalone", StandaloneRPC, StandaloneHorizon, StandalonePassphrase},
	}
	for _, c := range cases {
		setBase(t)
		t.Setenv("NETWORK", c.network)
		got, err := Load()
		if err != nil {
			t.Fatalf("NETWORK=%q: %v", c.network, err)
		}
		if got.RPCURL != c.rpc || got.HorizonURL != c.horizon || got.NetworkPassphrase != c.pass {
			t.Errorf("NETWORK=%q: got %q %q %q", c.network, got.RPCURL, got.HorizonURL, got.NetworkPassphrase)
		}
	}
}

func TestNetworkPresetOverrides(t *testing.T) {
	setBase(t)
	t.Setenv("NETWORK", "standalone")
	t.Setenv("STELLAR_RPC_URL", "http://rpc:9000/")
	t.Setenv("HORIZON_URL", "http://horizon:9001/")
	t.Setenv("STELLAR_NETWORK_PASSPHRASE", "custom")
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.RPCURL != "http://rpc:9000/" || got.HorizonURL != "http://horizon:9001" || got.NetworkPassphrase != "custom" {
		t.Errorf("overrides ignored: %+v", got)
	}
	if got.Network != "standalone" {
		t.Errorf("Network = %q", got.Network)
	}
}

func TestMainnetNeedsRPC(t *testing.T) {
	setBase(t)
	t.Setenv("NETWORK", "mainnet")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "STELLAR_RPC_URL is required") {
		t.Fatalf("want missing RPC error, got %v", err)
	}
	t.Setenv("STELLAR_RPC_URL", "https://rpc.example.org")
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.HorizonURL != MainnetHorizon || got.NetworkPassphrase != MainnetPassphrase {
		t.Errorf("mainnet defaults: %+v", got)
	}
}

func TestUnknownNetwork(t *testing.T) {
	setBase(t)
	t.Setenv("NETWORK", "futurenet")
	if _, err := Load(); err == nil || !strings.Contains(err.Error(), "NETWORK") {
		t.Fatalf("want unknown network error, got %v", err)
	}
}

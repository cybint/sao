package config_test

import (
	"testing"

	"github.com/cybint/sao/internal/config"
)

func TestFromEnvDefaults(t *testing.T) {
	t.Setenv("SAO_ADDR", "")
	t.Setenv("SAO_SHUTDOWN_TIMEOUT", "")
	t.Setenv("SAO_NATS_ADDR", "")
	t.Setenv("SAO_NATS_PORT", "")
	t.Setenv("SAO_PLUGIN_PATHS", "")

	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("from env: %v", err)
	}

	if cfg.Addr != ":8080" {
		t.Fatalf("addr = %q, want %q", cfg.Addr, ":8080")
	}

	if cfg.NATSAddr != "127.0.0.1" {
		t.Fatalf("nats addr = %q, want %q", cfg.NATSAddr, "127.0.0.1")
	}

	if cfg.NATSPort != 4222 {
		t.Fatalf("nats port = %d, want %d", cfg.NATSPort, 4222)
	}

	if len(cfg.PluginPaths) != 0 {
		t.Fatalf("plugin paths len = %d, want 0", len(cfg.PluginPaths))
	}
}

func TestFromEnvOverrides(t *testing.T) {
	t.Setenv("SAO_ADDR", ":9090")
	t.Setenv("SAO_SHUTDOWN_TIMEOUT", "15s")
	t.Setenv("SAO_NATS_ADDR", "0.0.0.0")
	t.Setenv("SAO_NATS_PORT", "5222")
	t.Setenv("SAO_PLUGIN_PATHS", "/tmp/plugin-a,/tmp/plugin-b")

	cfg, err := config.FromEnv()
	if err != nil {
		t.Fatalf("from env: %v", err)
	}

	if cfg.Addr != ":9090" {
		t.Fatalf("addr = %q, want %q", cfg.Addr, ":9090")
	}

	if cfg.NATSAddr != "0.0.0.0" {
		t.Fatalf("nats addr = %q, want %q", cfg.NATSAddr, "0.0.0.0")
	}

	if cfg.NATSPort != 5222 {
		t.Fatalf("nats port = %d, want %d", cfg.NATSPort, 5222)
	}

	if len(cfg.PluginPaths) != 2 {
		t.Fatalf("plugin paths len = %d, want 2", len(cfg.PluginPaths))
	}

	if cfg.PluginPaths[0] != "/tmp/plugin-a" || cfg.PluginPaths[1] != "/tmp/plugin-b" {
		t.Fatalf("plugin paths = %v, want [/tmp/plugin-a /tmp/plugin-b]", cfg.PluginPaths)
	}
}

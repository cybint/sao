package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cybint/sao/internal/config"
)

func TestLoadHCLCreatesDefaultFileWhenMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "etc", "sao", "config.hcl")

	cfg, err := config.LoadHCL(path)
	if err != nil {
		t.Fatalf("load hcl: %v", err)
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

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected config file to be created: %v", err)
	}
}

func TestLoadHCLParsesServerBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.hcl")
	content := `server {
  addr = ":9090"
  shutdown_timeout = "15s"
  plugin_paths = ["/tmp/plugin-a", "/tmp/plugin-b"]

  embedded_nats {
    addr = "0.0.0.0"
    port = 5222
  }
}
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := config.LoadHCL(path)
	if err != nil {
		t.Fatalf("load hcl: %v", err)
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
}

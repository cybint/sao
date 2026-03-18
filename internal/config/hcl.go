package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

const defaultConfigHCL = `server {
  addr = ":8080"
  shutdown_timeout = "10s"
  plugin_paths = []

  embedded_nats {
    addr = "127.0.0.1"
    port = 4222
  }
}
`

type fileConfig struct {
	Server *serverBlock `hcl:"server,block"`
}

type serverBlock struct {
	Addr            string             `hcl:"addr,optional"`
	ShutdownTimeout string             `hcl:"shutdown_timeout,optional"`
	PluginPaths     []string           `hcl:"plugin_paths,optional"`
	EmbeddedNATS    *embeddedNATSBlock `hcl:"embedded_nats,block"`
}

type embeddedNATSBlock struct {
	Addr string `hcl:"addr,optional"`
	Port int    `hcl:"port,optional"`
}

// LoadHCL loads config from an HCL file path.
// If the file does not exist, it is created with defaults.
func LoadHCL(path string) (Config, error) {
	if err := ensureConfigFile(path); err != nil {
		return Config{}, err
	}

	cfg := Defaults()
	var decoded fileConfig
	if err := hclsimple.DecodeFile(path, nil, &decoded); err != nil {
		return Config{}, fmt.Errorf("decode config file %q: %w", path, err)
	}

	if decoded.Server != nil {
		if decoded.Server.Addr != "" {
			cfg.Addr = decoded.Server.Addr
		}

		if decoded.Server.ShutdownTimeout != "" {
			timeout, err := time.ParseDuration(decoded.Server.ShutdownTimeout)
			if err != nil {
				return Config{}, fmt.Errorf("parse server.shutdown_timeout: %w", err)
			}
			cfg.ShutdownTimeout = timeout
		}

		cfg.PluginPaths = normalizePluginPaths(decoded.Server.PluginPaths)

		if decoded.Server.EmbeddedNATS != nil {
			if decoded.Server.EmbeddedNATS.Addr != "" {
				cfg.NATSAddr = decoded.Server.EmbeddedNATS.Addr
			}
			if decoded.Server.EmbeddedNATS.Port != 0 {
				cfg.NATSPort = decoded.Server.EmbeddedNATS.Port
			}
		}
	}

	return cfg, cfg.Validate()
}

func ensureConfigFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat config file %q: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory for %q: %w", path, err)
	}

	if err := os.WriteFile(path, []byte(defaultConfigHCL), 0o644); err != nil {
		return fmt.Errorf("create default config file %q: %w", path, err)
	}

	return nil
}

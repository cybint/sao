package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAddr               = ":8080"
	defaultShutdownTimeout    = 10 * time.Second
	defaultNATSAddr           = "127.0.0.1"
	defaultNATSPort           = 4222
	DefaultConfigPath         = "/etc/sao/config.hcl"
	DefaultShutdownTimeoutRaw = "10s"
)

// Config contains runtime configuration for the SAO server.
type Config struct {
	Addr            string
	ShutdownTimeout time.Duration
	NATSAddr        string
	NATSPort        int
	PluginPaths     []string
}

// Defaults returns the default runtime configuration.
func Defaults() Config {
	return Config{
		Addr:            defaultAddr,
		ShutdownTimeout: defaultShutdownTimeout,
		NATSAddr:        defaultNATSAddr,
		NATSPort:        defaultNATSPort,
	}
}

// FromEnv loads configuration from environment variables.
func FromEnv() (Config, error) {
	cfg := Defaults()

	if addr := os.Getenv("SAO_ADDR"); addr != "" {
		cfg.Addr = addr
	}

	if timeout := os.Getenv("SAO_SHUTDOWN_TIMEOUT"); timeout != "" {
		parsed, err := time.ParseDuration(timeout)
		if err != nil {
			return Config{}, fmt.Errorf("parse SAO_SHUTDOWN_TIMEOUT: %w", err)
		}

		cfg.ShutdownTimeout = parsed
	}

	if natsAddr := os.Getenv("SAO_NATS_ADDR"); natsAddr != "" {
		cfg.NATSAddr = natsAddr
	}

	if natsPort := os.Getenv("SAO_NATS_PORT"); natsPort != "" {
		parsed, err := strconv.Atoi(natsPort)
		if err != nil {
			return Config{}, fmt.Errorf("parse SAO_NATS_PORT: %w", err)
		}

		if parsed < 0 || parsed > 65535 {
			return Config{}, fmt.Errorf("parse SAO_NATS_PORT: out of range %d", parsed)
		}

		cfg.NATSPort = parsed
	}

	if pluginPaths := os.Getenv("SAO_PLUGIN_PATHS"); pluginPaths != "" {
		cfg.PluginPaths = normalizePluginPaths(strings.Split(pluginPaths, ","))
	}

	return cfg, cfg.Validate()
}

// Validate checks config values for correctness.
func (c Config) Validate() error {
	if c.NATSPort < 0 || c.NATSPort > 65535 {
		return fmt.Errorf("nats-port out of range: %d", c.NATSPort)
	}

	return nil
}

func normalizePluginPaths(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		for _, split := range strings.Split(path, ",") {
			trimmed := strings.TrimSpace(split)
			if trimmed == "" {
				continue
			}
			out = append(out, trimmed)
		}
	}

	return out
}

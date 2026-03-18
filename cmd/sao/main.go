package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cybint/sao/internal/config"
	"github.com/cybint/sao/internal/ext"
	"github.com/cybint/sao/internal/ext/pluginapi"
	"github.com/cybint/sao/internal/natsserver"
	"github.com/cybint/sao/internal/server"
	"github.com/nats-io/nats.go"
	"github.com/urfave/cli/v3"
)

const (
	defaultUIAddr         = ":8081"
	defaultUIShutdownTime = 10 * time.Second
)

const uiHTML = `<!doctype html>
<html>
  <head><meta charset="utf-8"><title>SAO Admin UI</title></head>
  <body>
    <h1>SAO Admin UI</h1>
    <p>Status endpoint: <a href="/api/status">/api/status</a></p>
  </body>
</html>`

func main() {
	app := &cli.Command{
		Name:           "sao",
		Usage:          "SAO command line interface",
		DefaultCommand: "server",
		Commands: []*cli.Command{
			serverCommand(),
			uiCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func serverCommand() *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "Run the SAO CoT routing server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Usage:   "Path to HCL config file",
				Value:   config.DefaultConfigPath,
				Sources: cli.EnvVars("SAO_CONFIG_FILE"),
			},
			&cli.StringFlag{
				Name:    "addr",
				Usage:   "HTTP bind address",
				Sources: cli.EnvVars("SAO_ADDR"), // override config file
			},
			&cli.DurationFlag{
				Name:    "shutdown-timeout",
				Usage:   "Graceful shutdown timeout",
				Sources: cli.EnvVars("SAO_SHUTDOWN_TIMEOUT"), // override config file
			},
			&cli.StringFlag{
				Name:    "nats-addr",
				Usage:   "Embedded NATS bind host",
				Sources: cli.EnvVars("SAO_NATS_ADDR"), // override config file
			},
			&cli.IntFlag{
				Name:    "nats-port",
				Usage:   "Embedded NATS bind port",
				Sources: cli.EnvVars("SAO_NATS_PORT"), // override config file
			},
			&cli.StringSliceFlag{
				Name:    "plugin-path",
				Usage:   "Plugin binary path (repeat or comma-separated)",
				Sources: cli.EnvVars("SAO_PLUGIN_PATHS"), // override config file
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			cfg, err := config.LoadHCL(cmd.String("config"))
			if err != nil {
				return err
			}

			applyOverrides(cmd, &cfg)
			return run(cfg)
		},
	}
}

func uiCommand() *cli.Command {
	return &cli.Command{
		Name:  "ui",
		Usage: "Run the SAO admin UI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "addr",
				Usage:   "Admin UI bind address",
				Value:   defaultUIAddr,
				Sources: cli.EnvVars("SAO_UI_ADDR"),
			},
			&cli.StringFlag{
				Name:    "nats-url",
				Usage:   "NATS URL for UI status checks (optional)",
				Sources: cli.EnvVars("SAO_UI_NATS_URL"),
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			return runUI(cmd.String("addr"), cmd.String("nats-url"))
		},
	}
}

func applyOverrides(cmd *cli.Command, cfg *config.Config) {
	if cmd.IsSet("addr") {
		cfg.Addr = cmd.String("addr")
	}

	if cmd.IsSet("shutdown-timeout") {
		cfg.ShutdownTimeout = cmd.Duration("shutdown-timeout")
	}

	if cmd.IsSet("nats-addr") {
		cfg.NATSAddr = cmd.String("nats-addr")
	}

	if cmd.IsSet("nats-port") {
		cfg.NATSPort = cmd.Int("nats-port")
	}

	if cmd.IsSet("plugin-path") {
		cfg.PluginPaths = cmd.StringSlice("plugin-path")
	}
}

func run(cfg config.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	listener, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	broker, err := natsserver.NewEmbedded(cfg)
	if err != nil {
		return err
	}

	if err := broker.Start(); err != nil {
		return err
	}
	defer broker.Shutdown()

	plugins, err := ext.Load(cfg.PluginPaths, pluginapi.StartRequest{
		NATSURL: broker.ClientURL(),
	})
	if err != nil {
		return err
	}
	defer plugins.Stop()

	srv := server.New(cfg)
	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.Serve(listener)
	}()

	log.Printf("sao-server listening on %s", listener.Addr().String())
	log.Printf("embedded nats listening on %s", broker.ClientURL())
	if names := plugins.Names(); len(names) > 0 {
		log.Printf("loaded plugins: %v", names)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		broker.Shutdown()

		return <-errCh
	case err := <-errCh:
		if errors.Is(err, net.ErrClosed) {
			return nil
		}
		return err
	}
}

func runUI(addr, natsURL string) error {
	conn, err := connectNATS(natsURL)
	if err != nil {
		return err
	}
	if conn != nil {
		defer conn.Close()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(uiHTML))
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, _ *http.Request) {
		status := "disconnected"
		if conn != nil && conn.IsConnected() {
			status = "connected"
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"service":      "ui",
			"status":       "ok",
			"nats_url":     natsURL,
			"nats_status":  status,
			"listening_on": addr,
		})
	})

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		err := httpServer.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			errCh <- nil
			return
		}
		errCh <- err
	}()

	log.Printf("sao admin ui listening on %s", addr)
	if natsURL != "" {
		log.Printf("sao admin ui connected to nats %s", natsURL)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultUIShutdownTime)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return <-errCh
	case err := <-errCh:
		return err
	}
}

func connectNATS(natsURL string) (*nats.Conn, error) {
	if natsURL == "" {
		return nil, nil
	}

	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

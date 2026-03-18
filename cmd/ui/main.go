package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/cybint/sao/internal/ext/pluginapi"
	"github.com/hashicorp/go-plugin"
	"github.com/nats-io/nats.go"
)

const defaultUIAddr = ":8081"

type uiPlugin struct {
	natsURL string
	conn    *nats.Conn
	server  *http.Server
}

func (p *uiPlugin) Name() (string, error) {
	return "ui", nil
}

func (p *uiPlugin) OnStart(req pluginapi.StartRequest) error {
	if req.NATSURL == "" {
		return fmt.Errorf("missing nats url")
	}

	conn, err := nats.Connect(req.NATSURL)
	if err != nil {
		return fmt.Errorf("connect nats: %w", err)
	}

	p.natsURL = req.NATSURL
	p.conn = conn

	addr := os.Getenv("SAO_UI_ADDR")
	if addr == "" {
		addr = defaultUIAddr
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handleIndex)
	mux.HandleFunc("/api/status", p.handleStatus)

	p.server = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("ui plugin server error: %v", err)
		}
	}()

	log.Printf("ui plugin started on %s (nats: %s)", addr, req.NATSURL)
	return nil
}

func (p *uiPlugin) OnStop() error {
	if p.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = p.server.Shutdown(ctx)
		p.server = nil
	}
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
	log.Printf("ui plugin stopped")
	return nil
}

func (p *uiPlugin) handleIndex(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(`<!doctype html>
<html>
  <head><meta charset="utf-8"><title>SAO UI</title></head>
  <body>
    <h1>SAO UI Plugin</h1>
    <p>Plugin: ui</p>
    <p>NATS: ` + p.natsURL + `</p>
    <p>Status endpoint: <a href="/api/status">/api/status</a></p>
  </body>
</html>`))
}

func (p *uiPlugin) handleStatus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"plugin":   "ui",
		"status":   "ok",
		"nats_url": p.natsURL,
	})
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginapi.HandshakeConfig,
		Plugins:         pluginapi.ServePluginMap(&uiPlugin{}),
	})
}

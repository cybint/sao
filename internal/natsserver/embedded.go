package natsserver

import (
	"fmt"
	"time"

	"github.com/cybint/sao/internal/config"
	nats "github.com/nats-io/nats-server/v2/server"
)

// Embedded wraps a local in-process NATS server.
type Embedded struct {
	server *nats.Server
}

// NewEmbedded initializes an embedded NATS server from config.
func NewEmbedded(cfg config.Config) (*Embedded, error) {
	opts := &nats.Options{
		ServerName: "sao-embedded-nats",
		Host:       cfg.NATSAddr,
		Port:       cfg.NATSPort,
		NoLog:      true,
		NoSigs:     true,
	}

	s, err := nats.NewServer(opts)
	if err != nil {
		return nil, err
	}

	return &Embedded{server: s}, nil
}

// Start runs the server and waits for readiness.
func (e *Embedded) Start() error {
	e.server.Start()
	if !e.server.ReadyForConnections(5 * time.Second) {
		return fmt.Errorf("embedded nats did not become ready")
	}

	return nil
}

// Shutdown cleanly stops the server.
func (e *Embedded) Shutdown() {
	if e.server == nil {
		return
	}

	e.server.Shutdown()
}

// ClientURL returns a connection URL for clients.
func (e *Embedded) ClientURL() string {
	return e.server.ClientURL()
}

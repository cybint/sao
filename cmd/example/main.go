package main

import (
	"fmt"
	"log"

	"github.com/cybint/sao/internal/ext/pluginapi"
	"github.com/hashicorp/go-plugin"
	"github.com/nats-io/nats.go"
)

type examplePlugin struct {
	conn *nats.Conn
	sub  *nats.Subscription
}

func (p *examplePlugin) Name() (string, error) {
	return "example-plugin", nil
}

func (p *examplePlugin) OnStart(req pluginapi.StartRequest) error {
	if req.NATSURL == "" {
		return fmt.Errorf("missing nats url")
	}

	conn, err := nats.Connect(req.NATSURL)
	if err != nil {
		return fmt.Errorf("connect nats: %w", err)
	}

	sub, err := conn.Subscribe("sao.plugin.example.ping", func(msg *nats.Msg) {
		if msg.Reply == "" {
			return
		}
		_ = conn.Publish(msg.Reply, []byte("pong"))
	})
	if err != nil {
		conn.Close()
		return fmt.Errorf("subscribe ping: %w", err)
	}

	p.conn = conn
	p.sub = sub
	log.Printf("example-plugin started and connected to %s", req.NATSURL)
	return nil
}

func (p *examplePlugin) OnStop() error {
	if p.sub != nil {
		_ = p.sub.Unsubscribe()
		p.sub = nil
	}
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
	log.Printf("example-plugin stopped")
	return nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginapi.HandshakeConfig,
		Plugins:         pluginapi.ServePluginMap(&examplePlugin{}),
	})
}

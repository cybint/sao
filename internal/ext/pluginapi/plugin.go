package pluginapi

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

const extensionPluginName = "extension"

// StartRequest includes host runtime metadata passed to plugin startup hooks.
type StartRequest struct {
	NATSURL string
}

// Extension is the shared contract between SAO and external plugins.
type Extension interface {
	Name() (string, error)
	OnStart(StartRequest) error
	OnStop() error
}

// HandshakeConfig pins the protocol used by host and plugin.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "SAO_PLUGIN",
	MagicCookieValue: "extension",
}

// PluginMap is used by the host to dispense plugins.
var PluginMap = map[string]plugin.Plugin{
	extensionPluginName: &ExtensionPlugin{},
}

// ServePluginMap is used by plugin binaries.
func ServePluginMap(impl Extension) map[string]plugin.Plugin {
	return map[string]plugin.Plugin{
		extensionPluginName: &ExtensionPlugin{Impl: impl},
	}
}

// ExtensionPlugin bridges net/rpc to an Extension implementation.
type ExtensionPlugin struct {
	plugin.Plugin
	Impl Extension
}

func (p *ExtensionPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ExtensionRPCServer{Impl: p.Impl}, nil
}

func (p *ExtensionPlugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ExtensionRPC{client: c}, nil
}

// ExtensionRPC is the client-side implementation of Extension over net/rpc.
type ExtensionRPC struct {
	client *rpc.Client
}

func (g *ExtensionRPC) Name() (string, error) {
	var resp string
	err := g.client.Call("Plugin.Name", new(interface{}), &resp)
	return resp, err
}

func (g *ExtensionRPC) OnStart(req StartRequest) error {
	var resp bool
	return g.client.Call("Plugin.OnStart", req, &resp)
}

func (g *ExtensionRPC) OnStop() error {
	var resp bool
	return g.client.Call("Plugin.OnStop", new(interface{}), &resp)
}

// ExtensionRPCServer serves Extension over net/rpc.
type ExtensionRPCServer struct {
	Impl Extension
}

func (s *ExtensionRPCServer) Name(_ interface{}, resp *string) error {
	name, err := s.Impl.Name()
	if err != nil {
		return err
	}
	*resp = name
	return nil
}

func (s *ExtensionRPCServer) OnStart(req StartRequest, resp *bool) error {
	if err := s.Impl.OnStart(req); err != nil {
		return err
	}
	*resp = true
	return nil
}

func (s *ExtensionRPCServer) OnStop(_ interface{}, resp *bool) error {
	if err := s.Impl.OnStop(); err != nil {
		return err
	}
	*resp = true
	return nil
}

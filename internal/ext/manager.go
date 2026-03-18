package ext

import (
	"fmt"
	"os/exec"

	"github.com/cybint/sao/internal/ext/pluginapi"
	"github.com/hashicorp/go-plugin"
)

type managedPlugin struct {
	name   string
	client *plugin.Client
	ext    pluginapi.Extension
}

// Manager controls external plugins loaded through hashicorp/go-plugin.
type Manager struct {
	plugins []managedPlugin
}

// Load starts plugin processes, dispenses extensions, and runs OnStart hooks.
func Load(paths []string, startReq pluginapi.StartRequest) (*Manager, error) {
	manager := &Manager{}
	for _, path := range paths {
		plugClient := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig:  pluginapi.HandshakeConfig,
			Plugins:          pluginapi.PluginMap,
			Cmd:              exec.Command(path),
			AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC},
		})

		rpcClient, err := plugClient.Client()
		if err != nil {
			plugClient.Kill()
			manager.Stop()
			return nil, fmt.Errorf("start plugin %q: %w", path, err)
		}

		raw, err := rpcClient.Dispense("extension")
		if err != nil {
			plugClient.Kill()
			manager.Stop()
			return nil, fmt.Errorf("dispense extension from %q: %w", path, err)
		}

		extImpl, ok := raw.(pluginapi.Extension)
		if !ok {
			plugClient.Kill()
			manager.Stop()
			return nil, fmt.Errorf("unexpected extension type from %q", path)
		}

		name, err := extImpl.Name()
		if err != nil {
			plugClient.Kill()
			manager.Stop()
			return nil, fmt.Errorf("get plugin name from %q: %w", path, err)
		}

		if err := extImpl.OnStart(startReq); err != nil {
			plugClient.Kill()
			manager.Stop()
			return nil, fmt.Errorf("plugin %q start hook: %w", name, err)
		}

		manager.plugins = append(manager.plugins, managedPlugin{
			name:   name,
			client: plugClient,
			ext:    extImpl,
		})
	}

	return manager, nil
}

// Names returns loaded extension names.
func (m *Manager) Names() []string {
	names := make([]string, 0, len(m.plugins))
	for _, p := range m.plugins {
		names = append(names, p.name)
	}
	return names
}

// Stop runs OnStop hooks and kills plugin processes.
func (m *Manager) Stop() {
	for i := len(m.plugins) - 1; i >= 0; i-- {
		p := m.plugins[i]
		_ = p.ext.OnStop()
		p.client.Kill()
	}
	m.plugins = nil
}

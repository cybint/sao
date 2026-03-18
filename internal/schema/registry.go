package schema

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Definition describes a client-provided schema for a message type.
type Definition struct {
	ClientID    string          `json:"client_id"`
	MessageType string          `json:"message_type"`
	Schema      json.RawMessage `json:"schema"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// Registry stores schema definitions in memory.
type Registry struct {
	mu    sync.RWMutex
	items map[string]Definition
}

// NewRegistry creates a schema registry.
func NewRegistry() *Registry {
	return &Registry{
		items: make(map[string]Definition),
	}
}

// Upsert validates and stores a schema definition.
func (r *Registry) Upsert(clientID, messageType string, schema json.RawMessage) (Definition, error) {
	clientID = strings.TrimSpace(clientID)
	messageType = strings.TrimSpace(messageType)
	if clientID == "" {
		return Definition{}, fmt.Errorf("client_id is required")
	}
	if messageType == "" {
		return Definition{}, fmt.Errorf("message_type is required")
	}
	if len(schema) == 0 {
		return Definition{}, fmt.Errorf("schema is required")
	}
	if !json.Valid(schema) {
		return Definition{}, fmt.Errorf("schema must be valid JSON")
	}

	def := Definition{
		ClientID:    clientID,
		MessageType: messageType,
		Schema:      append(json.RawMessage(nil), schema...),
		UpdatedAt:   time.Now().UTC(),
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[key(clientID, messageType)] = def

	return def, nil
}

// List returns all schemas or a filtered subset.
func (r *Registry) List(clientID, messageType string) []Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clientID = strings.TrimSpace(clientID)
	messageType = strings.TrimSpace(messageType)

	defs := make([]Definition, 0, len(r.items))
	for _, def := range r.items {
		if clientID != "" && def.ClientID != clientID {
			continue
		}
		if messageType != "" && def.MessageType != messageType {
			continue
		}
		defs = append(defs, def)
	}

	return defs
}

func key(clientID, messageType string) string {
	return clientID + "::" + messageType
}

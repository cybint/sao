package entity

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	// ErrTypeRequired indicates a missing entity type.
	ErrTypeRequired = errors.New("type is required")
	// ErrDataRequired indicates a missing entity data payload.
	ErrDataRequired = errors.New("data is required")
	// ErrDataInvalid indicates invalid JSON payload for entity data.
	ErrDataInvalid = errors.New("data must be valid JSON")
	// ErrEntityExists indicates a duplicate entity ID on create.
	ErrEntityExists = errors.New("entity already exists")
	// ErrEntityNotFound indicates an entity lookup/update/delete miss.
	ErrEntityNotFound = errors.New("entity not found")
)

// Entity is a generic client-managed communication object.
type Entity struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// Store keeps entities in memory.
type Store struct {
	mu    sync.RWMutex
	items map[string]Entity
}

// NewStore constructs an in-memory entity store.
func NewStore() *Store {
	return &Store{
		items: make(map[string]Entity),
	}
}

// Create inserts a new entity.
func (s *Store) Create(id, entityType string, data json.RawMessage) (Entity, error) {
	entityType, data, err := normalizePayload(entityType, data)
	if err != nil {
		return Entity{}, err
	}

	id = strings.TrimSpace(id)
	if id == "" {
		id, err = randomID()
		if err != nil {
			return Entity{}, err
		}
	}

	now := time.Now().UTC()
	entity := Entity{
		ID:        id,
		Type:      entityType,
		Data:      append(json.RawMessage(nil), data...),
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.items[id]; exists {
		return Entity{}, ErrEntityExists
	}
	s.items[id] = entity

	return entity, nil
}

// List returns all entities or a filtered subset.
func (s *Store) List(entityType string) []Entity {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entityType = strings.TrimSpace(entityType)
	entities := make([]Entity, 0, len(s.items))
	for _, entity := range s.items {
		if entityType != "" && entity.Type != entityType {
			continue
		}
		entities = append(entities, entity)
	}

	return entities
}

// Get returns an entity by ID.
func (s *Store) Get(id string) (Entity, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entity, ok := s.items[id]
	return entity, ok
}

// Update modifies an existing entity by ID.
func (s *Store) Update(id, entityType string, data json.RawMessage) (Entity, error) {
	entityType, data, err := normalizePayload(entityType, data)
	if err != nil {
		return Entity{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.items[id]
	if !ok {
		return Entity{}, ErrEntityNotFound
	}

	existing.Type = entityType
	existing.Data = append(json.RawMessage(nil), data...)
	existing.UpdatedAt = time.Now().UTC()
	s.items[id] = existing

	return existing, nil
}

// Delete removes an entity.
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}

func randomID() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate entity id: %w", err)
	}

	return hex.EncodeToString(buf), nil
}

func normalizePayload(entityType string, data json.RawMessage) (string, json.RawMessage, error) {
	entityType = strings.TrimSpace(entityType)
	if entityType == "" {
		return "", nil, ErrTypeRequired
	}

	if len(data) == 0 {
		return "", nil, ErrDataRequired
	}
	if !json.Valid(data) {
		return "", nil, ErrDataInvalid
	}

	return entityType, append(json.RawMessage(nil), data...), nil
}

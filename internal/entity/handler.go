package entity

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/cybint/sao/internal/apiserde"
)

type request struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// CollectionHandler handles /entities CRUD collection operations.
func CollectionHandler(store *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreate(w, r, store)
		case http.MethodGet:
			handleList(w, r, store)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

// ItemHandler handles /entities/{id} CRUD item operations.
func ItemHandler(store *Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/entities/")
		if id == "" || strings.Contains(id, "/") {
			http.NotFound(w, r)
			return
		}

		switch r.Method {
		case http.MethodGet:
			handleGet(w, r, store, id)
		case http.MethodPut:
			handleUpdate(w, r, store, id)
		case http.MethodDelete:
			handleDelete(w, r, store, id)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func handleCreate(w http.ResponseWriter, r *http.Request, store *Store) {
	var req request
	if err := apiserde.Decode(r, &req); err != nil {
		apiserde.Write(w, r, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	entity, err := store.Create(req.ID, req.Type, req.Data)
	if err != nil {
		apiserde.Write(w, r, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	apiserde.Write(w, r, http.StatusCreated, entity)
}

func handleList(w http.ResponseWriter, r *http.Request, store *Store) {
	entityType := r.URL.Query().Get("type")
	apiserde.Write(w, r, http.StatusOK, store.List(entityType))
}

func handleGet(w http.ResponseWriter, r *http.Request, store *Store, id string) {
	entity, ok := store.Get(id)
	if !ok {
		apiserde.Write(w, r, http.StatusNotFound, errorResponse{Error: "entity not found"})
		return
	}

	apiserde.Write(w, r, http.StatusOK, entity)
}

func handleUpdate(w http.ResponseWriter, r *http.Request, store *Store, id string) {
	var req request
	if err := apiserde.Decode(r, &req); err != nil {
		apiserde.Write(w, r, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	entity, err := store.Update(id, req.Type, req.Data)
	if err != nil {
		if errors.Is(err, ErrEntityNotFound) {
			apiserde.Write(w, r, http.StatusNotFound, errorResponse{Error: err.Error()})
			return
		}
		apiserde.Write(w, r, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	apiserde.Write(w, r, http.StatusOK, entity)
}

func handleDelete(w http.ResponseWriter, r *http.Request, store *Store, id string) {
	if ok := store.Delete(id); !ok {
		apiserde.Write(w, r, http.StatusNotFound, errorResponse{Error: "entity not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

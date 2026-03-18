package schema

import (
	"encoding/json"
	"net/http"

	"github.com/cybint/sao/internal/apiserde"
)

type upsertRequest struct {
	ClientID    string          `json:"client_id"`
	MessageType string          `json:"message_type"`
	Schema      json.RawMessage `json:"schema"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Handler provides schema registration and listing endpoints.
func Handler(registry *Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleUpsert(w, r, registry)
		case http.MethodGet:
			handleList(w, r, registry)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func handleUpsert(w http.ResponseWriter, r *http.Request, registry *Registry) {
	var req upsertRequest
	if err := apiserde.Decode(r, &req); err != nil {
		apiserde.Write(w, r, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	def, err := registry.Upsert(req.ClientID, req.MessageType, req.Schema)
	if err != nil {
		apiserde.Write(w, r, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	apiserde.Write(w, r, http.StatusCreated, def)
}

func handleList(w http.ResponseWriter, r *http.Request, registry *Registry) {
	clientID := r.URL.Query().Get("client_id")
	messageType := r.URL.Query().Get("message_type")
	apiserde.Write(w, r, http.StatusOK, registry.List(clientID, messageType))
}

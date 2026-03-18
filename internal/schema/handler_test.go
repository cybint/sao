package schema_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cybint/sao/internal/schema"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestHandlerRegistersAndListsSchemas(t *testing.T) {
	registry := schema.NewRegistry()
	handler := schema.Handler(registry)

	reqBody := map[string]any{
		"client_id":    "alpha",
		"message_type": "cot.position",
		"schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"uid": map[string]any{
					"type": "string",
				},
			},
		},
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	postReq := httptest.NewRequest(http.MethodPost, "/schemas", bytes.NewReader(raw))
	postRec := httptest.NewRecorder()
	handler.ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusCreated {
		t.Fatalf("post status = %d, want %d", postRec.Code, http.StatusCreated)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/schemas?client_id=alpha", nil)
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getRec.Code, http.StatusOK)
	}

	var defs []map[string]any
	if err := json.Unmarshal(getRec.Body.Bytes(), &defs); err != nil {
		t.Fatalf("decode list: %v", err)
	}

	if len(defs) != 1 {
		t.Fatalf("list length = %d, want 1", len(defs))
	}

	if defs[0]["client_id"] != "alpha" {
		t.Fatalf("client_id = %v, want alpha", defs[0]["client_id"])
	}
}

func TestHandlerRejectsInvalidSchema(t *testing.T) {
	registry := schema.NewRegistry()
	handler := schema.Handler(registry)

	postReq := httptest.NewRequest(http.MethodPost, "/schemas", bytes.NewReader([]byte(`{"client_id":"alpha"}`)))
	postRec := httptest.NewRecorder()
	handler.ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusBadRequest {
		t.Fatalf("post status = %d, want %d", postRec.Code, http.StatusBadRequest)
	}
}

func TestHandlerSupportsProtobuf(t *testing.T) {
	registry := schema.NewRegistry()
	handler := schema.Handler(registry)

	value, err := structpb.NewValue(map[string]any{
		"client_id":    "bravo",
		"message_type": "cot.chat",
		"schema": map[string]any{
			"type": "object",
		},
	})
	if err != nil {
		t.Fatalf("build protobuf request: %v", err)
	}

	raw, err := proto.Marshal(value)
	if err != nil {
		t.Fatalf("marshal protobuf request: %v", err)
	}

	postReq := httptest.NewRequest(http.MethodPost, "/schemas", bytes.NewReader(raw))
	postReq.Header.Set("Content-Type", "application/x-protobuf")
	postReq.Header.Set("Accept", "application/x-protobuf")
	postRec := httptest.NewRecorder()
	handler.ServeHTTP(postRec, postReq)

	if postRec.Code != http.StatusCreated {
		t.Fatalf("post status = %d, want %d", postRec.Code, http.StatusCreated)
	}
	if got := postRec.Header().Get("Content-Type"); got != "application/x-protobuf" {
		t.Fatalf("content-type = %q, want %q", got, "application/x-protobuf")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/schemas?client_id=bravo", nil)
	getReq.Header.Set("Accept", "application/x-protobuf")
	getRec := httptest.NewRecorder()
	handler.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getRec.Code, http.StatusOK)
	}

	respValue := &structpb.Value{}
	if err := proto.Unmarshal(getRec.Body.Bytes(), respValue); err != nil {
		t.Fatalf("decode protobuf response: %v", err)
	}

	list, ok := respValue.AsInterface().([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("protobuf response list invalid: %v", respValue.AsInterface())
	}
}

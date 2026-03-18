package entity_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cybint/sao/internal/entity"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestEntityCRUDHandlers(t *testing.T) {
	store := entity.NewStore()
	collection := entity.CollectionHandler(store)
	item := entity.ItemHandler(store)

	createReq := map[string]any{
		"id":   "entity-1",
		"type": "cot.position",
		"data": map[string]any{"uid": "A1"},
	}
	createRaw, err := json.Marshal(createReq)
	if err != nil {
		t.Fatalf("marshal create request: %v", err)
	}

	createHTTPReq := httptest.NewRequest(http.MethodPost, "/entities", bytes.NewReader(createRaw))
	createRec := httptest.NewRecorder()
	collection.ServeHTTP(createRec, createHTTPReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createRec.Code, http.StatusCreated)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/entities/entity-1", nil)
	getRec := httptest.NewRecorder()
	item.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getRec.Code, http.StatusOK)
	}

	updateReq := map[string]any{
		"type": "cot.position",
		"data": map[string]any{"uid": "A1", "lat": 10.0},
	}
	updateRaw, err := json.Marshal(updateReq)
	if err != nil {
		t.Fatalf("marshal update request: %v", err)
	}

	updateHTTPReq := httptest.NewRequest(http.MethodPut, "/entities/entity-1", bytes.NewReader(updateRaw))
	updateRec := httptest.NewRecorder()
	item.ServeHTTP(updateRec, updateHTTPReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d", updateRec.Code, http.StatusOK)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/entities?type=cot.position", nil)
	listRec := httptest.NewRecorder()
	collection.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listRec.Code, http.StatusOK)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/entities/entity-1", nil)
	deleteRec := httptest.NewRecorder()
	item.ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", deleteRec.Code, http.StatusNoContent)
	}
}

func TestEntityCRUDHandlersProtobuf(t *testing.T) {
	store := entity.NewStore()
	collection := entity.CollectionHandler(store)
	item := entity.ItemHandler(store)

	createValue, err := structpb.NewValue(map[string]any{
		"id":   "entity-2",
		"type": "cot.position",
		"data": map[string]any{"uid": "B2"},
	})
	if err != nil {
		t.Fatalf("build protobuf create payload: %v", err)
	}
	createRaw, err := proto.Marshal(createValue)
	if err != nil {
		t.Fatalf("marshal protobuf create payload: %v", err)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/entities", bytes.NewReader(createRaw))
	createReq.Header.Set("Content-Type", "application/x-protobuf")
	createReq.Header.Set("Accept", "application/x-protobuf")
	createRec := httptest.NewRecorder()
	collection.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createRec.Code, http.StatusCreated)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/entities/entity-2", nil)
	getReq.Header.Set("Accept", "application/x-protobuf")
	getRec := httptest.NewRecorder()
	item.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", getRec.Code, http.StatusOK)
	}

	resp := &structpb.Value{}
	if err := proto.Unmarshal(getRec.Body.Bytes(), resp); err != nil {
		t.Fatalf("decode protobuf get response: %v", err)
	}

	obj, ok := resp.AsInterface().(map[string]any)
	if !ok {
		t.Fatalf("expected object response, got %T", resp.AsInterface())
	}
	if obj["id"] != "entity-2" {
		t.Fatalf("id = %v, want entity-2", obj["id"])
	}
}

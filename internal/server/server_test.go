package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/cybint/sao/internal/config"
	"github.com/cybint/sao/internal/server"
)

func TestServerServesHealthEndpoint(t *testing.T) {
	srv := server.New(config.Config{Addr: "127.0.0.1:0"})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(listener)
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + listener.Addr().String() + "/healthz")
	if err != nil {
		t.Fatalf("request /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("serve returned error: %v", err)
	}
}

func TestServerAcceptsSchemaSubmissions(t *testing.T) {
	srv := server.New(config.Config{Addr: "127.0.0.1:0"})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(listener)
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	reqPayload := map[string]any{
		"client_id":    "alpha",
		"message_type": "cot.position",
		"schema": map[string]any{
			"type": "object",
		},
	}
	raw, err := json.Marshal(reqPayload)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	postResp, err := client.Post("http://"+listener.Addr().String()+"/schemas", "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("request /schemas post: %v", err)
	}
	defer postResp.Body.Close()

	if postResp.StatusCode != http.StatusCreated {
		t.Fatalf("post status code = %d, want %d", postResp.StatusCode, http.StatusCreated)
	}

	getResp, err := client.Get("http://" + listener.Addr().String() + "/schemas?client_id=alpha")
	if err != nil {
		t.Fatalf("request /schemas get: %v", err)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get status code = %d, want %d", getResp.StatusCode, http.StatusOK)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("serve returned error: %v", err)
	}
}

func TestServerEntityCRUD(t *testing.T) {
	srv := server.New(config.Config{Addr: "127.0.0.1:0"})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(listener)
	}()

	client := &http.Client{Timeout: 2 * time.Second}

	createPayload := map[string]any{
		"id":   "entity-1",
		"type": "cot.position",
		"data": map[string]any{"uid": "U1"},
	}
	createRaw, err := json.Marshal(createPayload)
	if err != nil {
		t.Fatalf("marshal create payload: %v", err)
	}

	createResp, err := client.Post("http://"+listener.Addr().String()+"/entities", "application/json", bytes.NewReader(createRaw))
	if err != nil {
		t.Fatalf("request /entities post: %v", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create status code = %d, want %d", createResp.StatusCode, http.StatusCreated)
	}

	getResp, err := client.Get("http://" + listener.Addr().String() + "/entities/entity-1")
	if err != nil {
		t.Fatalf("request /entities/{id} get: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get status code = %d, want %d", getResp.StatusCode, http.StatusOK)
	}

	updatePayload := map[string]any{
		"type": "cot.position",
		"data": map[string]any{"uid": "U1", "lat": 42.0},
	}
	updateRaw, err := json.Marshal(updatePayload)
	if err != nil {
		t.Fatalf("marshal update payload: %v", err)
	}

	updateReq, err := http.NewRequest(http.MethodPut, "http://"+listener.Addr().String()+"/entities/entity-1", bytes.NewReader(updateRaw))
	if err != nil {
		t.Fatalf("build put request: %v", err)
	}
	updateReq.Header.Set("Content-Type", "application/json")

	updateResp, err := client.Do(updateReq)
	if err != nil {
		t.Fatalf("request /entities/{id} put: %v", err)
	}
	defer updateResp.Body.Close()
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("update status code = %d, want %d", updateResp.StatusCode, http.StatusOK)
	}

	deleteReq, err := http.NewRequest(http.MethodDelete, "http://"+listener.Addr().String()+"/entities/entity-1", nil)
	if err != nil {
		t.Fatalf("build delete request: %v", err)
	}

	deleteResp, err := client.Do(deleteReq)
	if err != nil {
		t.Fatalf("request /entities/{id} delete: %v", err)
	}
	defer deleteResp.Body.Close()
	if deleteResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status code = %d, want %d", deleteResp.StatusCode, http.StatusNoContent)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("serve returned error: %v", err)
	}
}

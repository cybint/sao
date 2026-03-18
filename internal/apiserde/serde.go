package apiserde

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const protobufContentType = "application/x-protobuf"

// Decode unmarshals request bodies from JSON or protobuf.
// Protobuf payloads use google.protobuf.Value encoded in binary format.
func Decode(r *http.Request, target any) error {
	defer r.Body.Close()

	if isProtobufContentType(r.Header.Get("Content-Type")) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			return fmt.Errorf("read protobuf body: %w", err)
		}

		value := &structpb.Value{}
		if err := proto.Unmarshal(raw, value); err != nil {
			return fmt.Errorf("decode protobuf body: %w", err)
		}

		bodyJSON, err := json.Marshal(value.AsInterface())
		if err != nil {
			return fmt.Errorf("convert protobuf body to json: %w", err)
		}

		if err := json.Unmarshal(bodyJSON, target); err != nil {
			return fmt.Errorf("decode body from protobuf: %w", err)
		}

		return nil
	}

	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return fmt.Errorf("decode json body: %w", err)
	}

	return nil
}

// Write serializes responses as JSON or protobuf based on Accept header.
// If Accept requests protobuf, the payload is returned as a binary
// google.protobuf.Value message.
func Write(w http.ResponseWriter, r *http.Request, status int, payload any) {
	if wantsProtobuf(r) {
		writeProtobuf(w, status, payload)
		return
	}

	writeJSON(w, status, payload)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeProtobuf(w http.ResponseWriter, status int, payload any) {
	genericPayload, err := toGenericPayload(payload)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to encode protobuf payload"})
		return
	}

	value, err := structpb.NewValue(genericPayload)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to encode protobuf payload"})
		return
	}

	raw, err := proto.Marshal(value)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to encode protobuf payload"})
		return
	}

	w.Header().Set("Content-Type", protobufContentType)
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

func wantsProtobuf(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if isProtobufContentType(accept) {
		return true
	}

	// Fall back to request content type if Accept is absent.
	return accept == "" && isProtobufContentType(r.Header.Get("Content-Type"))
}

func isProtobufContentType(v string) bool {
	contentType := strings.ToLower(v)
	return strings.Contains(contentType, "application/x-protobuf") || strings.Contains(contentType, "application/protobuf")
}

func toGenericPayload(payload any) (any, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, err
	}

	return generic, nil
}

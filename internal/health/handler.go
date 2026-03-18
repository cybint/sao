package health

import (
	"encoding/json"
	"net/http"
)

type response struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// Handler returns a health endpoint handler.
func Handler(service string) http.Handler {
	if service == "" {
		service = "sao"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(response{
			Status:  "ok",
			Service: service,
		})
	})
}

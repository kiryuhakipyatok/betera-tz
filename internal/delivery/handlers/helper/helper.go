package helper

import (
	"betera-tz/internal/delivery/apierr"
	"encoding/json"
	"net/http"
)

func WriteJSONError(w http.ResponseWriter, apiErr apierr.ApiErr) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Code)
	json.NewEncoder(w).Encode(map[string]any{
		"code":    apiErr.Code,
		"message": apiErr.Message,
	})
}

func ToPtr[T any](v T) *T {
	return &v
}

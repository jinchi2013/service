package checkgrp

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// Handlers manages the set of check endpoints.
type Handlers struct {
	Build string
	Log   *zap.SugaredLogger
}

func (h Handlers) Readiness(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Build  string
		Status string `json:"status"`
	}{
		Build:  h.Build,
		Status: "OK",
	}

	statusCode := http.StatusOK

	if err := response(w, statusCode, data); err != nil {
		h.Log.Errorw("readiness", "ERROR", err)
	}

	h.Log.Infow("readiness",
		"statusCode", statusCode,
		"method", r.Method,
		"path", r.URL.Path,
		"remoteaddr", r.RemoteAddr,
	)

}

func response(w http.ResponseWriter, statusCode int, data any) error {
	// Convert response value to JSON
	jsonResponse, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Set the content type and headers once we know marshaling has succeeded.
	w.Header().Set("Content-Type", "application/json")

	// Write the status code to the response.
	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonResponse); err != nil {
		return err
	}

	return nil
}

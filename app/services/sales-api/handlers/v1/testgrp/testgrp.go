package testgrp

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

func (h Handlers) Test(w http.ResponseWriter, r *http.Request) {
	status := struct {
		Status string
	}{
		Status: "ok",
	}

	json.NewEncoder(w).Encode(status)
	statusCode := http.StatusOK
	h.Log.Infow("TEST_V1",
		"statusCode", statusCode,
		"method", r.Method,
		"path", r.URL.Path,
		"remoteaddr", r.RemoteAddr,
	)
}

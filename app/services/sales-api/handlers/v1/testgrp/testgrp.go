package testgrp

import (
	"context"
	"errors"
	"math/rand"
	"net/http"

	"github.com/jinchi2013/service/busniess/sys/validate"
	"github.com/jinchi2013/service/foundation/web"
	"go.uber.org/zap"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

func (h Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	n := rand.Intn(100)
	if n < 40 {
		return errors.New("untrusted error")
	}
	if n < 80 {
		return validate.NewRequestError(errors.New("trusted error"), http.StatusBadRequest)
	}
	if n < 100 {
		return web.NewShutdownError("Restart service")
	}
	status := struct {
		Status string
	}{
		Status: "ok_1",
	}

	return web.Response(ctx, w, status, http.StatusOK)
}

package testgrp

import (
	"context"
	"net/http"

	"github.com/jinchi2013/service/foundation/web"
	"go.uber.org/zap"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

func (h Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	status := struct {
		Status string
	}{
		Status: "ok_1",
	}

	return web.Response(ctx, w, status, http.StatusOK)
}

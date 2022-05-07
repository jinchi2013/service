// Package web contains a small web framework extension.
package web

import (
	"context"
	"net/http"
	"os"
	"syscall"

	"github.com/dimfeld/httptreemux/v5"
)

type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []Middleware
}

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	return &App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
		mw:         mw,
	}
}

func (a *App) Shutdown() {
	a.shutdown <- syscall.SIGTERM
}

func (a *App) Handle(method string, group string, path string, handler Handler, mw ...Middleware) {

	wrapMiddleware(mw, handler)

	wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		if err := handler(r.Context(), w, r); err != nil {
			return
		}
	}

	finalPath := ""
	if group != "" {
		finalPath = "/" + group + path
	}

	a.ContextMux.Handle(method, finalPath, h)
}

func GetValues(ctx context.Context) (string, error) {
	return "ok", nil
}

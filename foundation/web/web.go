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
}

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

func NewApp(shutdown chan os.Signal) *App {
	return &App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
	}
}

func (a *App) Shutdown() {
	a.shutdown <- syscall.SIGTERM
}

func (a *App) Handle(method string, group string, path string, handler Handler) {

	h := func(w http.ResponseWriter, r *http.Request) {
		// Pre Code handle
		if err := handler(r.Context(), w, r); err != nil {
			// Handle error
			return
		}

		// Post  Code handle
	}

	finalPath := ""
	if group != "" {
		finalPath = "/" + group + path
	}

	a.ContextMux.Handle(method, finalPath, h)
}

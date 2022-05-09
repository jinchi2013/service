// Package web contains a small web framework extension.
package web

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/uuid"
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

func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

func (a *App) Handle(method string, group string, path string, handler Handler, mw ...Middleware) {

	// First wrap handler specific middleware around this handler.
	handler = wrapMiddleware(mw, handler)

	// Add the application's general middleware to the handler chain.
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {

		// Pull the context from the request and
		// use it as a separate parameter.
		ctx := r.Context()

		// Set the context with the required values to
		// process the request.
		v := Values{
			TraceID: uuid.New().String(),
			Now:     time.Now().UTC(),
		}

		ctx = context.WithValue(ctx, key, &v)
		if err := handler(ctx, w, r); err != nil {
			a.SignalShutdown()
			return
		}
	}

	finalPath := ""
	if group != "" {
		finalPath = "/" + group + path
	}

	a.ContextMux.Handle(method, finalPath, h)
}

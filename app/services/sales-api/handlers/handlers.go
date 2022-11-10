package handlers

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/jinchi2013/service/app/services/sales-api/handlers/debug/checkgrp"
	v1TestGrp "github.com/jinchi2013/service/app/services/sales-api/handlers/v1/testgrp"
	v1UserGrp "github.com/jinchi2013/service/app/services/sales-api/handlers/v1/usergrp"
	userCore "github.com/jinchi2013/service/busniess/core/user"
	"github.com/jinchi2013/service/busniess/sys/auth"
	"github.com/jinchi2013/service/busniess/web/mid"
	"github.com/jinchi2013/service/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// DebugStandardLibraryMux registers all the debug routes from the standard library
// into a new mux bypassing the use of the DefaultServerMux. Using the
// DefaultServerMux would be a security risk since a dependency could inject a
// handler into our service without us knowing it.
func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

func DebugMux(build string, log *zap.SugaredLogger, db *sqlx.DB) *http.ServeMux {
	mux := DebugStandardLibraryMux()

	// Regster debug check endpoint
	cgh := checkgrp.Handlers{ // health check handlers group
		Build: build,
		Log:   log,
		DB:    db,
	}

	mux.HandleFunc("/debug/readiness", cgh.Readiness)
	mux.HandleFunc("/debug/liveness", cgh.Liveness)

	return mux
}

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	Auth     *auth.Auth
	DB       *sqlx.DB
}

func APIMux(cfg APIMuxConfig) *web.App {
	app := web.NewApp(
		cfg.Shutdown,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Metrics(),
		mid.Panics(),
	)

	// Load routes for different version of api
	v1(app, cfg)

	return app
}

// v1 binds all the version 1 routes
func v1(app *web.App, cfg APIMuxConfig) {
	const version = "v1"
	// Test handlers group
	tgh := v1TestGrp.Handlers{
		Log: cfg.Log,
	}
	app.Handle(http.MethodGet, version, "/test", tgh.Test)
	app.Handle(
		http.MethodGet,
		version,
		"/testauth",
		tgh.Test,
		mid.Authenticate(cfg.Auth),
		mid.Authorize("ADMIN"),
	)

	// Register user management and authenticate endpoints.
	ugh := v1UserGrp.Handlers{
		User: userCore.NewCore(cfg.Log, cfg.DB),
		Auth: cfg.Auth,
	}

	app.Handle(http.MethodGet, version, "/users/token", ugh.Token)
	app.Handle(
		http.MethodGet,
		version,
		"/users/:page/:rows",
		ugh.Query,
		mid.Authenticate((cfg.Auth)),
		mid.Authorize(auth.RoleAdmin),
	)
	app.Handle(
		http.MethodGet,
		version,
		"/users/:id",
		ugh.QueryByID,
		mid.Authenticate((cfg.Auth)),
		mid.Authorize(auth.RoleAdmin),
	)
	app.Handle(
		http.MethodPost,
		version,
		"/users",
		ugh.Create,
		mid.Authenticate((cfg.Auth)),
		mid.Authorize(auth.RoleAdmin),
	)
	app.Handle(
		http.MethodPut,
		version,
		"/users/:id",
		ugh.Update,
		mid.Authenticate((cfg.Auth)),
		mid.Authorize(auth.RoleAdmin),
	)
	app.Handle(
		http.MethodDelete,
		version,
		"/users/:id",
		ugh.Update,
		mid.Authenticate((cfg.Auth)),
		mid.Authorize(auth.RoleAdmin),
	)
}

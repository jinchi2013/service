package mid

import (
	"context"
	"net/http"
	"time"

	"github.com/jinchi2013/service/foundation/web"
	"go.uber.org/zap"
)

func Logger(log *zap.SugaredLogger) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// If the context is missing this value, request the service
			// to be shutdown gracefully.
			v, err := web.GetValues(ctx)

			if err != nil {
				return err
			}

			log.Infow("Request Started",
				"traceid", v.TraceID,
				"method", r.Method,
				"path", r.URL.Path,
				"remoteaddr", r.RemoteAddr,
			)

			err = handler(ctx, w, r)

			log.Infow("Request Completed",
				"traceid", v.TraceID,
				"method", r.Method,
				"path", r.URL.Path,
				"remoteaddr", r.RemoteAddr,
				"statusCode", v.StatusCode,
				"since", time.Since(v.Now),
			)
			// Return the error so it can be handled further up the chain.
			return err
		}

		return h

	}

	return m

}

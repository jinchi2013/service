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
			now := time.Now()

			if err != nil {
				return err
			}

			log.Infow("Request Started",
				"traceID", "00000000000000000",
				"statusCode", v,
				"method", r.Method,
				"path", r.URL.Path,
				"remoteaddr", r.RemoteAddr,
			)

			err = handler(ctx, w, r)

			log.Infow("Request Completed",
				"traceID", "00000000000000000",
				"statusCode", v,
				"method", r.Method,
				"path", r.URL.Path,
				"remoteaddr", r.RemoteAddr,
				"since", time.Since(now),
			)

			return err
		}

		return h

	}

	return m

}

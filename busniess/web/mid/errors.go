package mid

import (
	"context"
	"net/http"

	"github.com/jinchi2013/service/busniess/sys/validate"
	"github.com/jinchi2013/service/foundation/web"
	"go.uber.org/zap"
)

func Errors(log *zap.SugaredLogger) web.Middleware {
	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v, err := web.GetValues(ctx)

			if err != nil {
				return web.NewShutdownError("web value missing from context")
			}

			// Run the next handler and catch any propagated error.
			if err := handler(ctx, w, r); err != nil {
				log.Errorw("ERROR", "traceid", v.TraceID, "ERROR", err)

				// Build out error response
				var er validate.ErrorResponse
				var status int
				switch act := validate.Cause(err).(type) {
				case validate.FieldErrors: // data modal validation failed, note: use value sementics for slices
					er = validate.ErrorResponse{
						Error: "data validation error",
						Field: act.Error(),
					}
					status = http.StatusBadRequest
				case *validate.RequestError:
					er = validate.ErrorResponse{
						Error: act.Error(),
					}
					status = act.Status
				default: // non-trusted error
					er = validate.ErrorResponse{
						Error: http.StatusText(http.StatusInternalServerError),
					}
					status = http.StatusInternalServerError
				}

				// Response the error back to the client
				if err := web.Response(ctx, w, er, status); err != nil {
					return err
				}

				// The err below is the err on line 23
				// If we receive  the shutdown err we need to return it
				// back to the base  handler to shutdown the service
				if ok := web.IsShutdown(err); ok {
					return err
				}
			}

			return nil
		}

		return h
	}

	return m
}

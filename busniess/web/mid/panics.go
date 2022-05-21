package mid

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/jinchi2013/service/foundation/web"
)

func Panics() web.Middleware {
	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
			// the syntax (err error) above is to name the return error
			// which only make sense in the panic case,
			//	cause when the error occurs,
			// we are no longer in current function

			// Defer a function to recover from a panic and set err return
			// variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {
					trace := debug.Stack()
					// Stack trace will be provided
					// Assign rec to err as error, which is the return value on line 14
					err = fmt.Errorf("PANIC [%v], TRACE:[%s]", rec, string(trace))
				}
			}()

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

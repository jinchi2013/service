package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jinchi2013/service/busniess/sys/auth"
	"github.com/jinchi2013/service/busniess/sys/validate"
	"github.com/jinchi2013/service/foundation/web"
)

func Authenticate(a *auth.Auth) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// Expecting: bearer <token>
			authStr := r.Header.Get("authorization")

			// Parse the authorization header
			parts := strings.Split(authStr, " ")

			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				err := errors.New("expected autheriztion header format: bearer <token>")
				return validate.NewRequestError(err, http.StatusUnauthorized)
			}

			// Validate the token is singed by us
			claims, err := a.ValidateToken(parts[1])
			if err != nil {
				return validate.NewRequestError(err, http.StatusUnauthorized)
			}

			// Add claims to the context so they can be retrieved later
			ctx = auth.SetClaims(ctx, claims)

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

func Authorize(roles ...string) web.Middleware {
	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			claims, err := auth.GetClaims(ctx)

			if err != nil {
				return validate.NewRequestError(
					fmt.Errorf("you are not authorized for that action, no claims"),
					http.StatusForbidden,
				)
			}

			if !claims.Authorized(roles...) {
				return validate.NewRequestError(
					fmt.Errorf("you are not authorized for that action, claims[%v] roles[%v]", claims.Roles, roles),
					http.StatusForbidden,
				)
			}

			return handler(ctx, w, r)
		}
		return h
	}

	return m
}

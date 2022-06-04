package auth

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v4"
)

//
const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

type Claims struct {
	jwt.StandardClaims
	Roles []string `json:"roles"`
}

// Authorized returns true if the claims has at least one of the provided roles
func (c Claims) Authorized(roles ...string) bool {
	for _, has := range c.Roles {
		for _, want := range roles {
			if has == want {
				return true
			}
		}
	}

	return false
}

// ctxKey represents the type of value for the context key
type ctxKey int

// key is used to store/retrieve a Claims value for a context.Context
const key ctxKey = 1

// SetClaims stores the claims in the context
func SetClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, key, claims)
}

func GetClaims(ctx context.Context) (Claims, error) {
	v, ok := ctx.Value(key).(Claims)
	if !ok {
		return Claims{}, errors.New("claim value missing from context")
	}

	return v, nil
}

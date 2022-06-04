package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

// KeyLookup declares  a method set of behavior for looking up
// private and public key for jwt use
type KeyLookup interface {
	PrivateKey(kid string) (*rsa.PrivateKey, error)
	PublicKey(kid string) (*rsa.PublicKey, error)
}

// Auth is used to authenticate clients
// It can generate a token for a set of user claims and recreate the claims by parsing the token
type Auth struct {
	activeKID string // if we want to do some rotation, this activeKID will keep changing
	keyLookup KeyLookup
	method    jwt.SigningMethod
	keyFunc   func(t *jwt.Token) (any, error) // return public key by the kid in token header
	parser    jwt.Parser
}

// New creates an Auth to support authentication/authorization
func New(activeKID string, keyLookup KeyLookup) (*Auth, error) {

	// the new activeKID represents the privatekey used to signed new tokens
	_, err := keyLookup.PrivateKey(activeKID)

	if err != nil {
		return nil, errors.New("active KID does not exist in store")
	}

	method := jwt.GetSigningMethod("RS256")

	if method == nil {
		return nil, errors.New("configuraing algorithm RS256")
	}

	keyFunc := func(t *jwt.Token) (any, error) {
		kid, ok := t.Header["kid"]

		if !ok {
			return nil, errors.New("missing key id in token header")
		}

		kidID, ok := kid.(string)

		if !ok {
			return nil, errors.New("user token key id must be a string")
		}

		return keyLookup.PublicKey(kidID)
	}

	parser := jwt.Parser{
		ValidMethods: []string{"RS256"},
	}

	a := Auth{
		parser:    parser,
		activeKID: activeKID,
		keyLookup: keyLookup,
		keyFunc:   keyFunc,
		method:    method,
	}

	return &a, nil
}

func (a *Auth) GenerateToken(claims Claims) (string, error) {
	// Generate token based on claims
	token := jwt.NewWithClaims(a.method, claims)

	token.Header["kid"] = a.activeKID

	privateKey, err := a.keyLookup.PrivateKey(a.activeKID)
	if err != nil {
		return "", errors.New("kid lookup failed")
	}

	// use the privatekey to signed the token
	str, err := token.SignedString(privateKey)

	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return str, nil
}

// ValidateToken recreates the Claims that were used to generate a token
// It verfies that token was signed using our key
func (a *Auth) ValidateToken(tokenStr string) (Claims, error) {
	var claims Claims
	token, err := a.parser.ParseWithClaims(tokenStr, &claims, a.keyFunc)

	if err != nil {
		return Claims{}, fmt.Errorf("parsing token: %w", err)
	}

	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return claims, nil
}

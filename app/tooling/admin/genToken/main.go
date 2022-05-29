package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func main() {
	err := genToken()

	if err != nil {
		fmt.Println(err)
	}
}

func genToken() error {
	// read our file in zarf folder
	// name is the relative path to the root
	name := "zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem"
	file, err := os.Open(name)

	// in production, there will be a service to save the token

	if err != nil {
		return err
	}

	// limit PEM file size to 1 Megabyte. This should be
	// reasonable for almost any PEM file and prevents shenanigans like
	// linking the file to /dev/random ro something like that
	privatePEM, err := io.ReadAll(io.LimitReader(file, 1024*1024))
	if err != nil {
		return fmt.Errorf("reading auth private key: %w", err)
	}

	// Use jwt to parse the PEM to privateKey
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)

	if err != nil {
		return fmt.Errorf("parsing auth private key: %w", err)
	}

	// ========================================================================
	/*
		Generating a token requires defining a set of claims
		In this application case, we only care about defining the subject
		and the user in question and the roles they have on the database
		This token will expire in a year

		iss (issuer): Issuer of the JWT
		sub (subject): Subject of the JWT (the user)
		aud (audience): Recipient for which the JWT is intended
		exp (expiration time): Time after which jwt expires
		nbf (not before time): Time before which the jwt must not be accepted for processing
		iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
		jti (JWT ID): Unique identifier; can be used to prevent the jwt from being replayed
		(allow the token to used only once)
	*/
	// claim a JWT token
	// it is like the info you would like to save in the token
	claims := struct {
		jwt.StandardClaims
		Roles []string
	}{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "server project",
			Subject:   "123456789", // Customer ID
			ExpiresAt: time.Now().Add(8760 * time.Hour).Unix(),
			IssuedAt:  time.Now().UTC().Unix(),
		},
		Roles: []string{"ADMIN"},
	}

	method := jwt.GetSigningMethod("RS256")

	// Create new jwt token with method, claims
	token := jwt.NewWithClaims(method, claims)

	token.Header["kid"] = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

	// Use the private key to sign the signature
	str, err := token.SignedString(privateKey)

	if err != nil {
		return err
	}

	fmt.Println("======== TOKEN BEGIN ========")
	fmt.Println(str) // print signed token
	fmt.Println("======== TOKEN END ========")
	fmt.Println("======== ======== ======== ======== ======== ========")

	// ================================================================================
	// Inorder the validate the signed token above
	// We would need the publickey from the private key

	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return fmt.Errorf("marshing public key: %w", err)
	}

	// Contract a PEM block for the public key
	publicBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the public key file.
	if err := pem.Encode(os.Stdout, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	return nil
}

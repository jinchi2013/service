package keystore

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v4"
)

type KeyStore struct {
	mu    sync.RWMutex
	store map[string]*rsa.PrivateKey
}

// New Contructs an empty KeyStore ready for use
func New() *KeyStore {
	return &KeyStore{
		store: make(map[string]*rsa.PrivateKey),
	}
}

// NewMap constructs a KeyStore with an initial set of keys.
func NewMap(store map[string]*rsa.PrivateKey) *KeyStore {
	return &KeyStore{
		store: store,
	}
}

// NewFS constructs a KeyStore based on a set of PEM files rooted inside of a directory.
// The name of each PEM file with be used as the key id
// Example: keystore.NewFS(os.DirFS("/zarf/keys/"))
// Example: /zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem
func NewFS(fsys fs.FS) (*KeyStore, error) {
	ks := KeyStore{
		store: make(map[string]*rsa.PrivateKey),
	}

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walkdir failure: %w", err)
		}

		if dirEntry.IsDir() {
			return nil
		}

		if path.Ext(fileName) != ".pem" {
			return nil
		}

		file, err := fsys.Open(fileName)

		if err != nil {
			return fmt.Errorf("opening key file: %w", err)
		}

		defer file.Close()

		privatePEM, err := io.ReadAll(io.LimitReader(file, 1024*1024))
		if err != nil {
			return fmt.Errorf("reading auth private key: %w", err)
		}

		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
		if err != nil {
			return fmt.Errorf("parsing auth private key: %w", err)
		}

		ks.store[strings.TrimSuffix(dirEntry.Name(), ".pem")] = privateKey
		return nil
	}

	if err := fs.WalkDir(fsys, ".", fn); err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	return &ks, nil
}

func (ks *KeyStore) Add(privateKey *rsa.PrivateKey, kid string) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	ks.store[kid] = privateKey
}

func (ks *KeyStore) Remove(kid string) {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	delete(ks.store, kid)
}

// PrivateKey searches the key store for a given kid and returns
// the private key
func (ks *KeyStore) PrivateKey(kid string) (*rsa.PrivateKey, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	privateKey, found := ks.store[kid]

	if !found {
		return nil, errors.New("kid lookup failed")
	}

	return privateKey, nil
}

// PublicKey searches the key store for a given kid and returns
// the public key
func (ks *KeyStore) PublicKey(kid string) (*rsa.PublicKey, error) {
	ks.mu.RLock()
	defer ks.mu.RUnlock()

	privateKey, found := ks.store[kid]

	if !found {
		return nil, errors.New("kid lookup failed")
	}

	return &privateKey.PublicKey, nil
}

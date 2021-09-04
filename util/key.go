package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path"
)

const DefaultRSABit = 8192

func GenerateKeys(privpath, pubpath string) error {
	// Make key dirs
	err := os.MkdirAll(path.Dir(privpath), 0700)
	if err != nil {
		return err
	}
	err = os.MkdirAll(path.Dir(pubpath), 0700)
	if err != nil {
		return err
	}
	// Create pair of keys
	privkey, err := rsa.GenerateKey(rand.Reader, DefaultRSABit)
	if err != nil {
		return err
	}
	// Save private key
	privkeyraw, err := x509.MarshalPKCS8PrivateKey(privkey)
	if err != nil {
		return err
	}
	privkeypem := pem.Block{Type: "PRIVATE KEY", Bytes: privkeyraw}
	// #nosec G304
	f, err := os.OpenFile(privpath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	err = pem.Encode(f, &privkeypem)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	// Save public keys
	pubkey := privkey.PublicKey
	pubkeyraw, err := x509.MarshalPKIXPublicKey(&pubkey)
	if err != nil {
		return err
	}
	pubkeypem := pem.Block{Type: "PUBLIC KEY", Bytes: pubkeyraw}
	// #nosec G304
	f, err = os.OpenFile(pubpath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	err = pem.Encode(f, &pubkeypem)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

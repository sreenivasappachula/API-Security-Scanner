package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func PrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// 1. Convert private key to DER format (PKCS#1)
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)

	// 2. Create a PEM block
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	// 3. Encode to memory (returns []byte)
	return pem.EncodeToMemory(privateKeyPEM)
}

/*
Copyright 2020 The OpenEBS and Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// generates key-certificate pairs for the
// triple (CA, Server, Client).

package webhook

import (
	"crypto"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"net"
	"time"

	"github.com/pkg/errors"
	certutil "k8s.io/client-go/util/cert"
)

// DefaultEllipticCurve specifies the default elliptic curve to be used for key generation
var DefaultEllipticCurve = elliptic.P256()

const (
	rsaKeySize   = 2048
	duration365d = time.Hour * 24 * 365
	// PrivateKeyBlockType is a possible value for pem.Block.Type.
	PrivateKeyBlockType = "PRIVATE KEY"
	// PublicKeyBlockType is a possible value for pem.Block.Type.
	PublicKeyBlockType = "PUBLIC KEY"
	// CertificateBlockType is a possible value for pem.Block.Type.
	CertificateBlockType = "CERTIFICATE"
	// RSAPrivateKeyBlockType is a possible value for pem.Block.Type.
	RSAPrivateKeyBlockType = "RSA PRIVATE KEY"
)

// KeyPair ...
type KeyPair struct {
	Key  *rsa.PrivateKey
	Cert *x509.Certificate
}

// newPrivateKey returns randomly generated ECDSA private key.
//
//	func newPrivateKey(curve elliptic.Curve) (*ecdsa.PrivateKey, error) {
//		return ecdsa.GenerateKey(curve, rand.Reader)
//	}
//
// newPrivateKey creates an RSA private key
func newPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(cryptorand.Reader, rsaKeySize)
}

// NewCA ...
func NewCA(name string) (*KeyPair, error) {
	key, err := newPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("unable to create a private key for a new CA: %v", err)
	}

	config := certutil.Config{
		CommonName: name,
	}

	cert, err := certutil.NewSelfSignedCACert(config, key)
	if err != nil {
		return nil, fmt.Errorf("unable to create a self-signed certificate for a new CA: %v", err)
	}

	return &KeyPair{
		Key:  key,
		Cert: cert,
	}, nil
}

// NewServerKeyPair ...
func NewServerKeyPair(ca *KeyPair, commonName, svcName, svcNamespace, dnsDomain string, ips, hostnames []string) (*KeyPair, error) {
	key, err := newPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("unable to create a server private key: %v", err)
	}

	namespacedName := fmt.Sprintf("%s.%s", svcName, svcNamespace)
	internalAPIServerFQDN := []string{
		svcName,
		namespacedName,
		fmt.Sprintf("%s.svc", namespacedName),
		fmt.Sprintf("%s.svc.%s", namespacedName, dnsDomain),
	}

	altNames := certutil.AltNames{}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip != nil {
			altNames.IPs = append(altNames.IPs, ip)
		}
	}
	altNames.DNSNames = append(altNames.DNSNames, hostnames...)
	altNames.DNSNames = append(altNames.DNSNames, internalAPIServerFQDN...)

	config := &certutil.Config{
		CommonName: commonName,
		AltNames:   altNames,
		Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	cert, err := NewSignedCert(config, key, ca.Cert, ca.Key)
	if err != nil {
		return nil, fmt.Errorf("unable to sign the server certificate: %v", err)
	}

	return &KeyPair{
		Key:  key,
		Cert: cert,
	}, nil
}

// NewClientKeyPair ...
func NewClientKeyPair(ca *KeyPair, commonName string, organizations []string) (*KeyPair, error) {
	key, err := newPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("unable to create a client private key: %v", err)
	}

	config := &certutil.Config{
		CommonName:   commonName,
		Organization: organizations,
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	cert, err := NewSignedCert(config, key, ca.Cert, ca.Key)
	if err != nil {
		return nil, fmt.Errorf("unable to sign the client certificate: %v", err)
	}

	return &KeyPair{
		Key:  key,
		Cert: cert,
	}, nil
}

// NewSignedCert creates a signed certificate using the given CA certificate and key
func NewSignedCert(cfg *certutil.Config, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, error) {
	serial, err := cryptorand.Int(cryptorand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, err
	}
	if len(cfg.CommonName) == 0 {
		return nil, errors.New("must specify a CommonName")
	}
	if len(cfg.Usages) == 0 {
		return nil, errors.New("must specify at least one ExtKeyUsage")
	}

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:     cfg.AltNames.DNSNames,
		IPAddresses:  cfg.AltNames.IPs,
		SerialNumber: serial,
		NotBefore:    caCert.NotBefore,
		NotAfter:     time.Now().Add(duration365d).UTC(),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  cfg.Usages,
	}
	certDERBytes, err := x509.CreateCertificate(cryptorand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

// EncodeCertPEM returns PEM-endcoded certificate data
func EncodeCertPEM(cert *x509.Certificate) []byte {
	block := pem.Block{
		Type:  CertificateBlockType,
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(&block)
}

// EncodePrivateKeyPEM returns PEM-encoded private key data
func EncodePrivateKeyPEM(key *rsa.PrivateKey) []byte {
	block := pem.Block{
		Type:  RSAPrivateKeyBlockType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.EncodeToMemory(&block)
}

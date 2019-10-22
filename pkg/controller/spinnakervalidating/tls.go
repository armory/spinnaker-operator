package spinnakervalidating

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	certName = "tls.crt"
	keyName  = "tls.key"
	caName   = "ca.crt"

	certificateBlockType = "CERTIFICATE"
	rsaKeySize           = 2048
	duration365d         = time.Hour * 24 * 365
)

type certContext struct {
	cert        []byte
	key         []byte
	signingCert []byte
	certDir     string
}

func setupServerCert(ns string, serviceName string) (*certContext, error) {
	certDir := filepath.Join(os.TempDir(), "spinnaker-operator-certs")
	_ = os.Mkdir(certDir, 0700)
	signingKey, err := newPrivateKey()
	if err != nil {
		return nil, err
	}
	signingCert, err := cert.NewSelfSignedCACert(cert.Config{CommonName: "spinnaker-operator-ca"}, signingKey)
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(filepath.Join(certDir, caName), encodeCertPEM(signingCert), 0644); err != nil {
		return nil, err
	}
	key, err := newPrivateKey()
	if err != nil {
		return nil, err
	}
	signedCert, err := newSignedCert(
		&cert.Config{
			CommonName: serviceName + "." + ns + ".svc",
			Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		},
		key, signingCert, signingKey,
	)
	if err != nil {
		return nil, err
	}
	if err = ioutil.WriteFile(filepath.Join(certDir, certName), encodeCertPEM(signedCert), 0600); err != nil {
		return nil, err
	}
	privateKeyPEM, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return nil, err
	}
	if err = ioutil.WriteFile(filepath.Join(certDir, keyName), privateKeyPEM, 0644); err != nil {
		return nil, err
	}
	return &certContext{
		cert:        encodeCertPEM(signedCert),
		key:         privateKeyPEM,
		signingCert: encodeCertPEM(signingCert),
		certDir:     certDir,
	}, nil
}

// newPrivateKey creates an RSA private key
func newPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, rsaKeySize)
}

// encodeCertPEM returns PEM-encoded certificate data
func encodeCertPEM(cert *x509.Certificate) []byte {
	block := pem.Block{
		Type:  certificateBlockType,
		Bytes: cert.Raw,
	}
	return pem.EncodeToMemory(&block)
}

// newSignedCert creates a signed certificate using the given CA certificate and key
func newSignedCert(cfg *cert.Config, key crypto.Signer, caCert *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, error) {
	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
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
	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTmpl, caCert, key.Public(), caKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(certDERBytes)
}

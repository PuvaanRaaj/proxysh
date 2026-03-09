package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

type DomainCert struct {
	CertPEM []byte
	KeyPEM  []byte
}

func GenerateDomainCert(domain string, ca *CA, daysValid int) (*DomainCert, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: domain,
		},
		DNSNames:  []string{domain, "*." + domain},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().AddDate(0, 0, daysValid),
		KeyUsage:  x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, ca.Cert, &key.PublicKey, ca.Key)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return &DomainCert{CertPEM: certPEM, KeyPEM: keyPEM}, nil
}

func SaveDomainCert(dc *DomainCert, dir, domain string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	certPath := filepath.Join(dir, domain+".crt")
	keyPath := filepath.Join(dir, domain+".key")
	if err := os.WriteFile(certPath, dc.CertPEM, 0644); err != nil {
		return err
	}
	return os.WriteFile(keyPath, dc.KeyPEM, 0600)
}

func LoadDomainCert(dir, domain string) (*DomainCert, error) {
	certPEM, err := os.ReadFile(filepath.Join(dir, domain+".crt"))
	if err != nil {
		return nil, err
	}
	keyPEM, err := os.ReadFile(filepath.Join(dir, domain+".key"))
	if err != nil {
		return nil, err
	}
	return &DomainCert{CertPEM: certPEM, KeyPEM: keyPEM}, nil
}

func DomainCertExists(dir, domain string) bool {
	_, err1 := os.Stat(filepath.Join(dir, domain+".crt"))
	_, err2 := os.Stat(filepath.Join(dir, domain+".key"))
	return err1 == nil && err2 == nil
}

func RemoveDomainCert(dir, domain string) error {
	os.Remove(filepath.Join(dir, domain+".crt"))
	os.Remove(filepath.Join(dir, domain+".key"))
	return nil
}

// LoadTLSKeyPair returns (certPEM, keyPEM) suitable for tls.X509KeyPair
func LoadTLSKeyPair(dir, domain string) ([]byte, []byte, error) {
	dc, err := LoadDomainCert(dir, domain)
	if err != nil {
		return nil, nil, err
	}
	return dc.CertPEM, dc.KeyPEM, nil
}

// EnsureDomainCert generates a domain cert if it doesn't exist yet
func EnsureDomainCert(ca *CA, certDir, domain string, daysValid int) error {
	if DomainCertExists(certDir, domain) {
		return nil
	}
	dc, err := GenerateDomainCert(domain, ca, daysValid)
	if err != nil {
		return err
	}
	return SaveDomainCert(dc, certDir, domain)
}

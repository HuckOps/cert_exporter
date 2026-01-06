package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net"
	"os"
	"time"

	"go.uber.org/zap"
)

type CertificateInfo struct {
	ExpiryDays float64
	Valid      float64
	Subject    string
	SN         string
	Domains    []string
	Algorithm  string
}

func CheckRemoteCertificate(remote string) (time.Duration, []CertificateInfo, error) {
	// Check remote format is host:port
	host, port, err := net.SplitHostPort(remote)
	if err != nil {
		return 0, nil, err
	}

	start := time.Now()

	// Connect to remote host
	conn, err := tls.Dial("tcp", net.JoinHostPort(host, port), &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return 0, nil, err
	}
	defer conn.Close()
	elapsed := time.Since(start)
	// Get certificate
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		zap.L().Error("no certificate found", zap.String("remote", remote))
		err = errors.New("no certificate found")
		return elapsed, nil, err
	}

	var certInfos []CertificateInfo

	for _, cert := range state.PeerCertificates {
		expiryDays := float64(time.Until(cert.NotAfter).Hours() / 24)
		valid := 0
		if expiryDays > 0 {
			valid = 1
		}

		// Skip certificate if it has no DNS names
		if len(cert.DNSNames) == 0 {
			continue
		}
		SN := cert.SerialNumber.String()

		subject := cert.Subject.String()

		certInfo := CertificateInfo{
			ExpiryDays: expiryDays,
			Valid:      float64(valid),
			Subject:    subject,
			SN:         SN,
			Domains:    cert.DNSNames,
			Algorithm:  cert.SignatureAlgorithm.String(),
		}
		certInfos = append(certInfos, certInfo)
	}

	zap.L().Debug("Got certificate info", zap.String("remote", remote), zap.Int("certificate_number", len(certInfos)), zap.Duration("elapsed", elapsed))

	return elapsed, certInfos, nil
}

func CheckLocalCertificate(path string) (*CertificateInfo, error) {
	// Check local certificate file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	// Read certificate file
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(fileBytes)
	if block == nil {
		return nil, errors.New("failed to parse PEM block")
	}
	if block.Type != "CERTIFICATE" {
		return nil, errors.New("PEM block type is not CERTIFICATE")
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	expiryDays := float64(time.Until(cert.NotAfter).Hours() / 24)
	valid := 0
	if expiryDays > 0 {
		valid = 1
	}
	SN := cert.SerialNumber.String()

	subject := cert.Subject.String()

	certInfo := &CertificateInfo{
		ExpiryDays: expiryDays,
		Valid:      float64(valid),
		Subject:    subject,
		SN:         SN,
		Domains:    cert.DNSNames,
		Algorithm:  cert.SignatureAlgorithm.String(),
	}

	return certInfo, nil
}

package main

import (
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type probeResult uint

const (
	certificateOK probeResult = iota
	certificateNotFound
	certificateExpiresSoon
	certificateInvalid
	generalFailure
)

func (p probeResult) String() string {
	switch p {
	case certificateOK:
		return "Certificate OK"
	case certificateExpiresSoon:
		return fmt.Sprintf("Certificate expires within %s", config.ExpireWarning)
	case certificateInvalid:
		return "Certificate invalid / intermediate certificates not present"
	case certificateNotFound:
		return "Did not find a certificate valid for this domain"
	case generalFailure:
		return "Something went wrong in the request"
	}

	return "" // This does not happen.
}

func checkCertificate(probeURL *url.URL) (probeResult, *x509.Certificate) {
	checkLogger := log.WithFields(log.Fields{"probe_url": probeURL})

	req, _ := http.NewRequest("HEAD", probeURL.String(), nil)
	req.Header.Set("User-Agent", fmt.Sprintf("Mozilla/5.0 (compatible; PromCertcheck/%s; +https://github.com/Luzifer/promcertcheck)", version))

	resp, err := http.DefaultClient.Do(req)
	switch {
	case err == nil:
	case strings.Contains(err.Error(), redirectFoundError.Error()):
		checkLogger.WithError(err).Warn("A redirect was found")
	default:
		checkLogger.WithError(err).Error("HTTP request failed")
		return generalFailure, nil
	}
	resp.Body.Close()

	var (
		intermediatePool = x509.NewCertPool()
		verifyCert       *x509.Certificate
	)

	hostPort := strings.Split(probeURL.Host, ":")
	host := hostPort[0]

	for _, cert := range resp.TLS.PeerCertificates {
		wildHost := "*" + host[strings.Index(host, "."):]
		if !inSlice(cert.DNSNames, host) && !inSlice(cert.DNSNames, wildHost) {
			intermediatePool.AddCert(cert)
			continue
		}

		verifyCert = cert
	}

	if verifyCert == nil {
		checkLogger.Debug("Certificate not found")
		return certificateNotFound, nil
	}

	verificationResult := false
	if _, err := verifyCert.Verify(x509.VerifyOptions{
		Intermediates: intermediatePool,
		Roots:         rootPool,
	}); err == nil {
		verificationResult = true
	}

	if !verificationResult {
		checkLogger.Debug("Certificate invalid")
		return certificateInvalid, verifyCert
	}

	if verifyCert.NotAfter.Sub(time.Now()) < config.ExpireWarning {
		checkLogger.Debug("Certificate expires soon")
		return certificateExpiresSoon, verifyCert
	}

	checkLogger.Debug("Certificate OK")
	return certificateOK, verifyCert
}

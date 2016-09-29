package main

import (
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
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
	req, _ := http.NewRequest("HEAD", probeURL.String(), nil)
	req.Header.Set("User-Agent", fmt.Sprintf("Mozilla/5.0 (compatible; PromCertcheck/%s; +https://github.com/Luzifer/promcertcheck)", version))

	resp, err := http.DefaultClient.Do(req)
	switch err.(type) {
	case nil, redirectFoundError:
	default:
		if !strings.Contains(err.Error(), "Found a redirect.") {
			return generalFailure, nil
		}
	}
	resp.Body.Close()

	intermediatePool := x509.NewCertPool()
	var verifyCert *x509.Certificate

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
		return certificateNotFound, nil
	}

	verificationResult := false
	if _, err := verifyCert.Verify(x509.VerifyOptions{
		Intermediates: intermediatePool,
	}); err == nil {
		verificationResult = true
	}

	if !verificationResult {
		return certificateInvalid, verifyCert
	}

	if verifyCert.NotAfter.Sub(time.Now()) < config.expireWarning {
		return certificateExpiresSoon, verifyCert
	}

	return certificateOK, verifyCert
}

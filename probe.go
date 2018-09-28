package main

import (
	"crypto/x509"
	"fmt"
	"net/url"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type probe struct {
	Status      probeResult
	Certificate *x509.Certificate

	isValid prometheus.Gauge
	expires prometheus.Gauge
	url     *url.URL
}

func probeFromURL(u string) (*probe, error) {
	probeURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	p := &probe{
		url: probeURL,
		expires: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "certcheck_expires",
			Help: "Expiration date in unix timestamp (UTC)",
			ConstLabels: prometheus.Labels{
				"host": probeURL.Host,
			},
		}),
		isValid: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "certcheck_valid",
			Help: "Validity of the certificate (0/1)",
			ConstLabels: prometheus.Labels{
				"host": probeURL.Host,
			},
		}),
	}

	prometheus.MustRegister(p.expires)
	prometheus.MustRegister(p.isValid)

	return p, nil
}

func (p *probe) refresh() error {
	verificationResult, verifyCert := checkCertificate(p.url)

	probeLog := log.WithFields(log.Fields{
		"host":   p.url.Host,
		"result": verificationResult,
	})
	if verifyCert != nil {
		probeLog = probeLog.WithFields(log.Fields{
			"version":   verifyCert.Version,
			"serial":    verifyCert.SerialNumber,
			"subject":   verifyCert.Subject.CommonName,
			"expires":   verifyCert.NotAfter,
			"issuer":    verifyCert.Issuer.CommonName,
			"alt_names": strings.Join(verifyCert.DNSNames, ", "),
		})
	}
	probeLog.Debug("Probe finished")

	if err := p.update(verificationResult, verifyCert); err != nil {
		return fmt.Errorf("Unable to update probe state: %s", err)
	}

	return nil
}

func (p *probe) update(status probeResult, cert *x509.Certificate) error {
	p.Status = status
	p.Certificate = cert

	p.updatePrometheus(status, cert)

	return nil
}

func (p probe) updatePrometheus(status probeResult, cert *x509.Certificate) {
	if cert != nil {
		p.expires.Set(float64(cert.NotAfter.UTC().Unix()))
	}

	if status == certificateExpiresSoon || status == certificateOK {
		p.isValid.Set(1)
	} else {
		p.isValid.Set(0)
	}
}

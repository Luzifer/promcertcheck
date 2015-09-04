package main

import (
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Luzifer/rconfig"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron"
)

var (
	config = struct {
		Debug         bool     `flag:"debug" default:"false" description:"Output debugging data"`
		ExpireWarning string   `flag:"expire-warning" default:"744h" description:"When to warn about a soon expiring certificate"`
		Probes        []string `flag:"probe" default:"" description:"URLs to check for certificate issues"`

		expireWarning time.Duration
	}{}
	version       = "dev"
	probeMonitors = map[string]*probeMonitor{}
)

type probeMonitor struct {
	IsValid     prometheus.Gauge
	Expires     prometheus.Gauge
	Status      probeResult
	Certificate *x509.Certificate
}

type redirectFoundError struct{}

func (r redirectFoundError) Error() string {
	return "Found a redirect."
}

func init() {
	var err error

	rconfig.Parse(&config)
	config.expireWarning, err = time.ParseDuration(config.ExpireWarning)
	if err != nil {
		log.Fatalf("You need to specify a valid expire-warning: %s", err)
	}
}

func main() {
	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return redirectFoundError{}
	}

	registerProbes()
	refreshCertificateStatus()

	c := cron.New()
	c.AddFunc("0 0 * * * *", refreshCertificateStatus)
	c.Start()

	r := mux.NewRouter()
	r.Handle("/metrics", prometheus.Handler())
	r.HandleFunc("/", httpHandler)
	http.ListenAndServe(":3000", r)
}

func registerProbes() {
	for _, probe := range config.Probes {
		probeURL, _ := url.Parse(probe)

		monitors := &probeMonitor{}
		monitors.Expires = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "certcheck_expires",
			Help: "Expiration date in unix timestamp (UTC)",
			ConstLabels: prometheus.Labels{
				"host": probeURL.Host,
			},
		})
		monitors.IsValid = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "certcheck_valid",
			Help: "Validity of the certificate (0/1)",
			ConstLabels: prometheus.Labels{
				"host": probeURL.Host,
			},
		})

		prometheus.MustRegister(monitors.Expires)
		prometheus.MustRegister(monitors.IsValid)

		probeMonitors[probeURL.Host] = monitors
	}
}

func refreshCertificateStatus() {
	for _, probe := range config.Probes {
		probeURL, _ := url.Parse(probe)
		verificationResult, verifyCert := checkCertificate(probeURL)

		if config.Debug {
			fmt.Printf("---\nProbe: %s\nResult: %s\n",
				probeURL.Host,
				verificationResult,
			)
			if verifyCert != nil {
				fmt.Printf("Version: %d\nSerial: %d\nSubject: %s\nExpires: %s\nIssuer: %s\nAlt Names: %s\n",
					verifyCert.Version,
					verifyCert.SerialNumber,
					verifyCert.Subject.CommonName,
					verifyCert.NotAfter,
					verifyCert.Issuer.CommonName,
					strings.Join(verifyCert.DNSNames, ", "),
				)
			}
		}

		probeMonitors[probeURL.Host].Expires.Set(float64(verifyCert.NotAfter.UTC().Unix()))
		switch verificationResult {
		case certificateExpiresSoon, certificateOK:
			probeMonitors[probeURL.Host].IsValid.Set(1)
		case certificateInvalid, certificateNotFound:
			probeMonitors[probeURL.Host].IsValid.Set(0)
		default:
			probeMonitors[probeURL.Host].IsValid.Set(0)
		}
		probeMonitors[probeURL.Host].Status = verificationResult
		probeMonitors[probeURL.Host].Certificate = verifyCert
	}
}

func inSlice(slice []string, needle string) bool {
	for _, i := range slice {
		if i == needle {
			return true
		}
	}

	return false
}

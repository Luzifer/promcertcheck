package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Luzifer/rconfig"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

var (
	cfg struct {
		ExpireWarning  time.Duration `flag:"expire-warning" default:"744h" description:"When to warn about a soon expiring certificate"`
		RootsDir       string        `flag:"roots-dir" default:"" description:"Directory to load custom RootCA certs from to be trusted (*.pem)"`
		LogLevel       string        `flag:"log-level" default:"info" description:"Verbosity of logs to use (debug, info, warning, error, ...)"`
		Probes         []string      `flag:"probe" default:"" description:"URLs to check for certificate issues"`
		VersionAndExit bool          `flag:"version" default:"false" description:"Print program version and exit"`
	}

	version = "dev"

	probeMonitors = map[string]*probeMonitor{}
	rootPool      *x509.CertPool

	redirectFoundError = errors.New("Found a redirect")
)

type probeMonitor struct {
	IsValid     prometheus.Gauge `json:"-"`
	Expires     prometheus.Gauge `json:"-"`
	Status      probeResult
	Certificate *x509.Certificate
}

func (p *probeMonitor) Update(status probeResult, cert *x509.Certificate) error {
	p.Status = status
	p.Certificate = cert

	p.updatePrometheus(status, cert)

	return nil
}

func (p probeMonitor) updatePrometheus(status probeResult, cert *x509.Certificate) {
	if cert != nil {
		p.Expires.Set(float64(cert.NotAfter.UTC().Unix()))
	}

	if status == certificateExpiresSoon || status == certificateOK {
		p.IsValid.Set(1)
	} else {
		p.IsValid.Set(0)
	}
}

func init() {
	if err := rconfig.Parse(&cfg); err != nil {
		log.Fatalf("Unable to parse CLI parameters: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("promcertcheck %s\n", version)
		os.Exit(0)
	}

	if logLevel, err := log.ParseLevel(cfg.LogLevel); err == nil {
		log.SetLevel(logLevel)
	} else {
		log.Fatalf("Unable to parse log level: %s", err)
	}
}

func main() {
	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return redirectFoundError
	}
	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	var err error
	if rootPool, err = x509.SystemCertPool(); err != nil {
		log.WithError(err).Fatal("Unable to load system RootCA pool")
	}

	if err = loadAdditionalRootCAPool(); err != nil {
		log.WithError(err).Fatal("Could not load intermediate certificates")
	}

	registerProbes()
	refreshCertificateStatus()

	fmt.Printf("PromCertcheck %s...\nStarting to listen on 0.0.0.0:3000\n", version)

	c := cron.New()
	c.AddFunc("0 0 * * * *", refreshCertificateStatus)
	c.Start()

	r := mux.NewRouter()
	r.Handle("/metrics", prometheus.Handler())
	r.HandleFunc("/", htmlHandler)
	r.HandleFunc("/httpStatus", httpStatusHandler)
	r.HandleFunc("/results.json", jsonHandler)
	http.ListenAndServe(":3000", r)
}

func loadAdditionalRootCAPool() error {
	if cfg.RootsDir == "" {
		// Nothing specified, not loading anything but sys certs
		return nil
	}

	return filepath.Walk(cfg.RootsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".pem") || info.IsDir() {
			// Likely not a certificate, ignore
			return nil
		}

		pem, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		if ok := rootPool.AppendCertsFromPEM(pem); !ok {
			return fmt.Errorf("Failed to load certificate %q", path)
		}

		log.WithFields(log.Fields{"path": path}).Debug("Loaded RootCA certificate")

		return nil
	})
}

func registerProbes() {
	for _, probe := range cfg.Probes {
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
	for _, probe := range cfg.Probes {
		probeURL, _ := url.Parse(probe)
		verificationResult, verifyCert := checkCertificate(probeURL)

		probeLog := log.WithFields(log.Fields{
			"host":   probeURL.Host,
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

		if err := probeMonitors[probeURL.Host].Update(verificationResult, verifyCert); err != nil {
			probeLog.WithError(err).Error("Unable to update probe state")
			return
		}
	}
}

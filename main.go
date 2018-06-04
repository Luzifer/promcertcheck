package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Luzifer/rconfig"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

var (
	cfg struct {
		Listen         string        `flag:"listen" default:":3000" description:"Port/IP to listen on"`
		ExpireWarning  time.Duration `flag:"expire-warning" default:"744h" description:"When to warn about a soon expiring certificate"`
		RootsDir       string        `flag:"roots-dir" default:"" description:"Directory to load custom RootCA certs from to be trusted (*.pem)"`
		LogLevel       string        `flag:"log-level" default:"info" description:"Verbosity of logs to use (debug, info, warning, error, ...)"`
		Probes         []string      `flag:"probe" default:"" description:"URLs to check for certificate issues"`
		VersionAndExit bool          `flag:"version" default:"false" description:"Print program version and exit"`
	}

	version = "dev"

	probeMonitors = map[string]*probe{}
	rootPool      *x509.CertPool

	redirectFoundError = errors.New("Found a redirect")
)

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
	// Configuration to receive redirects and TLS errors
	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return redirectFoundError
	}
	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Load valid CAs from system and specified folder
	var err error
	if rootPool, err = x509.SystemCertPool(); err != nil {
		log.WithError(err).Fatal("Unable to load system RootCA pool")
	}

	if err = loadAdditionalRootCAPool(rootPool); err != nil {
		log.WithError(err).Fatal("Could not load intermediate certificates")
	}

	registerProbes()
	refreshCertificateStatus()

	log.WithFields(log.Fields{
		"version": version,
	}).Info("PromCertcheck started to listen on 0.0.0.0:3000")

	c := cron.New()
	c.AddFunc("0 0 * * * *", refreshCertificateStatus)
	c.Start()

	http.Handle("/metrics", prometheus.Handler())
	http.HandleFunc("/", htmlHandler)
	http.HandleFunc("/httpStatus", httpStatusHandler)
	http.HandleFunc("/results.json", jsonHandler)
	http.ListenAndServe(cfg.Listen, nil)
}

func loadAdditionalRootCAPool(pool *x509.CertPool) error {
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

		if ok := pool.AppendCertsFromPEM(pem); !ok {
			return fmt.Errorf("Failed to load certificate %q", path)
		}

		log.WithFields(log.Fields{"path": path}).Debug("Loaded RootCA certificate")

		return nil
	})
}

func registerProbes() {
	for _, probeURL := range cfg.Probes {
		p, err := probeFromURL(probeURL)
		if err != nil {
			log.WithError(err).Error("Unable to create probe")
			continue
		}

		probeMonitors[p.url.Host] = p
		log.WithFields(log.Fields{
			"host": p.url.Host,
		}).Info("Probe registered")
	}
}

func refreshCertificateStatus() {
	for _, p := range probeMonitors {

		go func(p *probe) {
			logger := log.WithFields(log.Fields{
				"host": p.url.Host,
			})

			if err := p.refresh(); err != nil {
				logger.WithError(err).Error("Unable to refresh probe status")
				return
			}

			logger.Debug("Probe refreshed")
		}(p)

	}
}

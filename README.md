[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/promcertcheck)](https://goreportcard.com/report/github.com/Luzifer/promcertcheck)
![](https://badges.fyi/github/license/Luzifer/promcertcheck)
![](https://badges.fyi/github/downloads/Luzifer/promcertcheck)
![](https://badges.fyi/github/latest-release/Luzifer/promcertcheck)

# Luzifer / PromCertcheck

This project contains a small monitoring tool to check URLs for their certificate validity. The URLs are polled once per hour and the certificates from that URLs are validated against the root certificates available to the program. (Provided by the operating systems distributor or manually set by you if you're using a docker container.)

## Features
- Validates the certification chain including provided intermediate certificates
- Warns before the certificates expires
- Gives a handy overview over all monitored URLs
- Data is made available in Prometheus readable format for monitoring
- Provide own root certificates to accept for chain validation

## Usage

```bash
# ./promcertcheck --help
Usage of ./promcertcheck:
      --expire-warning duration   When to warn about a soon expiring certificate (default 744h0m0s)
      --listen string             Port/IP to listen on (default ":3000")
      --log-level string          Verbosity of logs to use (debug, info, warning, error, ...) (default "info")
      --probe strings             URLs to check for certificate issues
      --roots-dir string          Directory to load custom RootCA certs from to be trusted (*.pem)
      --version                   Print program version and exit

# ./promcertcheck --probe=https://www.google.com/ --probe=https://www.facebook.com/
PromCertcheck dev...
Starting to listen on 0.0.0.0:3000
```

## URLs

| Endpoint | Description |
| ---- | ---- |
| `/` | Shows you a human readable version of the check data |
| `/httpStatus` | Endpoint for simple automated health checks: Delivers `HTTP200` in case everything is fine or `HTTP500` when one or more certificates are broken |
| `/metrics` | Prometheus compatible output of the check data |
| `/results.json` | Gives you a JSON version of the check results including certificate details |

----

![](https://d2o84fseuhwkxk.cloudfront.net/promcertcheck.svg)

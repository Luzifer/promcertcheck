# Luzifer / PromCertcheck

[![License: Apache v2.0](https://badge.luzifer.io/v1/badge?color=5d79b5&title=license&text=Apache+v2.0)](http://www.apache.org/licenses/LICENSE-2.0)

This project contains a small monitoring tool to check URLs for their certificate validity. The URLs are polled once per hour and the certificates from that URLs are validated against the root certificates available to the program. (Provided by the operating systems distributor or manually set by you if you're using a docker container.)

## Features
- Validates the certification chain including provided intermediate certificates
- Warns before the certificates expires
- Gives a handy overview over all monitored URLs
- Data is made available in Prometheus readable format for monitoring

## Usage

```bash
# ./certcheck --help
Usage of ./certcheck:
      --debug[=false]: Output debugging data
      --expire-warning="744h": When to warn about a soon expiring certificate
      --probe=[]: URLs to check for certificate issues

# ./certcheck --probe=https://www.google.com/ --probe=https://www.facebook.com/
PromCertcheck dev...
Starting to listen on 0.0.0.0:3000
```

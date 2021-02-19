# 0.9.0 / 2021-02-19

  * Update Dockerfile for readonly mods
  * Add go modules support
  * Remove old vendoring

# 0.8.0 / 2018-10-01

  * Only display cert info when a cert was found
  * Add error logging for http rendering errors

# 0.7.2 / 2018-09-28

  * Lint: Make go vet happy, remove useless tags

# 0.7.1 / 2018-09-28

  * Update Dockerfile to multi-stage build

# 0.7.0 / 2018-09-28

  * Update dependencies
  * Support configuration through ENV vars

# 0.6.0 / 2018-06-04

  * Improve date format
  * Add alternate names as mouse-over
  * Update vendored libs
  * Fix copyright line in LICENSE

# 0.5.0 / 2017-11-05

  * Allow loading of custom RootCA certificates
  * Add status logging for checks
  * Improve handling of errors for invalid certs
  * Switch to dep for dependency managment, update deps

# 0.4.1 / 2017-06-26

  * Add automated building

# 0.4.0 / 2017-06-26

  * Introduce /httpStatus endpoint  
    which will respond with HTTP200 if everything is fine or HTTP500 if one or more certificates are broken
  * Update rconfig, use duration parsing from rconfig

# 0.3.0 / 2016-09-29

  * Add support for ports in probe-URLs

0.2.1 / 2016-01-26
==================

  * Fix: Only set expiry when valid cert

0.2.0 / 2015-09-04
==================

  * Added JSON output

0.1.1 / 2015-09-04
==================

  * Added documentation

0.1.0 / 2015-09-04
==================

  * Initial version
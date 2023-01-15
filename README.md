# post-step-ca-renewal
A daemon that watches for cert renewals from step-ca and then copies them for use by other services.

![MIT](https://img.shields.io/github/license/joshuar/post-step-ca-renewal)
![GitHub last commit](https://img.shields.io/github/last-commit/joshuar/post-step-ca-renewal)
[![Go Report Card](https://goreportcard.com/badge/github.com/joshuar/post-step-ca-renewal?style=flat-square)](https://goreportcard.com/report/github.com/joshuar/post-step-ca-renewal)
[![Go Reference](https://pkg.go.dev/badge/github.com/joshuar/post-step-ca-renewal.svg)](https://pkg.go.dev/github.com/joshuar/post-step-ca-renewal)
[![Release](https://img.shields.io/github/release/joshuar/post-step-ca-renewal.svg?style=flat-square)](https://github.com/joshuar/post-step-ca-renewal/releases/latest)

## What is it?
A small daemon that watches for updates to certificates renewed using Step CA and then copies them elsewhere to be used by other programs (and optionally running some commands as well).

I created this as I had a number of services that utilised certs and didn't want a bunch of shell scripts all doing the same thing: copying the certs somewhere then restarting some service.

I use Fedora Linux, so that is where testing has been done. It should work with other distributions, YMMV.

## Installation

### Linux Packages (recommended)

RPM/DEB packages are available, see the [releases](https://github.com/joshuar/post-step-ca-renewal/releases) page.

### go get
```shell
go get -u github.com/joshuar/cf-ddns
```

## Contributions

I would welcome your contribution! If you find any improvement or issue you want
to fix, feel free to send a pull request!

## Creator

[Joshua Rich](https://github.com/joshuar) (joshua.rich@gmail.com)

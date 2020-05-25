# keylight [![Linux Test Status](https://github.com/mdlayher/keylight/workflows/Linux%20Test/badge.svg)](https://github.com/mdlayher/keylight/actions) [![GoDoc](https://godoc.org/github.com/mdlayher/keylight?status.svg)](https://godoc.org/github.com/mdlayher/keylight) [![Go Report Card](https://goreportcard.com/badge/github.com/mdlayher/keylight)](https://goreportcard.com/report/github.com/mdlayher/keylight)

Package `keylight` allows control of [Elgato Key Light](https://www.elgato.com/en/gaming/key-light)
devices. MIT Licensed.

## `keylight` CLI

Command `keylight` provides a command-line interface to control Elgato Key
Light devices.

```
$ go install github.com/mdlayher/keylight/cmd/keylight
```

At the moment, the only supported operation is toggling the light state for
a device. The default device address is `http://keylight:9123` which you can
set up as a DNS name or similar for ease of use.

```
$ keylight
device "DEADBEEF9999", light 0 on
$ keylight
device "DEADBEEF9999", light 0 off
```

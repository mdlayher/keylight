# keylight [![Linux Test Status](https://github.com/mdlayher/keylight/workflows/Linux%20Test/badge.svg)](https://github.com/mdlayher/keylight/actions) [![GoDoc](https://godoc.org/github.com/mdlayher/keylight?status.svg)](https://godoc.org/github.com/mdlayher/keylight) [![Go Report Card](https://goreportcard.com/badge/github.com/mdlayher/keylight)](https://goreportcard.com/report/github.com/mdlayher/keylight)

Package `keylight` allows control of [Elgato Key Light](https://www.elgato.com/en/gaming/key-light)
devices. MIT Licensed.

## `keylight` CLI

Command `keylight` provides a command-line interface to control Elgato Key
Light devices.

```
$ go get github.com/mdlayher/keylight/cmd/keylight
```

At the moment, the only supported operation is toggling the light state for
a device. The default device address is `http://keylight:9123` which you can
set up as a DNS name or similar for ease of use.

With no arguments, the device is toggled on and off:

```
$ keylight 
device "keylight", light 0 on: temperature 4200K, brightness 20%
$ keylight 
device "keylight", light 0 off
```

You can also query the device's status or modify its parameters using other flags:

```
$ keylight -h
Usage of keylight:
  -a string
        the address of an Elgato Key Light's HTTP API (default "http://keylight:9123")
  -b int
        set the brightness of a light to the specified percentage (valid: 3-100 %)
  -d string
        set the display name of an Elgato Key Light device
  -i    display the current status of an Elgato Key Light without changing its state
  -t int
        set the color temperature of a light to the specified value (valid: 2900-7000 K)
```

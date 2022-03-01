# keylight [![Test Status](https://github.com/mdlayher/keylight/workflows/Test/badge.svg)](https://github.com/mdlayher/keylight/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/mdlayher/keylight.svg)](https://pkg.go.dev/github.com/mdlayher/keylight)  [![Go Report Card](https://goreportcard.com/badge/github.com/mdlayher/keylight)](https://goreportcard.com/report/github.com/mdlayher/keylight)

Package `keylight` allows control of [Elgato Key Light](https://www.elgato.com/en/gaming/key-light)
devices. MIT Licensed.

## `keylight` CLI

Command `keylight` provides a command-line interface to control Elgato Key
Light devices.

```
$ go install github.com/mdlayher/keylight/cmd/keylight@latest
```

The default device address is `http://keylight:9123` which you can set up as a
DNS name or similar for ease of use. With no arguments, the device is toggled on
and off while retaining its existing brightness and color temperature settings:

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
  -b value
        set brightness to an absolute (between 0 and 100) or relative (-N or +N) percentage
  -d string
        set the display name of an Elgato Key Light device
  -i    display the current status of an Elgato Key Light without changing its state
  -t value
        set temperature to an absolute (between 2900 and 7000) or relative (-N or +N) degrees
```

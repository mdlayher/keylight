// Command keylight provides a command-line interface to control Elgato Key
// Light devices.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/mdlayher/keylight"
)

func main() {
	log.SetFlags(0)

	var (
		addr    = flag.String("a", "http://keylight:9123", "the address of an Elgato Key Light's HTTP API")
		display = flag.String("d", "", "set the display name of an Elgato Key Light device")
		info    = flag.Bool("i", false, "display the current status of an Elgato Key Light without changing its state")
	)
	var brightness, temperature signedNumber
	flag.Var(&brightness, "b", "set brightness to an absolute (between 0 and 100) or relative (-N or +N) percentage")
	flag.Var(&temperature, "t", "set temperature to an absolute (between 2900 and 7000) or relative (-N or +N) degrees")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, err := keylight.NewClient(*addr, nil)
	if err != nil {
		log.Fatalf("failed to create Key Light client: %v", err)
	}

	if *display != "" {
		// Set the device's display name and then force info display to show
		// the updated values.
		if err := c.SetDisplayName(ctx, *display); err != nil {
			log.Fatalf("failed to set display name: %v", err)
		}
		*info = true
	}

	d, err := c.AccessoryInfo(ctx)
	if err != nil {
		log.Fatalf("failed to fetch accessory info: %v", err)
	}

	lights, err := c.Lights(ctx)
	if err != nil {
		log.Fatalf("failed to fetch lights: %v", err)
	}

	if *info {
		// Log info and don't modify any settings.
		logInfo(d, lights)
		return
	}

	// Only toggle the light if no modification flags are set.
	toggle := !brightness.set && !temperature.set

	for _, l := range lights {
		if brightness.relative {
			l.Brightness += brightness.number
		} else if brightness.set {
			l.Brightness = brightness.number
		}
		if temperature.relative {
			l.Temperature += temperature.number
		} else if temperature.set {
			l.Temperature = temperature.number
		}

		if toggle {
			l.On = !l.On
		} else {
			// If the light is being modified, force it on.
			l.On = true
		}
	}

	if err := c.SetLights(ctx, lights); err != nil {
		log.Fatalf("failed to set lights: %v", err)
	}

	logInfo(d, lights)
}

type signedNumber struct {
	set      bool
	relative bool
	number   int
}

func (p signedNumber) String() string {
	if !p.set {
		return ""
	}
	if p.relative {
		return fmt.Sprintf("%+d", p.number)
	}
	return fmt.Sprintf("%d", p.number)
}

func (p *signedNumber) Set(s string) error {
	*p = signedNumber{}
	if s == "" {
		return nil
	}
	p.set = true
	negative := false
	if s[0] == '-' || s[0] == '+' {
		p.relative = true
		negative = s[0] == '-'
		s = s[1:]
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	if negative {
		p.number = -n
	} else {
		p.number = n
	}
	return nil
}

// logInfo logs information about a device and its lights.
func logInfo(d *keylight.Device, ls []*keylight.Light) {
	name := d.DisplayName
	if name == "" {
		name = d.SerialNumber
	}

	for i, l := range ls {
		onOff := "off"
		if l.On {
			onOff = fmt.Sprintf("on: temperature %dK, brightness %d%%",
				l.Temperature, l.Brightness)
		}

		log.Printf("device %q, light %d %s", name, i, onOff)
	}
}

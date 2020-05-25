// Command keylight provides a command-line interface to control Elgato Key
// Light devices.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/mdlayher/keylight"
)

func main() {
	log.SetFlags(0)

	addr := flag.String("addr", "http://keylight:9123", "the address of an Elgato Key Light's HTTP API")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// TODO: build out with more functionality than toggling the light.

	c, err := keylight.NewClient(*addr, nil)
	if err != nil {
		log.Fatalf("failed to create Key Light client: %v", err)
	}

	d, err := c.AccessoryInfo(ctx)
	if err != nil {
		log.Fatalf("failed to fetch accessory info: %v", err)
	}

	lights, err := c.Lights(ctx)
	if err != nil {
		log.Fatalf("failed to fetch lights: %v", err)
	}

	// Toggle the status of each light and report on it.
	for i, l := range lights {
		l.On = !l.On

		name := d.DisplayName
		if name == "" {
			name = d.SerialNumber
		}

		onOff := "off"
		if l.On {
			onOff = "on"
		}

		log.Printf("device %q, light %d %s", name, i, onOff)
	}

	if err := c.SetLights(ctx, lights); err != nil {
		log.Fatalf("failed to set lights: %v", err)
	}
}

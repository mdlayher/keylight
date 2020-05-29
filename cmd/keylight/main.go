// Command keylight provides a command-line interface to control Elgato Key
// Light devices.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/mdlayher/keylight"
)

func main() {
	log.SetFlags(0)

	var (
		addr        = flag.String("a", "http://keylight:9123", "the address of an Elgato Key Light's HTTP API")
		brightness  = flag.Int("b", 0, "set the brightness of a light to the specified percentage (valid: 3-100 %)")
		display     = flag.String("d", "", "set the display name of an Elgato Key Light device")
		info        = flag.Bool("i", false, "display the current status of an Elgato Key Light without changing its state")
		temperature = flag.Int("t", 0, "set the color temperature of a light to the specified value (valid: 2900-7000 K)")
	)
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
	toggle := *brightness == 0 && *temperature == 0

	for _, l := range lights {
		if *brightness != 0 {
			l.Brightness = *brightness
		}
		if *temperature != 0 {
			l.Temperature = *temperature
		}

		if toggle {
			l.On = !l.On
		} else {
			// If the light is being modified, force it on.
			l.On = true
		}
	}

	logInfo(d, lights)

	if err := c.SetLights(ctx, lights); err != nil {
		log.Fatalf("failed to set lights: %v", err)
	}
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

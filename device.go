package keylight

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"
)

const (
	brightnessMin = 3
	brightnessMax = 100
	tempMin       = 2900
	tempMax       = 7000

	tempConstant   = 9900
	tempCoefficent = 20.35
	tempHalfstep   = 25
	tempStep       = 50
)

// A Client can control Elgato Key Light devices.
type Client struct {
	c *http.Client
	u *url.URL
}

// NewClient creates a Client for the Key Light specified by addr. If c is nil,
// a default HTTP client will be configured.
func NewClient(addr string, c *http.Client) (*Client, error) {
	if c == nil {
		c = &http.Client{Timeout: 2 * time.Second}
	}

	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		c: c,
		u: u,
	}, nil
}

// A Device contains metadata about an Elgato Key Light device.
type Device struct {
	ProductName         string `json:"productName,omitempty"`
	FirmwareBuildNumber int    `json:"firmwareBuildNumber,omitempty"`
	FirmwareVersion     string `json:"firmwareVersion,omitempty"`
	SerialNumber        string `json:"serialNumber,omitempty"`
	DisplayName         string `json:"displayName,omitempty"`

	// TODO: add hardwareBoardType, features?
}

// AccessoryInfo fetches information about a Key Light device.
func (c *Client) AccessoryInfo(ctx context.Context) (*Device, error) {
	var d Device
	if err := c.do(ctx, http.MethodGet, "/elgato/accessory-info", nil, &d); err != nil {
		return nil, err
	}

	return &d, nil
}

// SetDisplayName updates the display name for a Key Light device.
func (c *Client) SetDisplayName(ctx context.Context, name string) error {
	b, err := json.Marshal(Device{DisplayName: name})
	if err != nil {
		return err
	}

	return c.do(ctx, http.MethodPut, "/elgato/accessory-info", bytes.NewReader(b), nil)
}

var (
	_ json.Marshaler   = &Light{}
	_ json.Unmarshaler = &Light{}
)

// A Light is the status of an individual light on a Key Light device.
type Light struct {
	// On reports whether the light is currently on or off.
	On bool

	// Brightness is the brightness level of the light with a valid range of 3-100.
	Brightness int

	// Temperature is the light's color temperature with a valid range of 2900-7000K.
	Temperature int
}

// A jsonLight is the raw JSON representation of a Light.
type jsonLight struct {
	On          int `json:"on"`
	Brightness  int `json:"brightness"`
	Temperature int `json:"temperature"`
}

// MarshalJSON implements json.Marshaler.
func (l *Light) MarshalJSON() ([]byte, error) {
	jl := jsonLight{
		Brightness: l.Brightness,
		// The API has its own format but Kelvin is more friendly for users.
		Temperature: convertToAPI(l.Temperature),
	}

	if l.On {
		jl.On = 1
	}

	return json.Marshal(jl)
}

// UnmarshalJSON implements json.Unmarshaler.
func (l *Light) UnmarshalJSON(b []byte) error {
	var jl jsonLight
	if err := json.Unmarshal(b, &jl); err != nil {
		return err
	}

	l.On = jl.On == 1
	l.Brightness = jl.Brightness
	// The API has its own format but Kelvin is more friendly for users.
	l.Temperature = convertToKelvin(jl.Temperature)

	return nil
}

// A lightsBody is the JSON API container for light information.
type lightsBody struct {
	Lights []*Light `json:"lights"`
}

// Lights retrieves the current state of all lights from a Key Light device.
func (c *Client) Lights(ctx context.Context) ([]*Light, error) {
	var body lightsBody
	if err := c.do(ctx, http.MethodGet, "/elgato/lights", nil, &body); err != nil {
		return nil, err
	}

	return body.Lights, nil
}

// SetLights configures the state of all lights on a Key Light device.
func (c *Client) SetLights(ctx context.Context, lights []*Light) error {
	for _, l := range lights {
		if l.Temperature < tempMin || l.Temperature > tempMax {
			return fmt.Errorf("temperature (%d) out of range 2900 <= x <= 7000", l.Temperature)
		}

		if l.Brightness < brightnessMin || l.Brightness > brightnessMax {
			return fmt.Errorf("brightness (%d) out of range 3 <= x <= 100", l.Brightness)
		}
	}

	// This structure is small enough where marshaling the whole thing in memory
	// is not a concern.
	b, err := json.Marshal(lightsBody{Lights: lights})
	if err != nil {
		return err
	}

	var body lightsBody
	if err := c.do(ctx, http.MethodPut, "/elgato/lights", bytes.NewReader(b), &body); err != nil {
		return err
	}

	// The device will ignore configuration for any lights which do not exist,
	// but we treat this as an error because the caller should only attempt to
	// configure the number of lights present on the device.
	if len(body.Lights) != len(lights) {
		return fmt.Errorf("keylight: attempted to configure %d lights, but %d are present",
			len(lights), len(body.Lights))
	}

	return nil
}

// do performs an HTTP request with the input parameters, optionally
// unmarshaling a JSON body into out if out is not nil.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader, out interface{}) error {
	// Make a copy of c.u before manipulating the path to avoid modifying the
	// base URL.
	u := *c.u
	u.Path = path

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return err
	}

	res, err := c.c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("keylight: device returned HTTP %d", res.StatusCode)
	}

	if out == nil {
		// No struct passed to unmarshal from JSON, exit early.
		return nil
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return err
	}

	return nil
}

// convertToKelvin converts the Elgato API temperatures to Kelvin.
func convertToKelvin(elgato int) int {
	kelvin := tempConstant - int(math.Round(float64(elgato)*tempCoefficent))
	remainder := kelvin % tempStep
	if remainder > tempHalfstep {
		return kelvin + tempStep - remainder
	}
	return kelvin - remainder
}

// convertToAPI converts Kelvin temperatures to those of the Elgato API.
func convertToAPI(kelvin int) int {
	elgato := float64(kelvin-tempConstant) / tempCoefficent
	return int(math.Abs(math.Trunc(elgato)))
}

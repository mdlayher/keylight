package keylight_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/keylight"
)

func TestClientAccessoryInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	want := &keylight.Device{
		ProductName:         "Elgato Key Light",
		FirmwareBuildNumber: 192,
		FirmwareVersion:     "1.0.3",
		SerialNumber:        "ABCDEFGHIJKL",
		DisplayName:         "Office",
	}

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
			panicf("unexpected HTTP method (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff("/elgato/accessory-info", r.URL.Path); diff != "" {
			panicf("unexpected URL path (-want +got):\n%s", diff)
		}

		_ = json.NewEncoder(w).Encode(want)
	})

	got, err := c.AccessoryInfo(ctx)
	if err != nil {
		t.Fatalf("failed to fetch device: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected device (-want +got):\n%s", diff)
	}
}

func TestClientLights(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	want := []*keylight.Light{{
		On:          true,
		Brightness:  15,
		Temperature: 3400,
	}}

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if diff := cmp.Diff(http.MethodGet, r.Method); diff != "" {
			panicf("unexpected HTTP method (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff("/elgato/lights", r.URL.Path); diff != "" {
			panicf("unexpected URL path (-want +got):\n%s", diff)
		}

		// These structures match the raw output from the API. To avoid
		// exporting unnecessary types, we make copies of their definitions
		// here.
		v := struct {
			NumberOfLights int               `json:"numberOfLights"`
			Lights         []*keylight.Light `json:"lights"`
		}{
			NumberOfLights: len(want),
			Lights:         want,
		}

		_ = json.NewEncoder(w).Encode(v)
	})

	got, err := c.Lights(ctx)
	if err != nil {
		t.Fatalf("failed to fetch lights: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected device (-want +got):\n%s", diff)
	}
}

func TestClientSetLights(t *testing.T) {
	light := &keylight.Light{
		On:          true,
		Brightness:  15,
		Temperature: 2900,
	}

	tests := []struct {
		name   string
		lights []*keylight.Light
		fn     http.HandlerFunc
		check  func(t *testing.T, err error)
	}{
		{
			name: "bad number of lights",
			lights: []*keylight.Light{
				light,
				// This light doesn't actually exist and should prompt the client
				// to return an error.
				light,
			},
			check: func(t *testing.T, err error) {
				if !strings.Contains(err.Error(), "attempted to configure 2 lights, but 1 are present") {
					t.Fatalf("error did not mention malformed lights input: %v", err)
				}
			},
		},
		{
			name:   "OK",
			lights: []*keylight.Light{light},
		},
		{
			name: "temperature outside of range",
			lights: []*keylight.Light{
				{On: true, Temperature: 2899, Brightness: 15},
			},
			check: func(t *testing.T, err error) {
				if !strings.Contains(err.Error(), "temperature (2899) out of range 2900 <= x <= 7000") {
					t.Fatalf("error did not mention malformed temperature input: %v", err)
				}
			},
		},
		{
			name: "brightness outside of range",
			lights: []*keylight.Light{
				{On: true, Temperature: 2900, Brightness: 101},
			},
			check: func(t *testing.T, err error) {
				if !strings.Contains(err.Error(), "brightness (101) out of range 3 <= x <= 100") {
					t.Fatalf("error did not mention malformed brightness input: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
			defer cancel()

			c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
				if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
					panicf("unexpected HTTP method (-want +got):\n%s", diff)
				}

				if diff := cmp.Diff("/elgato/lights", r.URL.Path); diff != "" {
					panicf("unexpected URL path (-want +got):\n%s", diff)
				}

				var v struct {
					NumberOfLights int               `json:"numberOfLights"`
					Lights         []*keylight.Light `json:"lights"`
				}

				if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
					panicf("failed to decode JSON: %v", err)
				}

				// Only return a single light to the caller to mimic the
				// behavior of the Key Light when you attempt to configure
				// multiple lights which may not exist.
				v.Lights = v.Lights[:1]
				_ = json.NewEncoder(w).Encode(v)
			})

			err := c.SetLights(ctx, tt.lights)
			if err == nil && tt.check != nil {
				t.Fatal("an error was expected, but none occurred")
			}
			if err != nil {
				tt.check(t, err)
			}
		})
	}
}

func TestClientErrors(t *testing.T) {
	tests := []struct {
		name  string
		fn    http.HandlerFunc
		check func(t *testing.T, err error)
	}{
		{
			name: "HTTP 404",
			fn:   http.NotFound,
			check: func(t *testing.T, err error) {
				if !strings.Contains(err.Error(), "HTTP 404") {
					t.Fatalf("error text did not contain HTTP 404: %v", err)
				}
			},
		},
		{
			name: "context canceled",
			fn: func(w http.ResponseWriter, r *http.Request) {
				// The client's context will be canceled while this sleep is
				// occurring.
				time.Sleep(500 * time.Millisecond)

				if err := r.Context().Err(); !errors.Is(err, context.Canceled) {
					panicf("expected context canceled, but got: %v", err)
				}
			},
			check: func(t *testing.T, err error) {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("expected context deadline exceeded, but got: %v", err)
				}
			},
		},
		{
			name: "malformed JSON",
			fn: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte{0xff})
			},
			check: func(t *testing.T, err error) {
				var jerr *json.SyntaxError
				if !errors.As(err, &jerr) {
					t.Fatalf("expected JSON syntax error, but got: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
			defer cancel()

			c := testClient(t, tt.fn)

			// We are testing against the exported API for conciseness but
			// ultimately this test only really cares about various error
			// conditions, so we will ignore the Device return value entirely.
			_, err := c.AccessoryInfo(ctx)
			if err == nil && tt.check != nil {
				t.Fatal("an error was expected, but none occurred")
			}

			tt.check(t, err)
		})
	}
}

// testClient creates a *keylight.Client pointed at the HTTP server running with
// handler fn.
func testClient(t *testing.T, fn http.HandlerFunc) *keylight.Client {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if fn != nil {
			fn(w, r)
		}
	}))

	t.Cleanup(func() {
		srv.Close()
	})

	c, err := keylight.NewClient(srv.URL, nil)
	if err != nil {
		t.Fatalf("failed to create keylight client: %v", err)
	}

	return c

}

func panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}

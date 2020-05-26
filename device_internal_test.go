package keylight

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestLightTemperatureConversion(t *testing.T) {
	tests := []struct {
		kelvin, elgato int
	}{
		{
			kelvin: 2900,
			elgato: 343,
		},
		{
			kelvin: 7000,
			elgato: 142,
		},
		{
			kelvin: 3800,
			elgato: 300,
		},
		{
			kelvin: 3750,
			elgato: 301,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%dK", tt.kelvin), func(t *testing.T) {
			kelvin := convertToKelvin(tt.elgato)
			if diff := cmp.Diff(tt.kelvin, kelvin); diff != "" {
				t.Fatalf("unexpected temperature Kelvin value(-want +got):\n%s", diff)
			}

			el := convertToAPI(kelvin)
			if diff := cmp.Diff(float64(tt.elgato), float64(el), cmpopts.EquateApprox(0, 1)); diff != "" {
				t.Fatalf("unexpected temperature Elgato value(-want +got):\n%s", diff)
			}
		})
	}
}

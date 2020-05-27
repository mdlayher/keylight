package keylight_test

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/keylight"
)

func TestClientSetWiFiInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	var (
		device = &keylight.Device{
			ProductName:         "Elgato Key Light",
			FirmwareBuildNumber: 192,
			FirmwareVersion:     "1.0.3",
			SerialNumber:        "ABCDEFGHIJKL",
			DisplayName:         "Office",
			HardwareBoardType:   42,
		}

		want = &keylight.WiFiInfo{
			SSID:         "Elgato SSID",
			Passphrase:   "Elgato",
			SecurityType: keylight.WPA,
		}

		got *keylight.WiFiInfo
	)

	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if diff := cmp.Diff(http.MethodPut, r.Method); diff != "" {
			panicf("unexpected HTTP method (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff("/elgato/wifi-info", r.URL.Path); diff != "" {
			panicf("unexpected URL path (-want +got):\n%s", diff)
		}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		got, err = decrypt(data, device)
		if err != nil {
			panic(err)
		}
	})

	if err := c.SetWiFiInfo(ctx, want, device); err != nil {
		t.Fatalf("failed to set WiFi info: %v", err)
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("unexpected WiFiInfo(-want +got):\n%s", diff)
	}
}

func decrypt(ciphertext []byte, device *keylight.Device) (*keylight.WiFiInfo, error) {
	key := keylight.AESKey(device.HardwareBoardType, device.FirmwareBuildNumber)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("keylight: WiFi Info ciphertext is shorter than one aes.BlockSize")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("keylight: WiFi Info ciphertext is not a multiple of aes.BlockSize")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	// Unpad the zero padded plaintext by searching for the first null byte. If
	// we don't find one, then we expect the message to be an exact multiple of
	// aes.BlockSize.
	idx := bytes.IndexByte(ciphertext, 0)
	if idx < 0 {
		idx = len(ciphertext)
	}

	wifi := &keylight.WiFiInfo{}
	if err = json.Unmarshal(ciphertext[:idx], wifi); err != nil {
		return nil, err
	}

	return wifi, nil
}

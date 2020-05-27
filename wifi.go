package keylight

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
)

// pathWiFiInfo indicates an unusual request is required compared to standard
// JSON endpoints.
const pathWiFiInfo = "/elgato/wifi-info"

// A WiFiInfo contains the information neccessary for a Key Light to connect
// to a wireless network.
type WiFiInfo struct {
	SSID         string       `json:"ssid,omitempty"`
	Passphrase   string       `json:"passphrase,omitempty"`
	SecurityType WiFiSecurity `json:"securityType,omitempty"`
}

// WiFiSecurity is the security type of the wireless network.
//
// The Elgato Key Light supports none, WEP, and WPA/WPA2 Personal
type WiFiSecurity int

// Possible WiFiSecurity values for use in a WiFiInfo.
const (
	None WiFiSecurity = 0
	WEP  WiFiSecurity = 1
	WPA  WiFiSecurity = 2
)

// SetWiFiInfo updates a Key Light's WiFi configuration.
func (c *Client) SetWiFiInfo(ctx context.Context, wifi *WiFiInfo, device *Device) error {
	b, err := json.Marshal(wifi)
	if err != nil {
		return err
	}

	// Zero pad the plaintext to aes.BlockSize.
	blen := len(b)
	padlen := aes.BlockSize - (blen % aes.BlockSize)
	pad := bytes.Repeat([]byte{0}, padlen)

	plaintext := append(b, pad...)
	key := aesKey(device.HardwareBoardType, device.FirmwareBuildNumber)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	return c.do(ctx, http.MethodPut, pathWiFiInfo, bytes.NewReader(ciphertext), nil)
}

// aesKey returns the AES key based on the device's board type and firmware
// build number.
func aesKey(boardType, firmwareBuildNumber int) []byte {
	// Key generation code fetched from the Elgato Control Center application.
	return []byte{
		76, 180, byte(boardType >> 0), byte(boardType >> 8),
		176, 234, 221, 238,
		235, 42, 3, 138,
		49, byte(firmwareBuildNumber >> 0), byte(firmwareBuildNumber >> 8), 86,
	}
}

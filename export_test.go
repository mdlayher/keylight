package keylight

// AESKey exports aesKey for tests.
func AESKey(boardType, firmwareBuildNumber int) []byte {
	return aesKey(boardType, firmwareBuildNumber)
}

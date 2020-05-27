package keylight

// Constants exported for tests.
const (
	ContentBinary = contentBinary
	ContentJSON   = contentJSON
)

// AESKey exports aesKey for tests.
func AESKey(boardType, firmwareBuildNumber int) []byte {
	return aesKey(boardType, firmwareBuildNumber)
}

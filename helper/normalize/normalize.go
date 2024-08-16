package normalize

import (
	"regexp"
	"strings"
)

// processInput memproses input menjadi huruf kecil, menghapus spasi,
// dan hanya memperbolehkan karakter a-z, 0-9, _ dan -
func SetIntoID(input string) string {
	// Konversi ke huruf kecil
	input = strings.ToLower(input)

	// Hapus spasi
	input = strings.ReplaceAll(input, " ", "")

	// Hapus karakter khusus kecuali _ dan -
	re := regexp.MustCompile(`[^a-z0-9_-]`)
	input = re.ReplaceAllString(input, "")

	return input
}

func removeInvisibleChars(text string) string {
	// Create a regular expression to match invisible characters
	re := regexp.MustCompile(`\p{C}`)

	// Replace all matches with an empty string
	return re.ReplaceAllString(text, "")
}

func removeZeroWidthSpaces(text string) string {
	// Create a regular expression to match specific zero-width characters
	re := regexp.MustCompile(`\p{Cf}`)

	// Replace all matches with an empty string
	return re.ReplaceAllString(text, "")
}

func NormalizeHiddenChar(text string) string {
	return removeZeroWidthSpaces(removeInvisibleChars(text))
}

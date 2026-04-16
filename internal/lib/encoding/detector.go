package encoding

import (
	"bytes"
	"io"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// DetectAndConvert attempts to detect the encoding of the input and convert to UTF-8
func DetectAndConvert(r io.Reader) (*bytes.Buffer, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// Check if already valid UTF-8
	if utf8.Valid(data) {
		return bytes.NewBuffer(data), nil
	}

	// Try ISO-8859-1 (Latin-1) conversion
	isoDecoder := charmap.ISO8859_1.NewDecoder()
	isoConverted, _, err := transform.Bytes(isoDecoder, data)
	if err == nil && utf8.Valid(isoConverted) {
		return bytes.NewBuffer(isoConverted), nil
	}

	// Try Windows-1252 conversion
	winDecoder := charmap.Windows1252.NewDecoder()
	winConverted, _, err := transform.Bytes(winDecoder, data)
	if err == nil && utf8.Valid(winConverted) {
		return bytes.NewBuffer(winConverted), nil
	}

	// Fallback: return as-is
	return bytes.NewBuffer(data), nil
}

// Package compression provides HTTP compression utilities.
package compression

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strings"
)

// SupportsGzip checks if the client accepts gzip encoding.
// It inspects the Accept-Encoding header and returns true if gzip is supported.
func SupportsGzip(r *http.Request) bool {
	acceptEncoding := r.Header.Get("Accept-Encoding")
	return strings.Contains(acceptEncoding, "gzip")
}

// IsGzipped checks if data is already gzip compressed.
// It checks for the gzip magic bytes (0x1f 0x8b) at the start of the data.
func IsGzipped(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

// CompressGzip compresses data with the specified compression level.
// The level parameter should be one of the gzip constants (e.g., gzip.BestSpeed).
// Returns the compressed data or an error if compression fails.
func CompressGzip(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

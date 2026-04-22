package compression

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http/httptest"
	"testing"
)

// TestSupportsGzip tests detection of gzip support in Accept-Encoding header.
func TestSupportsGzip(t *testing.T) {
	tests := []struct {
		name          string
		acceptHeader  string
		expectSupport bool
	}{
		{
			name:          "with gzip",
			acceptHeader:  "gzip",
			expectSupport: true,
		},
		{
			name:          "with gzip and deflate",
			acceptHeader:  "gzip, deflate",
			expectSupport: true,
		},
		{
			name:          "with gzip and other encodings",
			acceptHeader:  "gzip, deflate, br",
			expectSupport: true,
		},
		{
			name:          "without gzip",
			acceptHeader:  "deflate",
			expectSupport: false,
		},
		{
			name:          "empty header",
			acceptHeader:  "",
			expectSupport: false,
		},
		{
			name:          "with br only",
			acceptHeader:  "br",
			expectSupport: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.acceptHeader != "" {
				req.Header.Set("Accept-Encoding", tt.acceptHeader)
			}

			result := SupportsGzip(req)
			if result != tt.expectSupport {
				t.Errorf("expected %v, got %v for Accept-Encoding: %s",
					tt.expectSupport, result, tt.acceptHeader)
			}
		})
	}
}

// TestIsGzipped tests detection of gzip magic bytes.
func TestIsGzipped(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{
			name:     "gzipped data",
			data:     []byte{0x1f, 0x8b, 0x08, 0x00},
			expected: true,
		},
		{
			name:     "non-gzipped data",
			data:     []byte{0x00, 0x01, 0x02, 0x03},
			expected: false,
		},
		{
			name:     "empty data",
			data:     []byte{},
			expected: false,
		},
		{
			name:     "single byte",
			data:     []byte{0x1f},
			expected: false,
		},
		{
			name:     "only first byte matches",
			data:     []byte{0x1f, 0x00},
			expected: false,
		},
		{
			name:     "only second byte matches",
			data:     []byte{0x00, 0x8b},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGzipped(tt.data)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for data %v",
					tt.expected, result, tt.data)
			}
		})
	}
}

// TestCompressGzip tests gzip compression functionality.
func TestCompressGzip(t *testing.T) {
	tests := []struct {
		name  string
		data  []byte
		level int
	}{
		{
			name:  "small data with best speed",
			data:  []byte("Hello, World!"),
			level: gzip.BestSpeed,
		},
		{
			name:  "small data with best compression",
			data:  []byte("Hello, World!"),
			level: gzip.BestCompression,
		},
		{
			name:  "large data",
			data:  bytes.Repeat([]byte("test data "), 1000),
			level: gzip.BestSpeed,
		},
		{
			name:  "empty data",
			data:  []byte{},
			level: gzip.BestSpeed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := CompressGzip(tt.data, tt.level)
			if err != nil {
				t.Fatalf("compression failed: %v", err)
			}

			// Verify it's gzipped
			if !IsGzipped(compressed) {
				t.Error("compressed data doesn't have gzip magic bytes")
			}

			// Verify we can decompress it
			reader, err := gzip.NewReader(bytes.NewReader(compressed))
			if err != nil {
				t.Fatalf("failed to create gzip reader: %v", err)
			}
			defer reader.Close()

			decompressed, err := io.ReadAll(reader)
			if err != nil {
				t.Fatalf("failed to decompress: %v", err)
			}

			// Verify decompressed matches original
			if !bytes.Equal(decompressed, tt.data) {
				t.Errorf("decompressed data doesn't match original")
			}

			// Verify compression actually reduces size (for non-empty data)
			if len(tt.data) > 100 {
				if len(compressed) >= len(tt.data) {
					t.Logf("warning: compressed size (%d) >= original size (%d)",
						len(compressed), len(tt.data))
				}
			}
		})
	}
}

// TestCompressGzip_InvalidLevel tests behavior with invalid compression level.
func TestCompressGzip_InvalidLevel(t *testing.T) {
	data := []byte("test data")
	invalidLevel := 100 // Invalid compression level

	_, err := CompressGzip(data, invalidLevel)
	if err == nil {
		t.Error("expected error for invalid compression level, got nil")
	}
}

// TestCompressGzip_CompressionRatio tests that realistic tile data compresses well.
func TestCompressGzip_CompressionRatio(t *testing.T) {
	// Simulate MVT tile data - highly compressible due to repeated structures
	tileData := bytes.Repeat([]byte("feature{id:1,type:road,name:main_st,geometry:[0,0,1,1]}"), 100)

	tests := []struct {
		name         string
		level        int
		minRatio     float64 // minimum expected compression ratio
	}{
		{
			name:     "best speed",
			level:    gzip.BestSpeed,
			minRatio: 2.0, // expect at least 2x compression
		},
		{
			name:     "best compression",
			level:    gzip.BestCompression,
			minRatio: 3.0, // expect at least 3x compression
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := CompressGzip(tileData, tt.level)
			if err != nil {
				t.Fatalf("compression failed: %v", err)
			}

			ratio := float64(len(tileData)) / float64(len(compressed))
			if ratio < tt.minRatio {
				t.Errorf("compression ratio %.2fx is below minimum %.2fx",
					ratio, tt.minRatio)
			}

			t.Logf("compression ratio: %.2fx (original: %d, compressed: %d)",
				ratio, len(tileData), len(compressed))
		})
	}
}

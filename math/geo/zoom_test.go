package geo_test

import (
	"testing"

	"github.com/swayrider/swlib/math/floats"
	"github.com/swayrider/swlib/math/geo"
)

func TestZoom2Radius(t *testing.T) {
	tests := []struct {
		name     string
		zoom     int
		expected uint32
	}{
		{"zoom 0", 0, 33554432},
		{"zoom 1", 1, 16777216},
		{"zoom 5", 5, 1048576},
		{"zoom 10", 10, 32768},
		{"zoom 15", 15, 1024},
		{"zoom 18", 18, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := geo.Zoom2Radius(tt.zoom)
			if result != tt.expected {
				t.Errorf("Zoom2Radius(%d) = %d, want %d", tt.zoom, result, tt.expected)
			}
		})
	}
}

func TestZoom2Radius_NegativeZoom(t *testing.T) {
	result := geo.Zoom2Radius(-1)
	expected := geo.Zoom2Radius(0)
	if result != expected {
		t.Errorf("Zoom2Radius(-1) = %d, want %d (should clamp to zoom 0)", result, expected)
	}

	result = geo.Zoom2Radius(-100)
	if result != expected {
		t.Errorf("Zoom2Radius(-100) = %d, want %d (should clamp to zoom 0)", result, expected)
	}
}

func TestZoom2Radius_ZoomAbove18(t *testing.T) {
	result := geo.Zoom2Radius(19)
	expected := geo.Zoom2Radius(18)
	if result != expected {
		t.Errorf("Zoom2Radius(19) = %d, want %d (should clamp to zoom 18)", result, expected)
	}

	result = geo.Zoom2Radius(100)
	if result != expected {
		t.Errorf("Zoom2Radius(100) = %d, want %d (should clamp to zoom 18)", result, expected)
	}
}

func TestZoom2RadiusKm(t *testing.T) {
	tests := []struct {
		name     string
		zoom     int
		expected float64
	}{
		{"zoom 0", 0, 33554.432},
		{"zoom 10", 10, 32.768},
		{"zoom 18", 18, 0.128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := geo.Zoom2RadiusKm(tt.zoom)
			if floats.Compare64(result, tt.expected) != 0 {
				t.Errorf("Zoom2RadiusKm(%d) = %v, want %v", tt.zoom, result, tt.expected)
			}
		})
	}
}

func TestZoom2RadiusKm_ConsistentWithZoom2Radius(t *testing.T) {
	for zoom := 0; zoom <= 18; zoom++ {
		radiusMeters := geo.Zoom2Radius(zoom)
		radiusKm := geo.Zoom2RadiusKm(zoom)
		expectedKm := float64(radiusMeters) / 1000.0

		if floats.Compare64(radiusKm, expectedKm) != 0 {
			t.Errorf("Zoom2RadiusKm(%d) = %v, want %v (should equal Zoom2Radius/1000)", zoom, radiusKm, expectedKm)
		}
	}
}

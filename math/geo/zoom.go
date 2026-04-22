// Package geo provides geographic utility functions for map-related calculations.
//
// # Zoom Level to Radius Conversion
//
// Converts map zoom levels (0-18) to search radii:
//
//	Zoom 0:  ~33.5km (world view)
//	Zoom 5:  ~1km    (country)
//	Zoom 10: ~32m    (city)
//	Zoom 15: ~1m     (street)
//	Zoom 18: ~0.1m   (building)
package geo

// z0 is the base radius in meters at zoom level 0
const z0 = uint32(0x2000000)

// Zoom2Radius converts a map zoom level to a search radius in meters.
// Zoom levels are clamped to the range 0-18.
// Higher zoom levels result in smaller radii (more zoomed in = smaller area).
//
// Example:
//
//	radius := geo.Zoom2Radius(10) // Returns radius for city-level zoom
func Zoom2Radius(zoom int) uint32 {
	if zoom < 0 {
		zoom = 0
	}
	if zoom > 18 {
		zoom = 18
	}
	return z0 >> zoom
}

// Zoom2RadiusKm converts a map zoom level to a search radius in kilometers.
// This is a convenience wrapper around Zoom2Radius that returns kilometers.
func Zoom2RadiusKm(zoom int) float64 {
	return float64(Zoom2Radius(zoom)) / 1000.0
}

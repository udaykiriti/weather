package weather

// DewPointComfort returns a human-readable comfort label for a dew-point
// temperature in °C. Higher dew points feel increasingly muggy and oppressive.
// Thresholds follow the widely-used meteorological comfort scale.
func DewPointComfort(dewPointC float64) string {
	switch {
	case dewPointC < 10:
		return "Dry"
	case dewPointC < 13:
		return "Comfortable"
	case dewPointC < 16:
		return "Slightly humid"
	case dewPointC < 18:
		return "Humid"
	case dewPointC < 21:
		return "Very humid"
	default:
		return "Oppressive"
	}
}

// HumidityLabel returns a short label for a relative-humidity percentage.
func HumidityLabel(humidity int) string {
	switch {
	case humidity < 25:
		return "Very dry"
	case humidity < 40:
		return "Dry"
	case humidity < 60:
		return "Comfortable"
	case humidity < 75:
		return "Humid"
	case humidity < humidityHigh:
		return "Very humid"
	default:
		return "Oppressive"
	}
}

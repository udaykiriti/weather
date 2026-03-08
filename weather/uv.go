package weather

import (
	"fmt"
	"strings"
)

// UVLevel returns a short label for a UV index value.
func UVLevel(uv float64) string {
	switch {
	case uv < uvLow:
		return "Low"
	case uv < uvModerate:
		return "Moderate"
	case uv < uvHigh:
		return "High"
	case uv < uvVeryHigh:
		return "Very High"
	default:
		return "Extreme"
	}
}

// UVAdvice returns sun-protection advice for a UV index value.
func UVAdvice(uv float64) string {
	switch {
	case uv < uvLow:
		return "No protection needed. Enjoy the sun safely."
	case uv < uvModerate:
		return "Wear sunscreen SPF 30+. Hat recommended."
	case uv < uvHigh:
		return "SPF 50+ sunscreen, hat and sunglasses. Seek shade 11am–3pm."
	case uv < uvVeryHigh:
		return "SPF 50+ and protective clothing essential. Minimize sun exposure."
	default:
		return "Extreme UV. Stay indoors if possible. Full protection required."
	}
}

// UVColorClass returns a CSS class name for the UV level badge color.
func UVColorClass(uv float64) string {
	switch {
	case uv < uvLow:
		return "uv-low"
	case uv < uvModerate:
		return "uv-moderate"
	case uv < uvHigh:
		return "uv-high"
	case uv < uvVeryHigh:
		return "uv-veryhigh"
	default:
		return "uv-extreme"
	}
}

// UVBar returns a filled/empty string progress bar (0-12 scale) for CLI.
func UVBar(uv float64, width int) string {
	n := int(uv / 12.0 * float64(width))
	if n > width {
		n = width
	}
	return fmt.Sprintf("[%s%s] %.1f", strings.Repeat("█", n), strings.Repeat("░", width-n), uv)
}

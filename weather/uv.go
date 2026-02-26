package weather

import "fmt"

// UVLevel returns a short label for a UV index value.
func UVLevel(uv float64) string {
	switch {
	case uv < 3:
		return "Low"
	case uv < 6:
		return "Moderate"
	case uv < 8:
		return "High"
	case uv < 11:
		return "Very High"
	default:
		return "Extreme"
	}
}

// UVAdvice returns sun-protection advice for a UV index value.
func UVAdvice(uv float64) string {
	switch {
	case uv < 3:
		return "No protection needed. Enjoy the sun safely."
	case uv < 6:
		return "Wear sunscreen SPF 30+. Hat recommended."
	case uv < 8:
		return "SPF 50+ sunscreen, hat and sunglasses. Seek shade 11am–3pm."
	case uv < 11:
		return "SPF 50+ and protective clothing essential. Minimize sun exposure."
	default:
		return "Extreme UV. Stay indoors if possible. Full protection required."
	}
}

// UVColorClass returns a CSS class name for the UV level badge color.
func UVColorClass(uv float64) string {
	switch {
	case uv < 3:
		return "uv-low"
	case uv < 6:
		return "uv-moderate"
	case uv < 8:
		return "uv-high"
	case uv < 11:
		return "uv-veryhigh"
	default:
		return "uv-extreme"
	}
}

// UVBar returns a filled/empty string progress bar (0-12 scale) for CLI.
func UVBar(uv float64, width int) string {
	max := 12.0
	n := int(uv / max * float64(width))
	if n > width {
		n = width
	}
	bar := ""
	for i := 0; i < width; i++ {
		if i < n {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return fmt.Sprintf("[%s] %.1f", bar, uv)
}

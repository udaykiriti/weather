package weather

import "strconv"

// OutfitItem represents a single clothing or accessory suggestion.
type OutfitItem struct {
	Icon  string // SVG path data or emoji fallback
	Label string // Short label shown under icon
	Note  string // One-line reason / tip
	Color string // CSS class for card accent
}

// OutfitAdvice holds the full outfit recommendation for a weather snapshot.
type OutfitAdvice struct {
	Headline string       // e.g. "Layer up — cold and wet"
	Items    []OutfitItem // 3–6 items
	TempTier string       // "freezing" | "cold" | "cool" | "mild" | "warm" | "hot"
}

// BuildOutfit generates outfit suggestions from current conditions.
// tempC is always in Celsius (converted internally if needed), units is "metric"|"imperial".
func BuildOutfit(info *WeatherInfo) OutfitAdvice {
	cur := info.Current

	// Normalise feels-like to °C for threshold logic (reuse shared helper)
	feelsC := toCelsius(cur.FeelsLike, info.TempUnit)

	// Wind in km/h for thresholds (reuse shared helper)
	windKmh := toKmh(cur.WindSpeed, info.WindUnit)

	// Today's precipitation probability (first forecast day)
	precipProb := 0
	if len(info.Forecast) > 0 {
		precipProb = info.Forecast[0].PrecipProb
	}

	uv := cur.UVIndex

	var tier string
	switch {
	case feelsC < 0:
		tier = "freezing"
	case feelsC < 8:
		tier = "cold"
	case feelsC < 15:
		tier = "cool"
	case feelsC < 22:
		tier = "mild"
	case feelsC < 29:
		tier = "warm"
	default:
		tier = "hot"
	}

	headlines := map[string]string{
		"freezing": "Bundle up — it's freezing out there",
		"cold":     "Dress warm, it's a cold one",
		"cool":     "A jacket will do nicely today",
		"mild":     "Perfect weather — dress easy",
		"warm":     "Light layers, you'll be comfortable",
		"hot":      "Stay cool — it's scorching",
	}

	advice := OutfitAdvice{
		Headline: headlines[tier],
		TempTier: tier,
	}

	switch tier {
	case "freezing":
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "thermal", Label: "Thermal Base", Color: "oi-blue",
			Note: "Moisture-wicking thermals keep heat in",
		})
	case "cold":
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "sweater", Label: "Thick Sweater", Color: "oi-blue",
			Note: "Wool or fleece sweater recommended",
		})
	case "cool":
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "longsleeve", Label: "Long Sleeve", Color: "oi-sky",
			Note: "A long-sleeve shirt is enough inside",
		})
	case "mild":
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "tshirt", Label: "T-Shirt", Color: "oi-green",
			Note: "Any casual top works great",
		})
	case "warm", "hot":
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "tshirt", Label: "Light Top", Color: "oi-orange",
			Note: "Breathable, light-coloured fabric is best",
		})
	}

	switch tier {
	case "freezing":
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "coat", Label: "Heavy Coat", Color: "oi-indigo",
			Note: "Insulated or down-filled coat essential",
		})
	case "cold":
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "coat", Label: "Winter Coat", Color: "oi-indigo",
			Note: "Lined coat with hood recommended",
		})
	case "cool":
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "jacket", Label: "Light Jacket", Color: "oi-sky",
			Note: "Zip-up or denim jacket keeps the chill off",
		})
	}

	if windKmh >= 30 && tier != "freezing" && tier != "cold" {
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "windbreaker", Label: "Windbreaker", Color: "oi-teal",
			Note: "Gusts up to " + windLabel(windKmh, info.WindUnit) + " — block the wind",
		})
	}

	if precipProb >= 60 {
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "umbrella", Label: "Umbrella", Color: "oi-blue",
			Note: "Rain likely today (" + strconv.Itoa(precipProb) + "% chance)",
		})
	} else if precipProb >= 30 {
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "raincoat", Label: "Rain Jacket", Color: "oi-sky",
			Note: "Pack one just in case (" + strconv.Itoa(precipProb) + "% chance)",
		})
	}

	if uv >= 8 {
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "sunscreen", Label: "SPF 50+", Color: "oi-orange",
			Note: "UV is very high — reapply every 2 hours",
		})
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "sunglasses", Label: "Sunglasses", Color: "oi-amber",
			Note: "Protect your eyes from UV " + strconv.Itoa(int(uv)) + " index",
		})
	} else if uv >= 5 {
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "sunscreen", Label: "Sunscreen", Color: "oi-amber",
			Note: "UV " + strconv.Itoa(int(uv)) + " — SPF 30 before heading out",
		})
	}

	if tier == "freezing" {
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "beanie", Label: "Beanie + Gloves", Color: "oi-indigo",
			Note: "Extremities lose heat fast in freezing temps",
		})
	} else if uv >= 6 {
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "hat", Label: "Sun Hat", Color: "oi-amber",
			Note: "Wide brim hat shields face and neck",
		})
	}

	if precipProb >= 50 || tier == "freezing" {
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "boots", Label: "Waterproof Boots", Color: "oi-teal",
			Note: "Keep feet dry on wet ground",
		})
	} else if tier == "hot" {
		advice.Items = append(advice.Items, OutfitItem{
			Icon: "sandals", Label: "Sandals", Color: "oi-orange",
			Note: "Let your feet breathe in the heat",
		})
	}

	// Cap to 6 items max to keep the card compact
	if len(advice.Items) > 6 {
		advice.Items = advice.Items[:6]
	}

	return advice
}

func windLabel(kmh float64, unit string) string {
	switch unit {
	case "mph":
		return strconv.Itoa(int(kmh/1.60934)) + " mph"
	case "kn":
		return strconv.Itoa(int(kmh/1.852)) + " kn"
	default:
		return strconv.Itoa(int(kmh)) + " km/h"
	}
}

package weather

// beaufortEntry maps a minimum km/h value to a Beaufort number and description.
type beaufortEntry struct {
	minKmh float64
	number int
	desc   string
}

// beaufortScale lists entries in descending order of minKmh so the first match wins.
// Based on the international Beaufort wind force scale.
var beaufortScale = []beaufortEntry{
	{117.0, 12, "Hurricane"},
	{102.0, 11, "Violent storm"},
	{88.0, 10, "Storm"},
	{74.0, 9, "Strong gale"},
	{61.0, 8, "Gale"},
	{50.0, 7, "Near gale"},
	{38.0, 6, "Strong breeze"},
	{28.0, 5, "Fresh breeze"},
	{19.0, 4, "Moderate breeze"},
	{12.0, 3, "Gentle breeze"},
	{6.0, 2, "Light breeze"},
	{1.0, 1, "Light air"},
	{0.0, 0, "Calm"},
}

// WindBeaufort returns the Beaufort scale number and a standard description for
// a wind speed given in km/h. It always returns a valid result.
func WindBeaufort(kmh float64) (number int, description string) {
	for _, b := range beaufortScale {
		if kmh >= b.minKmh {
			return b.number, b.desc
		}
	}
	return 0, "Calm"
}

// WindDescription returns the Beaufort description string for kmh.
func WindDescription(kmh float64) string {
	_, desc := WindBeaufort(kmh)
	return desc
}

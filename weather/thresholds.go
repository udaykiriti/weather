package weather

// UV index thresholds follow the WHO/WMO classification scale.
const (
	uvLow      = 3.0  // 0–2: Low
	uvModerate = 6.0  // 3–5: Moderate
	uvHigh     = 8.0  // 6–7: High
	uvVeryHigh = 11.0 // 8–10: Very High; ≥11: Extreme
)

// Wind speed thresholds in km/h (Beaufort-based).
const (
	windBreezy    = 39.0  // Beaufort 6–7: fresh/strong breeze — info alert
	windStrong    = 62.0  // Beaufort 8–9: gale force — warning
	windHurricane = 118.0 // Beaufort 12: hurricane force — danger
)

// Temperature thresholds in °C for alert logic.
const (
	tempExtremeCold = -15.0 // frostbite risk within 30 minutes
	tempFreezing    = 0.0   // freezing point; black-ice risk
	tempHeatwave    = 35.0  // heatwave warning threshold
	tempExtremeHeat = 40.0  // heat-stroke danger threshold
)

// humidityHigh is the % relative humidity threshold for a high-humidity advisory.
const humidityHigh = 85

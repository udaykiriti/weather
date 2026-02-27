package weather

// AlertLevel classifies the severity of a weather alert.
type AlertLevel string

const (
	AlertDanger  AlertLevel = "danger"
	AlertWarning AlertLevel = "warning"
	AlertInfo    AlertLevel = "info"
)

// Alert is a single weather alert to display.
type Alert struct {
	Level   AlertLevel
	Icon    string
	Title   string
	Message string
}

// toCelsius converts a temperature to Celsius regardless of the unit label.
func toCelsius(temp float64, unitSymbol string) float64 {
	if unitSymbol == "°F" {
		return (temp - 32) * 5 / 9
	}
	return temp
}

// toKmh converts wind speed to km/h regardless of the unit label.
func toKmh(speed float64, unitLabel string) float64 {
	if unitLabel == "mph" {
		return speed * 1.60934
	}
	return speed
}

// Alerts analyses a WeatherInfo and returns triggered alerts ordered by severity.
func Alerts(info *WeatherInfo) []Alert {
	var alerts []Alert

	cur := info.Current
	tempC := toCelsius(cur.Temp, info.TempUnit)
	feelsC := toCelsius(cur.FeelsLike, info.TempUnit)
	windKmh := toKmh(cur.WindSpeed, info.WindUnit)

	if isThunder(cur.Icon) {
		alerts = append(alerts, Alert{
			Level:   AlertDanger,
			Icon:    "wi-thunderstorm",
			Title:   "THUNDERSTORM ACTIVE",
			Message: "Lightning risk. Stay indoors. Unplug electronics. Avoid open areas.",
		})
	}

	if feelsC <= -15 {
		alerts = append(alerts, Alert{
			Level:   AlertDanger,
			Icon:    "wi-snowflake-cold",
			Title:   "EXTREME COLD",
			Message: "Dangerously cold. Risk of frostbite in under 30 minutes. Limit time outdoors.",
		})
	}

	if feelsC >= 40 {
		alerts = append(alerts, Alert{
			Level:   AlertDanger,
			Icon:    "wi-hot",
			Title:   "EXTREME HEAT",
			Message: "Heat index critical. Risk of heat stroke. Stay in the shade and hydrate constantly.",
		})
	}

	if windKmh >= 118 {
		alerts = append(alerts, Alert{
			Level:   AlertDanger,
			Icon:    "wi-strong-wind",
			Title:   "HURRICANE-FORCE WIND",
			Message: "Extremely dangerous winds. Take shelter immediately. Do not drive.",
		})
	}

	if isHeavyRain(cur.Icon) {
		alerts = append(alerts, Alert{
			Level:   AlertWarning,
			Icon:    "wi-rain-wind",
			Title:   "HEAVY RAIN",
			Message: "Reduced visibility and possible flash flooding. Drive carefully.",
		})
	}

	if isHeavySnow(cur.Icon) {
		alerts = append(alerts, Alert{
			Level:   AlertWarning,
			Icon:    "wi-snow-wind",
			Title:   "HEAVY SNOW",
			Message: "Roads may be impassable. Allow extra travel time and check road conditions.",
		})
	}

	// Freeze: -15 °C < tempC < 0 °C (extreme cold already covers ≤ -15)
	if tempC < 0 && tempC > -15 {
		alerts = append(alerts, Alert{
			Level:   AlertWarning,
			Icon:    "wi-thermometer-exterior",
			Title:   "FREEZING CONDITIONS",
			Message: "Black ice possible on roads. Wrap up warm and watch your step.",
		})
	}

	// Heatwave: 35 °C ≤ feelsC < 40 °C (extreme heat covers ≥ 40)
	if feelsC >= 35 && feelsC < 40 {
		alerts = append(alerts, Alert{
			Level:   AlertWarning,
			Icon:    "wi-day-sunny",
			Title:   "HEATWAVE WARNING",
			Message: "Dangerously warm. Drink water, avoid peak sun hours (11am–3pm), check on vulnerable people.",
		})
	}

	if windKmh >= 62 && windKmh < 118 {
		alerts = append(alerts, Alert{
			Level:   AlertWarning,
			Icon:    "wi-strong-wind",
			Title:   "STRONG WIND WARNING",
			Message: "Gale-force winds. Secure loose outdoor objects. Drive with care.",
		})
	}

	if isFog(cur.Icon) {
		alerts = append(alerts, Alert{
			Level:   AlertInfo,
			Icon:    "wi-fog",
			Title:   "FOG ADVISORY",
			Message: "Low visibility on roads. Use fog lights and reduce speed.",
		})
	}

	if cur.Humidity >= 85 {
		alerts = append(alerts, Alert{
			Level:   AlertInfo,
			Icon:    "wi-humidity",
			Title:   "HIGH HUMIDITY",
			Message: "Air feels heavy and muggy. Stay hydrated and take it easy outdoors.",
		})
	}

	if windKmh >= 39 && windKmh < 62 {
		alerts = append(alerts, Alert{
			Level:   AlertInfo,
			Icon:    "wi-windy",
			Title:   "WINDY CONDITIONS",
			Message: "Fresh to strong breeze. Hold onto your hat — literally.",
		})
	}

	return alerts
}

func isThunder(icon string) bool   { return icon == "wi-thunderstorm" || icon == "wi-storm-showers" }
func isHeavyRain(icon string) bool { return icon == "wi-rain-wind" }
func isHeavySnow(icon string) bool { return icon == "wi-snow-wind" }
func isFog(icon string) bool       { return icon == "wi-fog" }

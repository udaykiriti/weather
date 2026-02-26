package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	geoURL      = "https://geocoding-api.open-meteo.com/v1/search"
	forecastURL = "https://api.open-meteo.com/v1/forecast"
	reverseURL  = "https://nominatim.openstreetmap.org/reverse"
)

// Client calls Open-Meteo APIs (no API key required).
type Client struct {
	HTTP *http.Client
}

// NewClient returns a Client with a 30-second timeout and a resilient DNS
// resolver that falls back to public DNS servers when the system resolver fails.
func NewClient() *Client {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 5 * time.Second}
			// Try public DNS servers when the system resolver is unavailable.
			for _, dns := range []string{"8.8.8.8:53", "1.1.1.1:53", "8.8.4.4:53"} {
				if conn, err := d.DialContext(ctx, "udp", dns); err == nil {
					return conn, nil
				}
			}
			return d.DialContext(ctx, network, address) // last-resort: system
		},
	}
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:  30 * time.Second,
			Resolver: resolver,
		}).DialContext,
	}
	return &Client{HTTP: &http.Client{Timeout: 30 * time.Second, Transport: transport}}
}

// --- internal API types ---

type geoResponse struct {
	Results []GeoLocation `json:"results"`
}

type GeoLocation struct {
	Name        string  `json:"name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Timezone    string  `json:"timezone"`
}

type currentRaw struct {
	Time        string  `json:"time"`
	Temperature float64 `json:"temperature_2m"`
	FeelsLike   float64 `json:"apparent_temperature"`
	Humidity    int     `json:"relative_humidity_2m"`
	WeatherCode int     `json:"weather_code"`
	CloudCover  int     `json:"cloud_cover"`
	WindSpeed   float64 `json:"wind_speed_10m"`
	WindDir     int     `json:"wind_direction_10m"`
	Pressure    float64 `json:"pressure_msl"`
	DewPoint    float64 `json:"dew_point_2m"`
	UVIndex     float64 `json:"uv_index"`
}

type dailyRaw struct {
	Time          []string  `json:"time"`
	WeatherCode   []int     `json:"weather_code"`
	TempMax       []float64 `json:"temperature_2m_max"`
	TempMin       []float64 `json:"temperature_2m_min"`
	WindMax       []float64 `json:"wind_speed_10m_max"`
	PrecipProbMax []int     `json:"precipitation_probability_max"`
	Sunrise       []string  `json:"sunrise"`
	Sunset        []string  `json:"sunset"`
}

type hourlyRaw struct {
	Time        []string  `json:"time"`
	Temperature []float64 `json:"temperature_2m"`
	PrecipProb  []int     `json:"precipitation_probability"`
	WeatherCode []int     `json:"weather_code"`
	WindSpeed   []float64 `json:"wind_speed_10m"`
}

type forecastRaw struct {
	Current currentRaw `json:"current"`
	Daily   dailyRaw   `json:"daily"`
	Hourly  hourlyRaw  `json:"hourly"`
}

// --- public display types ---

type CurrentDisplay struct {
	Time        string
	Temp        float64
	FeelsLike   float64
	Humidity    int
	Description string
	Icon        string
	CloudCover  int
	WindSpeed   float64
	WindDir     int
	Pressure    float64
	DewPoint    float64
	UVIndex     float64
}

type ForecastDay struct {
	Date        string
	Description string
	Icon        string
	TempMax     float64
	TempMin     float64
	WindMax     float64
	PrecipProb  int // 0-100 percent probability of precipitation
}

// HourlyPoint holds weather data for one hour.
type HourlyPoint struct {
	Time        string // "HH:MM"
	Temp        float64
	PrecipProb  int
	Description string
	Icon        string
	WindSpeed   float64
}

// SunBar holds values needed to render the sunrise/sunset arc.
type SunBar struct {
	SunriseTime    string
	SunsetTime     string
	CurrentTime    string
	SunPositionPct float64 // 0-100, float for smooth SVG positioning
	IsDay          bool
	DaylightHours  string
	MoonPhase      float64 // 0.0 = new moon, 0.5 = full moon, 1.0 = new moon
	MoonPhaseName  string  // e.g. "Waxing Crescent"
}

type WeatherInfo struct {
	CityName    string
	Country     string
	CountryCode string
	TempUnit    string
	WindUnit    string
	Current     CurrentDisplay
	Forecast    []ForecastDay
	Hourly      []HourlyPoint // next 24 hours
	Sun         SunBar
	Consensus   *ConsensusInfo
	Outfit      OutfitAdvice
}

// getJSON makes a GET request with one automatic retry on timeout/connection error.
func (c *Client) getJSON(rawURL string, v any) error {
	const maxAttempts = 2
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(1 * time.Second) // brief pause before retry
		}
		resp, err := c.HTTP.Get(rawURL)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue // retry on network error
		}
		// Close body explicitly (not deferred) so each retry releases its connection.
		err = func() error {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
				return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
			}
			if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
				return fmt.Errorf("decode failed: %w", err)
			}
			return nil
		}()
		return err // success or non-retryable error
	}
	return lastErr
}

// Geocode resolves a city name to coordinates.
func (c *Client) Geocode(city string) (*GeoLocation, error) {
	u := fmt.Sprintf("%s?name=%s&count=1&language=en&format=json", geoURL, url.QueryEscape(city))
	var geo geoResponse
	if err := c.getJSON(u, &geo); err != nil {
		return nil, fmt.Errorf("geocode: %w", err)
	}
	if len(geo.Results) == 0 {
		return nil, fmt.Errorf("city %q not found", city)
	}
	return &geo.Results[0], nil
}

// ReverseGeocode converts coordinates to a city name via Nominatim.
// Returns the best available city-level name (city → town → village → county).
func (c *Client) ReverseGeocode(lat, lon float64) (string, error) {
	u := fmt.Sprintf("%s?lat=%.6f&lon=%.6f&format=json&zoom=10&addressdetails=1",
		reverseURL, lat, lon)

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	// Nominatim policy requires a descriptive User-Agent
	req.Header.Set("User-Agent", "GoWeatherApp/1.0")
	req.Header.Set("Accept-Language", "en")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("reverse geocode request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("reverse geocode: HTTP %d", resp.StatusCode)
	}

	var result struct {
		DisplayName string `json:"display_name"`
		Address     struct {
			City         string `json:"city"`
			Town         string `json:"town"`
			Village      string `json:"village"`
			Municipality string `json:"municipality"`
			County       string `json:"county"`
		} `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("reverse geocode decode: %w", err)
	}

	addr := result.Address
	for _, candidate := range []string{
		addr.City, addr.Town, addr.Village, addr.Municipality, addr.County,
	} {
		if candidate != "" {
			return candidate, nil
		}
	}
	if result.DisplayName != "" {
		// Fall back to first part of display name (before first comma)
		if idx := len(result.DisplayName); idx > 0 {
			for i, c := range result.DisplayName {
				if c == ',' {
					return result.DisplayName[:i], nil
				}
			}
		}
		return result.DisplayName, nil
	}
	return "", fmt.Errorf("no city found for coordinates %.4f,%.4f", lat, lon)
}

// GetWeather fetches current weather + 5-day forecast for a city.
func (c *Client) GetWeather(city, units string) (*WeatherInfo, error) {
	loc, err := c.Geocode(city)
	if err != nil {
		return nil, err
	}

	tempUnit, windUnit := "celsius", "kmh"
	if units == "imperial" {
		tempUnit, windUnit = "fahrenheit", "mph"
	}

	u := fmt.Sprintf(
		"%s?latitude=%.4f&longitude=%.4f"+
			"&current=temperature_2m,apparent_temperature,relative_humidity_2m,weather_code,cloud_cover,wind_speed_10m,wind_direction_10m,pressure_msl,dew_point_2m,uv_index"+
			"&hourly=temperature_2m,precipitation_probability,weather_code,wind_speed_10m"+
			"&daily=weather_code,temperature_2m_max,temperature_2m_min,wind_speed_10m_max,precipitation_probability_max,sunrise,sunset"+
			"&temperature_unit=%s&wind_speed_unit=%s&timezone=%s&forecast_days=5",
		forecastURL, loc.Latitude, loc.Longitude,
		tempUnit, windUnit, url.QueryEscape(loc.Timezone),
	)

	var raw forecastRaw
	if err := c.getJSON(u, &raw); err != nil {
		return nil, fmt.Errorf("forecast: %w", err)
	}

	info := &WeatherInfo{
		CityName:    loc.Name,
		Country:     loc.Country,
		CountryCode: loc.CountryCode,
		TempUnit:    TempUnitSymbol(units),
		WindUnit:    WindUnitLabel(units),
		Current: CurrentDisplay{
			Time:        raw.Current.Time,
			Temp:        raw.Current.Temperature,
			FeelsLike:   raw.Current.FeelsLike,
			Humidity:    raw.Current.Humidity,
			Description: WMODescription(raw.Current.WeatherCode),
			Icon:        WMOIconClass(raw.Current.WeatherCode),
			CloudCover:  raw.Current.CloudCover,
			WindSpeed:   raw.Current.WindSpeed,
			WindDir:     raw.Current.WindDir,
			Pressure:    raw.Current.Pressure,
			DewPoint:    raw.Current.DewPoint,
			UVIndex:     raw.Current.UVIndex,
		},
	}

	for i, date := range raw.Daily.Time {
		if i >= len(raw.Daily.WeatherCode) || i >= len(raw.Daily.TempMax) {
			break
		}
		precipProb := 0
		if i < len(raw.Daily.PrecipProbMax) {
			precipProb = raw.Daily.PrecipProbMax[i]
		}
		info.Forecast = append(info.Forecast, ForecastDay{
			Date:        date,
			Description: WMODescription(raw.Daily.WeatherCode[i]),
			Icon:        WMOIconClass(raw.Daily.WeatherCode[i]),
			TempMax:     raw.Daily.TempMax[i],
			TempMin:     raw.Daily.TempMin[i],
			WindMax:     raw.Daily.WindMax[i],
			PrecipProb:  precipProb,
		})
	}

	if len(raw.Daily.Sunrise) > 0 && len(raw.Daily.Sunset) > 0 {
		info.Sun = buildSunBar(raw.Current.Time, raw.Daily.Sunrise[0], raw.Daily.Sunset[0], loc.Timezone)
	}

	// Parse next 24 hourly points starting from the current hour.
	info.Hourly = parseHourly(raw.Hourly, raw.Current.Time, loc.Timezone)

	// Build outfit advice from current conditions.
	info.Outfit = BuildOutfit(info)

	// Fetch multi-model consensus in parallel (non-fatal if it fails)
	info.Consensus = c.FetchConsensus(loc.Latitude, loc.Longitude, loc.Timezone, tempUnit, windUnit)

	return info, nil
}

// buildSunBar computes all values for the sunrise/sunset progress bar.
func buildSunBar(currentTimeStr, sunriseStr, sunsetStr, timezone string) SunBar {
	const layout = "2006-01-02T15:04"

	tz, err := time.LoadLocation(timezone)
	if err != nil {
		tz = time.UTC
	}

	parse := func(s string) time.Time {
		t, _ := time.ParseInLocation(layout, s, tz)
		return t
	}

	sunrise := parse(sunriseStr)
	sunset := parse(sunsetStr)
	now := parse(currentTimeStr)

	daylightMins := sunset.Sub(sunrise).Minutes()
	h := int(daylightMins) / 60
	m := int(daylightMins) % 60

	isDay := now.After(sunrise) && now.Before(sunset)

	var positionPct float64
	if daylightMins > 0 {
		elapsed := now.Sub(sunrise).Minutes()
		positionPct = elapsed / daylightMins * 100
		positionPct = math.Max(0, math.Min(100, positionPct))
	}

	mp := moonPhase(now.UTC())
	return SunBar{
		SunriseTime:    sunrise.Format("15:04"),
		SunsetTime:     sunset.Format("15:04"),
		CurrentTime:    now.Format("15:04"),
		SunPositionPct: positionPct,
		IsDay:          isDay,
		DaylightHours:  fmt.Sprintf("%dh %dm", h, m),
		MoonPhase:      mp,
		MoonPhaseName:  moonPhaseName(mp),
	}
}

// parseHourly extracts the next 24 hourly points starting from currentTimeStr.
func parseHourly(h hourlyRaw, currentTimeStr, timezone string) []HourlyPoint {
	const layout = "2006-01-02T15:04"
	tz, err := time.LoadLocation(timezone)
	if err != nil {
		tz = time.UTC
	}
	now, err := time.ParseInLocation(layout, currentTimeStr, tz)
	if err != nil {
		return nil
	}

	var points []HourlyPoint
	for i, ts := range h.Time {
		t, err := time.ParseInLocation(layout, ts, tz)
		if err != nil {
			continue
		}
		if t.Before(now) {
			continue
		}
		if len(points) >= 24 {
			break
		}
		wc := safeInt(h.WeatherCode, i)
		pp := safeInt(h.PrecipProb, i)
		ws := safeFloat(h.WindSpeed, i)
		temp := safeFloat(h.Temperature, i)
		points = append(points, HourlyPoint{
			Time:        t.Format("15:04"),
			Temp:        temp,
			PrecipProb:  pp,
			Description: WMODescription(wc),
			Icon:        WMOIconClass(wc),
			WindSpeed:   ws,
		})
	}
	return points
}

func safeInt(s []int, i int) int {
	if i < len(s) {
		return s[i]
	}
	return 0
}

func safeFloat(s []float64, i int) float64 {
	if i < len(s) {
		return s[i]
	}
	return 0
}

// TempUnitSymbol returns the temperature symbol.
func TempUnitSymbol(units string) string {
	if units == "imperial" {
		return "°F"
	}
	return "°C"
}

// WindUnitLabel returns the wind speed label.
func WindUnitLabel(units string) string {
	if units == "imperial" {
		return "mph"
	}
	return "km/h"
}

// WMODescription converts a WMO weather code to a description.
func WMODescription(code int) string {
	switch {
	case code == 0:
		return "Clear sky"
	case code == 1:
		return "Mainly clear"
	case code == 2:
		return "Partly cloudy"
	case code == 3:
		return "Overcast"
	case code == 45 || code == 48:
		return "Fog"
	case code >= 51 && code <= 53:
		return "Light drizzle"
	case code == 55:
		return "Dense drizzle"
	case code == 61:
		return "Slight rain"
	case code == 63:
		return "Moderate rain"
	case code == 65:
		return "Heavy rain"
	case code == 71:
		return "Slight snow"
	case code == 73:
		return "Moderate snow"
	case code == 75:
		return "Heavy snow"
	case code == 77:
		return "Snow grains"
	case code >= 80 && code <= 82:
		return "Rain showers"
	case code >= 85 && code <= 86:
		return "Snow showers"
	case code == 95:
		return "Thunderstorm"
	case code == 96 || code == 99:
		return "Thunderstorm with hail"
	default:
		return "Unknown"
	}
}

// WMOIconClass maps a WMO weather code to an erikflowers/weather-icons CSS class.
func WMOIconClass(code int) string {
	switch {
	case code == 0:
		return "wi-day-sunny"
	case code == 1:
		return "wi-day-sunny-overcast"
	case code == 2:
		return "wi-day-cloudy"
	case code == 3:
		return "wi-cloudy"
	case code == 45 || code == 48:
		return "wi-fog"
	case code >= 51 && code <= 53:
		return "wi-sprinkle"
	case code == 55:
		return "wi-rain-mix"
	case code == 61:
		return "wi-rain-mix"
	case code == 63:
		return "wi-rain"
	case code == 65:
		return "wi-rain-wind"
	case code == 71:
		return "wi-snow"
	case code == 73:
		return "wi-snow"
	case code == 75:
		return "wi-snow-wind"
	case code == 77:
		return "wi-snowflake-cold"
	case code >= 80 && code <= 82:
		return "wi-showers"
	case code >= 85 && code <= 86:
		return "wi-snow"
	case code == 95:
		return "wi-thunderstorm"
	case code == 96 || code == 99:
		return "wi-storm-showers"
	default:
		return "wi-na"
	}
}

// WindCompass converts a wind direction in degrees to an 8-point compass label.
func WindCompass(deg int) string {
	var dirs = [8]string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	idx := int((float64(deg)+22.5)/45.0) % 8
	return dirs[idx]
}

// ── Moon phase ────────────────────────────────────────────────────────────────

// synodicMonth is the average length of a lunar cycle in days.
const synodicMonth = 29.530589

// referenceNewMoon is a known new moon date used as the calculation epoch.
var referenceNewMoon = time.Date(2000, 1, 6, 18, 14, 0, 0, time.UTC)

// moonPhase returns a value in [0, 1) representing the current lunar phase:
// 0 = new moon, 0.25 = first quarter, 0.5 = full moon, 0.75 = last quarter.
func moonPhase(t time.Time) float64 {
	days := t.Sub(referenceNewMoon).Hours() / 24
	phase := math.Mod(days, synodicMonth) / synodicMonth
	if phase < 0 {
		phase += 1
	}
	return phase
}

// moonPhaseName maps a phase fraction to a human-readable name.
func moonPhaseName(p float64) string {
	switch {
	case p < 0.034 || p >= 0.966:
		return "New Moon"
	case p < 0.216:
		return "Waxing Crescent"
	case p < 0.284:
		return "First Quarter"
	case p < 0.466:
		return "Waxing Gibbous"
	case p < 0.534:
		return "Full Moon"
	case p < 0.716:
		return "Waning Gibbous"
	case p < 0.784:
		return "Last Quarter"
	default:
		return "Waning Crescent"
	}
}

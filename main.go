package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"WeatherApp/weather"
)

// outfitIcons is built once at startup; each value is a safe inline SVG string.
var outfitIcons = map[string]template.HTML{
	"thermal":     `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M6 4h12v4l2 2v8a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2v-8l2-2V4z"/><line x1="9" y1="4" x2="9" y2="8"/><line x1="15" y1="4" x2="15" y2="8"/><path d="M6 12h12"/></svg>`,
	"sweater":     `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 7l3-3h4l2 3 2-3h4l3 3-3 3v10H6V10L3 7z"/><path d="M9 4c0 1.7 1.3 3 3 3s3-1.3 3-3"/></svg>`,
	"longsleeve":  `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l3-3h4l2 2 2-2h4l3 3-3 2v8H6v-8L3 9z"/><path d="M9 6c0 1.1.9 2 3 2s3-.9 3-2"/></svg>`,
	"tshirt":      `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l3-3h4l2 2 2-2h4l3 3-3 2v9H6v-9L3 9z"/></svg>`,
	"coat":        `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 7l3-3 3 4V20H3V7z"/><path d="M21 7l-3-3-3 4v13h6V7z"/><path d="M9 8l3-4 3 4"/><path d="M9 20v-7h6v7"/></svg>`,
	"jacket":      `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 8l3-3 3 3v12H3V8z"/><path d="M21 8l-3-3-3 3v12h6V8z"/><path d="M9 8l3-2 3 2"/><path d="M9 14h6"/><circle cx="10.5" cy="11" r=".5" fill="currentColor"/></svg>`,
	"windbreaker": `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 8l3-3 3 3v12H3V8z"/><path d="M21 8l-3-3-3 3v12h6V8z"/><path d="M9 8l3-2 3 2"/><path d="M2 13h5"/><path d="M17 13h5"/></svg>`,
	"umbrella":    `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M23 12a11 11 0 0 0-22 0z"/><path d="M12 12v7a2 2 0 0 0 4 0"/></svg>`,
	"raincoat":    `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 8l3-3 3 3v12H3V8z"/><path d="M21 8l-3-3-3 3v12h6V8z"/><path d="M9 8l3-2 3 2"/><line x1="7" y1="17" x2="7" y2="19"/><line x1="12" y1="16" x2="12" y2="18"/><line x1="17" y1="17" x2="17" y2="19"/></svg>`,
	"sunscreen":   `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="8" y="3" width="8" height="18" rx="3"/><path d="M12 3v2"/><path d="M10 7h4"/><path d="M8 13h8"/><circle cx="12" cy="17" r="1" fill="currentColor"/></svg>`,
	"sunglasses":  `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="7" cy="13" r="4"/><circle cx="17" cy="13" r="4"/><path d="M11 13h2"/><path d="M1 10l2 3"/><path d="M23 10l-2 3"/></svg>`,
	"beanie":      `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M5 15c0-3.9 3.1-7 7-7s7 3.1 7 7"/><path d="M3 15h18"/><path d="M5 15v4h14v-4"/><circle cx="12" cy="7" r="1.5" fill="currentColor"/><path d="M12 7V5"/></svg>`,
	"hat":         `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><ellipse cx="12" cy="17" rx="9" ry="2.5"/><path d="M8 17V9a4 4 0 0 1 8 0v8"/></svg>`,
	"boots":       `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M6 4h6v10l4 2v4H4v-4l2-2V4z"/><path d="M8 4h4"/><path d="M4 20h12"/></svg>`,
	"sandals":     `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="16" width="18" height="4" rx="2"/><path d="M7 16v-3"/><path d="M12 16V10"/><path d="M17 16v-3"/><path d="M7 13h10"/></svg>`,
}

// outfitIconFallback is returned when an icon key is not found.
const outfitIconFallback template.HTML = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l3-3h4l2 2 2-2h4l3 3-3 2v9H6v-9L3 9z"/></svg>`

// moonPhaseSVG generates a 26×26 inline SVG representing the lunar phase.
// It uses two SVG arcs: an outer semicircle (lit side) and an inner ellipse arc
// (crescent or gibbous extension), drawn on top of a dark disc background.
func moonPhaseSVG(phase float64) template.HTML {
	const (
		cx        = 12.0
		cy        = 12.0
		radius    = 9.0
		arcTop    = cy - radius // 3.0 — top of the moon disc
		arcBottom = cy + radius // 21.0 — bottom of the moon disc
		litColor  = "#fef3c7"  // amber-100: lit side fill
		darkColor = "#1e293b"  // slate-800: dark side fill
		strokeClr = "#6366f1"  // indigo-500: outline stroke
	)

	// New moon: fully dark disc.
	if phase < 0.02 || phase > 0.98 {
		return template.HTML(fmt.Sprintf(
			`<svg viewBox="0 0 24 24" width="26" height="26"><circle cx="12" cy="12" r="9" fill="%s" stroke="%s" stroke-width="1.2"/></svg>`,
			darkColor, strokeClr))
	}

	// Full moon: fully lit disc.
	if phase > 0.48 && phase < 0.52 {
		return template.HTML(fmt.Sprintf(
			`<svg viewBox="0 0 24 24" width="26" height="26"><circle cx="12" cy="12" r="9" fill="%s" stroke="%s" stroke-width="1.2"/></svg>`,
			litColor, strokeClr))
	}

	// isWaxing: moon is growing (phase 0 → 0.5).
	// isWaning: moon is shrinking (phase 0.5 → 1).
	isWaxing := phase < 0.5

	// illumination goes from 0 (new) to 1 (full) and back to 0.
	// For waxing: phase 0→0.5 maps to illumination 0→1.
	// For waning: phase 0.5→1 maps to illumination 1→0.
	var illumination float64
	if isWaxing {
		illumination = phase * 2
	} else {
		illumination = (1 - phase) * 2
	}

	// innerRadiusX: the x-radius of the inner ellipse arc.
	// It is 0 at quarter moon (illumination=0.5) and max at new/full.
	innerRadiusX := math.Max(0.5, radius*math.Abs(math.Cos(math.Pi*illumination)))

	// SVG arc sweep flags: 1 = clockwise, 0 = counter-clockwise.
	// outerSweep draws the lit semicircle on the correct side.
	// innerSweep draws the crescent/gibbous inner boundary.
	var outerSweep, innerSweep int

	if isWaxing {
		// Lit side is on the right: outer arc goes clockwise.
		outerSweep = 1
		innerSweep = 0
	} else {
		// Lit side is on the left: outer arc goes counter-clockwise.
		outerSweep = 0
		innerSweep = 1
	}

	// Gibbous phase (more than half lit): flip the inner arc to the shadow side.
	if illumination > 0.5 {
		if isWaxing {
			innerSweep = 1
		} else {
			innerSweep = 0
		}
	}

	// Build the lit-area SVG path using two arc commands:
	// 1. Outer arc: semicircle along the lit side (top → bottom).
	// 2. Inner arc: ellipse arc back along the terminator (bottom → top).
	litPath := fmt.Sprintf(
		"M %.2f %.2f A %.2f %.2f 0 0 %d %.2f %.2f A %.2f %.2f 0 0 %d %.2f %.2f Z",
		cx, arcTop,
		radius, radius, outerSweep, cx, arcBottom,
		innerRadiusX, radius, innerSweep, cx, arcTop,
	)

	svg := fmt.Sprintf(
		`<svg viewBox="0 0 24 24" width="26" height="26"><circle cx="12" cy="12" r="9" fill="%s" stroke="%s" stroke-width="1.2"/><path d="%s" fill="%s"/></svg>`,
		darkColor, strokeClr, litPath, litColor,
	)
	return template.HTML(svg)
}

// sunArcY maps a 0–100 sun-position percentage to an SVG y-coordinate.
// The sun travels along a half-ellipse; viewBox is 0 0 400 80, horizon at y=72.
func sunArcY(pct float64) float64 {
	t := pct / 100.0
	y := 72.0 - 72.0*math.Sin(math.Pi*t)
	return math.Max(-5, y)
}

// sunArcX maps a 0–100 sun-position percentage to an SVG x-coordinate.
// The usable x range is 10–390 across the 400-wide viewBox.
func sunArcX(pct float64) float64 {
	return 10 + (pct/100)*380
}

// precipBarColor returns the fill colour for a precipitation-probability bar.
func precipBarColor(pct int) string {
	switch {
	case pct >= 80:
		return "#1d4ed8" // dark blue — very likely rain
	case pct >= 60:
		return "#3b82f6" // blue — likely rain
	case pct >= 30:
		return "#93c5fd" // light blue — possible rain
	default:
		return "#dbeafe" // pale blue — unlikely rain
	}
}

// agreementColor returns a CSS colour reflecting model-agreement percentage.
func agreementColor(pct int) string {
	switch {
	case pct >= 80:
		return "#22c55e" // green — high agreement
	case pct >= 50:
		return "#f59e0b" // amber — moderate agreement
	default:
		return "#ef4444" // red — low agreement
	}
}

// consensusBarWidth converts a model temperature into a 5–95 % bar width
// relative to the min/max temperature range across all models.
func consensusBarWidth(temp, minTemp, maxTemp float64) float64 {
	spread := maxTemp - minTemp
	if spread <= 0 {
		return 50.0
	}
	// Map temp linearly to 10–90% range, then clamp to 5–95%.
	position := (temp-minTemp)/spread*80.0 + 10.0
	if position < 5 {
		position = 5
	}
	if position > 95 {
		position = 95
	}
	return position
}

// dewComfortLabel returns a comfort label for a dew-point value,
// normalising Fahrenheit to Celsius before comparison.
func dewComfortLabel(dp float64, unit string) string {
	celsius := dp
	if unit == "°F" {
		celsius = (dp - 32) * 5 / 9
	}
	switch {
	case celsius > 21:
		return "Oppressive"
	case celsius > 18:
		return "Humid"
	case celsius > 13:
		return "Comfortable"
	case celsius > 7:
		return "Dry"
	default:
		return "Very Dry"
	}
}

// hourlyPrecipSVG builds an inline SVG bar chart of hourly rain probability.
// Each of the 24 hourly points gets one bar; bars are coloured by probability.
func hourlyPrecipSVG(hourly []weather.HourlyPoint) template.HTML {
	if len(hourly) == 0 {
		return ""
	}

	const (
		slotWidth = 30  // px allocated per hour slot
		barWidth  = 27  // drawn bar width (3 px gap between bars)
		baseline  = 70  // y-coordinate of the x-axis baseline
		maxBarH   = 60  // bar height for 100% probability
		labelY    = 88  // y-coordinate for time labels
		viewH     = 92  // total SVG height
		viewW     = 720 // 24 × slotWidth
	)

	// Cap to 24 data points.
	n := len(hourly)
	if n > 24 {
		n = 24
	}

	var sb strings.Builder
	sb.Grow(4096) // pre-allocate: 24 bars × ~120 bytes each + header/footer

	// SVG opening tag.
	fmt.Fprintf(&sb,
		`<svg viewBox="0 0 %d %d" width="100%%" preserveAspectRatio="none" class="precip-svg" aria-label="Hourly precipitation probability">`,
		viewW, viewH)

	// 50% reference gridline.
	gridY := baseline - maxBarH/2
	fmt.Fprintf(&sb, `<line x1="0" y1="%d" x2="%d" y2="%d" class="pchart-grid"/>`, gridY, viewW, gridY)
	fmt.Fprintf(&sb, `<text x="2" y="%d" class="pchart-lbl" text-anchor="start">50%%</text>`, gridY-2)

	// Draw one bar per hour.
	for i := 0; i < n; i++ {
		point := hourly[i]
		prob := point.PrecipProb

		// Bar geometry.
		barHeight := maxBarH * prob / 100
		barX := i * slotWidth
		barY := baseline - barHeight

		// Ensure zero-probability bars are still visible as a 1 px line.
		if barHeight == 0 {
			barHeight = 1
			barY = baseline - 1
		}

		fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s" rx="2"/>`,
			barX, barY, barWidth, barHeight, precipBarColor(prob))

		// Time label every 4 hours.
		if i%4 == 0 {
			labelX := barX + slotWidth/2
			fmt.Fprintf(&sb, `<text x="%d" y="%d" class="pchart-lbl" text-anchor="middle">%s</text>`,
				labelX, labelY, point.Time)
		}
	}

	// Baseline rule.
	fmt.Fprintf(&sb, `<line x1="0" y1="%d" x2="%d" y2="%d" class="pchart-base"/>`, baseline, viewW, baseline)
	sb.WriteString(`</svg>`)

	return template.HTML(sb.String())
}

var tmpl = template.Must(
	template.New("").Funcs(template.FuncMap{
		// Simple arithmetic helpers for SVG positioning.
		"subf": func(a, b float64) float64 { return a - b },
		"mulf": func(a, b float64) float64 { return a * b },
		// Sun arc position on the sunrise/sunset SVG.
		"arcY": sunArcY,
		"arcX": sunArcX,
		"not":           func(b bool) bool { return !b },
		"uvLevel":       weather.UVLevel,
		"uvColorClass":  weather.UVColorClass,
		"windCompass":   weather.WindCompass,
		"windDesc":      weather.WindDescription,
		"humidityLabel": weather.HumidityLabel,
		"moonPhaseSVG":  moonPhaseSVG,
		"dewComfort":    dewComfortLabel,
		"outfitSVG": func(icon string) template.HTML {
			if svg, ok := outfitIcons[icon]; ok {
				return svg
			}
			return outfitIconFallback
		},
		"hourlyPrecipSVG": hourlyPrecipSVG,
		"agreeColor":      agreementColor,
		"consBarW":        consensusBarWidth,
	}).ParseFiles("templates/index.html"),
)

const (
	cacheTTL     = 10 * time.Minute
	cacheMaxSize = 200             // max entries before oldest-first eviction
	cacheCleanup = 5 * time.Minute // how often to sweep expired entries
)

type cacheEntry struct {
	info      *weather.WeatherInfo
	expires   time.Time
	createdAt time.Time
}

var (
	cacheMu sync.RWMutex
	cache   = make(map[string]*cacheEntry)
)

func cacheKey(city, units string) string {
	return strings.ToLower(city) + "|" + units
}

// oldestCacheKey returns the key of the entry with the earliest createdAt time.
// Must be called with cacheMu held.
func oldestCacheKey() string {
	var oldestKey string
	var oldestTime time.Time
	for key, entry := range cache {
		if oldestKey == "" || entry.createdAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.createdAt
		}
	}
	return oldestKey
}

func cacheGet(key string) *weather.WeatherInfo {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	if e, ok := cache[key]; ok && time.Now().Before(e.expires) {
		return e.info
	}
	return nil
}

func cacheSet(key string, info *weather.WeatherInfo) {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	// Evict oldest entry if at capacity (before inserting new one).
	if _, exists := cache[key]; !exists && len(cache) >= cacheMaxSize {
		delete(cache, oldestCacheKey())
	}

	now := time.Now()
	cache[key] = &cacheEntry{
		info:      info,
		expires:   now.Add(cacheTTL),
		createdAt: now,
	}
}

// startCacheCleanup launches a background goroutine that periodically removes
// expired entries so the map does not grow without bound.
func startCacheCleanup() {
	go func() {
		ticker := time.NewTicker(cacheCleanup)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			cacheMu.Lock()
			for k, e := range cache {
				if now.After(e.expires) {
					delete(cache, k)
				}
			}
			cacheMu.Unlock()
		}
	}()
}

type PageData struct {
	City   string
	Units  string
	Info   *weather.WeatherInfo
	Alerts []weather.Alert
	Quote  string
	Advice string
	Error  string
}


func main() {
	client := weather.NewClient()
	startCacheCleanup()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// /api/reverse?lat=...&lon=... — server-side reverse geocode proxy (avoids browser CORS)
	http.HandleFunc("/api/reverse", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		latStr := r.FormValue("lat")
		lonStr := r.FormValue("lon")
		var lat, lon float64
		if _, err := fmt.Sscanf(latStr, "%f", &lat); err != nil || latStr == "" {
			http.Error(w, `{"error":"missing lat"}`, http.StatusBadRequest)
			return
		}
		if _, err := fmt.Sscanf(lonStr, "%f", &lon); err != nil || lonStr == "" {
			http.Error(w, `{"error":"missing lon"}`, http.StatusBadRequest)
			return
		}
		if lat < -90 || lat > 90 {
			http.Error(w, `{"error":"lat out of range [-90, 90]"}`, http.StatusBadRequest)
			return
		}
		if lon < -180 || lon > 180 {
			http.Error(w, `{"error":"lon out of range [-180, 180]"}`, http.StatusBadRequest)
			return
		}
		city, err := client.ReverseGeocode(r.Context(), lat, lon)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		fmt.Fprintf(w, `{"city":%q}`, city)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET and HEAD
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		city := strings.TrimSpace(r.FormValue("city"))
		units := r.FormValue("units")
		if units != "imperial" {
			units = "metric"
		}

		data := PageData{City: city, Units: units}

		if city != "" {
			// Input validation
			if len(city) > 100 {
				w.WriteHeader(http.StatusBadRequest)
				data.Error = "City name is too long (max 100 characters)."
				_ = tmpl.ExecuteTemplate(w, "index.html", data)
				return
			}

			// Check cache first
			key := cacheKey(city, units)
			info := cacheGet(key)

			if info == nil {
				var err error
				info, err = client.GetWeather(r.Context(), city, units)
				if err != nil {
					// Distinguish not-found from network/server errors
					status := http.StatusBadGateway
					errMsg := err.Error()
					switch {
					case strings.Contains(errMsg, "not found"):
						status = http.StatusNotFound
					case strings.Contains(errMsg, "too long"):
						status = http.StatusBadRequest
					case strings.Contains(errMsg, "deadline exceeded") ||
						strings.Contains(errMsg, "timeout") ||
						strings.Contains(errMsg, "request failed"):
						errMsg = "Could not reach the weather service — please check your connection and try again."
					}
					w.WriteHeader(status)
					data.Error = errMsg
					_ = tmpl.ExecuteTemplate(w, "index.html", data)
					return
				}
				cacheSet(key, info)
			}

			data.Info = info
			data.Alerts = weather.Alerts(info)
			data.Quote = weather.QuoteFromIcon(info.Current.Icon)
			data.Advice = weather.Advice(info.Current.FeelsLike, info.TempUnit)
		}

		if err := tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Weather Web App running at http://localhost:%s  (cache TTL: %s)", port, cacheTTL)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

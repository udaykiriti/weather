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

// ── Template ────────────────────────────────────────────────────────────────

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
		cx, cy    = 12.0, 12.0
		r         = 9.0
		top       = cy - r // 3.0
		bot       = cy + r // 21.0
		litColor  = "#fef3c7" // amber-100
		darkColor = "#1e293b" // slate-800
		strokeClr = "#6366f1" // indigo-500
	)

	// Full and new moon special cases
	if phase < 0.02 || phase > 0.98 {
		return template.HTML(fmt.Sprintf(
			`<svg viewBox="0 0 24 24" width="26" height="26"><circle cx="12" cy="12" r="9" fill="%s" stroke="%s" stroke-width="1.2"/></svg>`,
			darkColor, strokeClr))
	}
	if phase > 0.48 && phase < 0.52 {
		return template.HTML(fmt.Sprintf(
			`<svg viewBox="0 0 24 24" width="26" height="26"><circle cx="12" cy="12" r="9" fill="%s" stroke="%s" stroke-width="1.2"/></svg>`,
			litColor, strokeClr))
	}

	waxing := phase < 0.5
	illum := phase * 2
	if !waxing {
		illum = (1 - phase) * 2
	}

	// Inner ellipse x-radius: 0 at quarter, r at new/full
	rx := math.Max(0.5, r*math.Abs(math.Cos(math.Pi*illum)))

	// Outer sweep: 1 = right semicircle (waxing), 0 = left (waning)
	outerSweep, innerSweep := 1, 0
	if !waxing {
		outerSweep = 0
		innerSweep = 1
	}
	// Gibbous: inner arc extends toward shadow side (flip inner sweep)
	if illum > 0.5 {
		if waxing {
			innerSweep = 1
		} else {
			innerSweep = 0
		}
	}

	litPath := fmt.Sprintf(
		"M %.2f %.2f A %.2f %.2f 0 0 %d %.2f %.2f A %.2f %.2f 0 0 %d %.2f %.2f Z",
		cx, top,
		r, r, outerSweep, cx, bot,
		rx, r, innerSweep, cx, top,
	)
	svg := fmt.Sprintf(
		`<svg viewBox="0 0 24 24" width="26" height="26"><circle cx="12" cy="12" r="9" fill="%s" stroke="%s" stroke-width="1.2"/><path d="%s" fill="%s"/></svg>`,
		darkColor, strokeClr, litPath, litColor,
	)
	return template.HTML(svg)
}

var tmpl = template.Must(
	template.New("").Funcs(template.FuncMap{
		// Math funcs use float64 for smooth SVG positioning.
		"subf": func(a, b float64) float64 { return a - b },
		"mulf": func(a, b float64) float64 { return a * b },
		// arcY: maps 0-100 progress to SVG y on a half-ellipse
		// viewBox 0 0 400 80, horizon y=72, semi-axes rx=190 ry=72
		"arcY": func(pct float64) float64 {
			t := pct / 100.0
			y := 72.0 - 72.0*math.Sin(math.Pi*t)
			return math.Max(-5, y)
		},
		// arcX: maps 0-100 progress to SVG x (linear across viewBox)
		"arcX": func(pct float64) float64 {
			return 10 + (pct/100)*380
		},
		"not":          func(b bool) bool { return !b },
		"uvLevel":      weather.UVLevel,
		"uvColorClass": weather.UVColorClass,
		"windCompass":  weather.WindCompass,
		"moonPhaseSVG": func(phase float64) template.HTML { return moonPhaseSVG(phase) },
		// dewComfort returns a comfort label for dew point, normalising to °C first.
		"dewComfort": func(dp float64, unit string) string {
			c := dp
			if unit == "°F" {
				c = (dp - 32) * 5 / 9
			}
			switch {
			case c > 21:
				return "Oppressive"
			case c > 18:
				return "Humid"
			case c > 13:
				return "Comfortable"
			case c > 7:
				return "Dry"
			default:
				return "Very Dry"
			}
		},
		"outfitSVG": func(icon string) template.HTML {
			if s, ok := outfitIcons[icon]; ok {
				return s
			}
			return outfitIconFallback
		},
		// hourlyPrecipSVG builds an inline SVG bar chart of hourly rain probability.
		"hourlyPrecipSVG": func(hourly []weather.HourlyPoint) template.HTML {
			if len(hourly) == 0 {
				return ""
			}
			const (
				slot      = 30  // px per bar slot
				barW      = 27  // drawn bar width (3px gap)
				baseline  = 70  // y of the x-axis line
				maxBarH   = 60  // height for 100%
				labelY    = 88  // y of time labels
				viewH     = 92  // total SVG height
				viewW     = 720 // 24 × 30
			)
			barColor := func(pct int) string {
				switch {
				case pct >= 80:
					return "#1d4ed8"
				case pct >= 60:
					return "#3b82f6"
				case pct >= 30:
					return "#93c5fd"
				default:
					return "#dbeafe"
				}
			}
			var sb strings.Builder
			sb.Grow(4096) // pre-allocate: 24 bars × ~120 bytes each + header/footer
			fmt.Fprintf(&sb, `<svg viewBox="0 0 %d %d" width="100%%" preserveAspectRatio="none" class="precip-svg" aria-label="Hourly precipitation probability">`, viewW, viewH)
			// 50% gridline
			fmt.Fprintf(&sb, `<line x1="0" y1="%d" x2="%d" y2="%d" class="pchart-grid"/>`,
				baseline-maxBarH/2, viewW, baseline-maxBarH/2)
			// 50% label (fixed position, not stretched)
			fmt.Fprintf(&sb, `<text x="2" y="%d" class="pchart-lbl" text-anchor="start">50%%</text>`,
				baseline-maxBarH/2-2)

			n := len(hourly)
			if n > 24 {
				n = 24
			}
			for i := 0; i < n; i++ {
				h := hourly[i]
				pct := h.PrecipProb
				bh := maxBarH * pct / 100
				bx := i * slot
				by := baseline - bh
				// bar (min height 1 so 0% bars are still visible as a thin line)
				if bh == 0 {
					bh = 1
					by = baseline - 1
				}
				fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s" rx="2"/>`,
					bx, by, barW, bh, barColor(pct))
				// time label every 4 hours
				if i%4 == 0 {
					lx := bx + slot/2
					fmt.Fprintf(&sb, `<text x="%d" y="%d" class="pchart-lbl" text-anchor="middle">%s</text>`,
						lx, labelY, h.Time)
				}
			}
			// baseline
			fmt.Fprintf(&sb, `<line x1="0" y1="%d" x2="%d" y2="%d" class="pchart-base"/>`,
				baseline, viewW, baseline)
			sb.WriteString(`</svg>`)
			return template.HTML(sb.String())
		},
		"agreeColor": func(pct int) string {
			switch {
			case pct >= 80:
				return "#22c55e"
			case pct >= 50:
				return "#f59e0b"
			default:
				return "#ef4444"
			}
		},
		// consBarW: returns 0-100 bar width % for a model's temp relative to min/max range
		"consBarW": func(temp, minT, maxT float64) float64 {
			spread := maxT - minT
			if spread <= 0 {
				return 50.0
			}
			v := (temp-minT)/spread*80.0 + 10.0 // 10-90% range
			if v < 5 {
				v = 5
			}
			if v > 95 {
				v = 95
			}
			return v
		},
	}).ParseFiles("templates/index.html"),
)

// ── Cache ────────────────────────────────────────────────────────────────────

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

func cacheKey(city, units string) string { return strings.ToLower(city) + "|" + units }

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
		var oldest string
		var oldestTime time.Time
		for k, e := range cache {
			if oldest == "" || e.createdAt.Before(oldestTime) {
				oldest = k
				oldestTime = e.createdAt
			}
		}
		delete(cache, oldest)
	}

	cache[key] = &cacheEntry{
		info:      info,
		expires:   time.Now().Add(cacheTTL),
		createdAt: time.Now(),
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

// ── Page data ────────────────────────────────────────────────────────────────

type PageData struct {
	City   string
	Units  string
	Info   *weather.WeatherInfo
	Alerts []weather.Alert
	Quote  string
	Advice string
	Error  string
}

// ── Main ─────────────────────────────────────────────────────────────────────

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
		city, err := client.ReverseGeocode(lat, lon)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
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
				info, err = client.GetWeather(city, units)
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

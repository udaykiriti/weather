package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"WeatherApp/weather"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	italic = "\033[3m"

	red     = "\033[31m"
	yellow  = "\033[33m"
	green   = "\033[32m"
	blue    = "\033[34m"
	cyan    = "\033[36m"
	magenta = "\033[35m"
	white   = "\033[97m"
	orange  = "\033[38;5;208m"

	bgRed    = "\033[41m"
	bgYellow = "\033[43m"
	bgBlue   = "\033[44m"
)

const W = 70 // total line width for boxes

func clr(color, s string) string { return color + s + reset }

func tempColor(temp float64, unit string) string {
	t := temp
	if unit == "°F" {
		t = (temp - 32) * 5 / 9
	}
	switch {
	case t <= 0:
		return "\033[94m" // bright blue
	case t <= 10:
		return cyan
	case t <= 20:
		return green
	case t <= 30:
		return yellow
	default:
		return red
	}
}

func uvColor(uv float64) string {
	switch {
	case uv >= 11:
		return magenta
	case uv >= 8:
		return red
	case uv >= 6:
		return orange
	case uv >= 3:
		return yellow
	default:
		return green
	}
}

func topBar(title string) string {
	inner := W - 4 // space inside ┌ ... ┐
	if title == "" {
		return "  ┌" + strings.Repeat("─", inner) + "┐"
	}
	dash := inner - 2 - len(title) - 1 // "─ " + title + " "
	if dash < 0 {
		dash = 0
	}
	return "  ┌─ " + clr(bold+white, title) + " " + strings.Repeat("─", dash) + "┐"
}

func botBar() string { return "  └" + strings.Repeat("─", W-4) + "┘" }
func midBar() string { return "  ├" + strings.Repeat("─", W-4) + "┤" }

// row adds the left border prefix. Caller provides the content string.
func row(s string) string { return "  │  " + s }
func blankRow() string    { return "  │" }

func progressBar(pct, width int, barColor string) string {
	n := pct * width / 100
	if n > width {
		n = width
	}
	return clr(barColor, strings.Repeat("█", n)) +
		clr(dim, strings.Repeat("░", width-n))
}

func alertColor(level weather.AlertLevel) string {
	switch level {
	case weather.AlertDanger:
		return bgRed + bold + white
	case weather.AlertWarning:
		return bgYellow + bold + "\033[30m"
	default:
		return bgBlue + bold + white
	}
}

func alertPrefix(level weather.AlertLevel) string {
	switch level {
	case weather.AlertDanger:
		return " !! DANGER  "
	case weather.AlertWarning:
		return " !  WARNING "
	default:
		return " i  INFO    "
	}
}

// sunLine renders the daylight arc as a coloured text bar.
// Dots (·) represent sky, dashes (─) represent horizon edges.
func sunLine(sun weather.SunBar, width int) string {
	if sun.SunriseTime == "" {
		return strings.Repeat("─", width)
	}
	pct := int(sun.SunPositionPct)
	pos := pct * width / 100
	if pos >= width {
		pos = width - 1
	}

	bar := make([]rune, width)
	for i := range bar {
		t := float64(i) / float64(width-1)
		if math.Sin(math.Pi*t) > 0.25 {
			bar[i] = '·' // sky
		} else {
			bar[i] = '─' // horizon
		}
	}
	if sun.IsDay {
		bar[pos] = '☀'
	} else {
		bar[pos] = '☽'
	}

	colored := make([]string, width)
	for i, ch := range bar {
		switch {
		case i < pos:
			colored[i] = clr(yellow, string(ch))
		case i == pos:
			if sun.IsDay {
				colored[i] = clr(bold+yellow, string(ch))
			} else {
				colored[i] = clr(bold+"\033[94m", string(ch))
			}
		default:
			colored[i] = clr(dim, string(ch))
		}
	}
	return strings.Join(colored, "")
}

func wordWrap(s string, maxWidth int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) <= maxWidth {
			line += " " + w
		} else {
			lines = append(lines, line)
			line = w
		}
	}
	return append(lines, line)
}

func startSpinner(msg string) chan struct{} {
	done := make(chan struct{})
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	go func() {
		i := 0
		for {
			select {
			case <-done:
				fmt.Print("\r\033[K")
				return
			default:
				fmt.Printf("\r  %s %s", clr(cyan, frames[i%len(frames)]), msg)
				time.Sleep(80 * time.Millisecond)
				i++
			}
		}
	}()
	return done
}

/* Main func*/
func main() {
	cityFlag := flag.String("city", "", "City name (or first positional argument)")
	units := flag.String("units", "metric", "Units: metric (°C/km·h) or imperial (°F/mph)")
	flag.Parse()

	// Support positional arg: weather-cli London  or  weather-cli New York
	city := *cityFlag
	if city == "" {
		if args := flag.Args(); len(args) > 0 {
			city = strings.Join(args, " ")
		} else {
			city = "London"
		}
	}

	fmt.Println()
	done := startSpinner("Fetching weather for " + clr(bold+white, city) + " ...")

	client := weather.NewClient()
	info, err := client.GetWeather(city, *units)
	close(done)
	time.Sleep(20 * time.Millisecond) // let spinner goroutine clear line

	if err != nil {
		fmt.Println()
		fmt.Fprintf(os.Stderr, "  %sError:%s %v\n\n", red+bold, reset, err)
		os.Exit(1)
	}

	cur := info.Current
	unitLabel := "Metric"
	if *units == "imperial" {
		unitLabel = "Imperial"
	}

	fmt.Println()

	cityUpper := strings.ToUpper(info.CityName) + ", " + strings.ToUpper(info.Country)
	hLeft := "  WEATHER  —  " + cityUpper
	hRight := unitLabel + "  ●  LIVE"
	innerW := W - 4
	hPad := innerW - len(hLeft) + 2 - len(hRight) // +2 for "  " prefix before ║
	if hPad < 1 {
		hPad = 1
	}

	fmt.Println("  ╔" + strings.Repeat("═", W-4) + "╗")
	fmt.Printf("  ║%s%s%s%s║\n",
		clr(bold+white, hLeft),
		strings.Repeat(" ", hPad),
		clr(dim, hRight),
		"  ")
	fmt.Println("  ╚" + strings.Repeat("═", W-4) + "╝")
	fmt.Println()

	alerts := weather.Alerts(info)
	if len(alerts) > 0 {
		for _, a := range alerts {
			prefix := alertColor(a.Level) + alertPrefix(a.Level) + reset
			fmt.Printf("  %s %s\n", prefix, clr(bold, a.Title))
			fmt.Printf("     %s\n", clr(dim, a.Message))
		}
		fmt.Println()
	}

	fmt.Println(topBar("Current Conditions"))

	quote := weather.QuoteFromIcon(cur.Icon)
	for _, line := range wordWrap(quote, W-8) {
		fmt.Println(row(clr(italic+dim+magenta, line)))
	}
	fmt.Println(blankRow())

	tc := tempColor(cur.Temp, info.TempUnit)
	fc := tempColor(cur.FeelsLike, info.TempUnit)
	tempStr := clr(bold+tc, fmt.Sprintf("%.1f%s", cur.Temp, info.TempUnit))
	feelStr := clr(fc, fmt.Sprintf("%.1f%s", cur.FeelsLike, info.TempUnit))
	condStr := clr(bold+white, cur.Description)
	fmt.Println(row(fmt.Sprintf("%-30s  %s %s  %s %s",
		condStr,
		clr(dim+cyan, "Temp"), tempStr,
		clr(dim+cyan, "Feels"), feelStr,
	)))

	advice := weather.Advice(cur.FeelsLike, info.TempUnit)
	fmt.Println(row(clr(dim, "  → ") + clr(green, advice)))
	fmt.Println(blankRow())

	// Stats row 1: Humidity + Cloud Cover
	humBar := progressBar(cur.Humidity, 14, cyan)
	cldBar := progressBar(cur.CloudCover, 14, blue)
	fmt.Println(row(fmt.Sprintf(
		"%s %s %3d%%    %s %s %3d%%",
		clr(dim+cyan, "Humidity  "), humBar, cur.Humidity,
		clr(dim+cyan, "Cloud    "), cldBar, cur.CloudCover,
	)))

	// Stats row 2: Pressure + Wind (with direction)
	windDir := weather.WindCompass(cur.WindDir)
	fmt.Println(row(fmt.Sprintf(
		"%s %s      %s %s %s",
		clr(dim+cyan, "Pressure  "), clr(white, fmt.Sprintf("%.0f hPa", cur.Pressure)),
		clr(dim+cyan, "Wind     "), clr(white, fmt.Sprintf("%.1f %s", cur.WindSpeed, info.WindUnit)),
		clr(dim+cyan, windDir),
	)))

	// Stats row 2b: Dew Point + FeelsLike
	fmt.Println(row(fmt.Sprintf(
		"%s %s      %s %s",
		clr(dim+cyan, "Dew Point "), clr(white, fmt.Sprintf("%.1f%s", cur.DewPoint, info.TempUnit)),
		clr(dim+cyan, "Feels    "), clr(white, fmt.Sprintf("%.1f%s", cur.FeelsLike, info.TempUnit)),
	)))

	// Stats row 3: UV Index + Updated
	uvc := uvColor(cur.UVIndex)
	uvLvl := weather.UVLevel(cur.UVIndex)
	fmt.Println(row(fmt.Sprintf(
		"%s %s %s      %s %s",
		clr(dim+cyan, "UV Index  "),
		clr(uvc, fmt.Sprintf("%.1f", cur.UVIndex)),
		clr(uvc, "("+uvLvl+")"),
		clr(dim+cyan, "Updated  "), clr(dim, cur.Time),
	)))
	fmt.Println(row(fmt.Sprintf(
		"%s %s",
		clr(dim+cyan, "UV Advice "), clr(dim, weather.UVAdvice(cur.UVIndex)),
	)))

	fmt.Println(botBar())
	fmt.Println()

	if info.Sun.SunriseTime != "" {
		arcWidth := W - 22 // leave room for sunrise/sunset labels
		fmt.Println(topBar("Daylight  " + info.Sun.DaylightHours))

		arcLine := clr(yellow, "☀ "+info.Sun.SunriseTime) +
			"  " + sunLine(info.Sun, arcWidth) +
			"  " + clr(dim, info.Sun.SunsetTime+" ☽")
		fmt.Println(row(arcLine))

		if info.Sun.IsDay {
			// indent the ↑ marker under the sun position
			indent := int(info.Sun.SunPositionPct) * arcWidth / 100
			labelOffset := 14 // "☀ HH:MM  " prefix width
			if indent > arcWidth-8 {
				indent = arcWidth - 8
			}
			fmt.Println(row(fmt.Sprintf("%s%s %s",
				strings.Repeat(" ", labelOffset+indent),
				clr(dim, "↑"),
				clr(bold+yellow, "now "+info.Sun.CurrentTime),
			)))
		}

		fmt.Println(botBar())
		fmt.Println()
	}

	if info.Sun.MoonPhaseName != "" {
		moonIcon := moonPhaseIcon(info.Sun.MoonPhase)
		illumPct := 0
		if info.Sun.MoonPhase > 0.5 {
			illumPct = int((1 - info.Sun.MoonPhase) * 100 * 2)
		} else {
			illumPct = int(info.Sun.MoonPhase * 100 * 2)
		}
		fmt.Println(topBar("Moon Phase"))
		fmt.Println(row(fmt.Sprintf(
			"%s  %s  %s",
			clr(bold+"\033[94m", moonIcon),
			clr(bold+white, info.Sun.MoonPhaseName),
			clr(dim, fmt.Sprintf("(%d%% illuminated)", illumPct)),
		)))
		fmt.Println(botBar())
		fmt.Println()
	}
	if len(info.Hourly) > 0 {
		fmt.Println(topBar("Next 24 Hours"))

		// Sparkline: map temperatures to block characters
		sparkChars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
		temps := make([]float64, len(info.Hourly))
		minT, maxT := info.Hourly[0].Temp, info.Hourly[0].Temp
		for i, h := range info.Hourly {
			temps[i] = h.Temp
			if h.Temp < minT {
				minT = h.Temp
			}
			if h.Temp > maxT {
				maxT = h.Temp
			}
		}
		spark := make([]string, len(temps))
		for i, t := range temps {
			var idx int
			if maxT > minT {
				idx = int((t - minT) / (maxT - minT) * 7)
			}
			if idx < 0 {
				idx = 0
			}
			if idx > 7 {
				idx = 7
			}
			spark[i] = clr(tempColor(t, info.TempUnit), string(sparkChars[idx]))
		}
		fmt.Println(row(clr(dim+cyan, "Temp  ") + strings.Join(spark, "") +
			clr(dim, fmt.Sprintf("  %.0f%s–%.0f%s", minT, info.TempUnit, maxT, info.TempUnit))))
		fmt.Println(blankRow())

		// Compact table: every 3 hours (8 rows)
		fmt.Println(row(fmt.Sprintf("%-5s  %-14s  %5s  %-12s  %s",
			clr(dim, "TIME"), clr(dim, "CONDITION"),
			clr(dim, "TEMP"), clr(dim, "RAIN"),
			clr(dim, "WIND"),
		)))
		fmt.Println(row(strings.Repeat("─", W-10)))

		for i, h := range info.Hourly {
			if i%3 != 0 {
				continue
			}
			tc := tempColor(h.Temp, info.TempUnit)
			pBars := h.PrecipProb / 10
			pBar := clr("\033[34m", strings.Repeat("█", pBars)) +
				clr(dim, strings.Repeat("░", 10-pBars))
			cond := h.Description
			if len(cond) > 14 {
				cond = cond[:13] + "…"
			}
			fmt.Println(row(fmt.Sprintf("%-5s  %-14s  %s  %s %s  %s",
				clr(bold, h.Time),
				clr(dim, cond),
				clr(tc, fmt.Sprintf("%4.0f%s", h.Temp, info.TempUnit)),
				pBar,
				clr("\033[34m", fmt.Sprintf("%3d%%", h.PrecipProb)),
				clr(blue, fmt.Sprintf("%.0f %s", h.WindSpeed, info.WindUnit)),
			)))
		}

		fmt.Println(botBar())
		fmt.Println()
	}

	fmt.Println(topBar("5-Day Forecast"))
	fmt.Println(row(fmt.Sprintf("%-10s  %-15s  %5s  %5s  %8s  %-12s",
		clr(dim, "DATE"),
		clr(dim, "CONDITION"),
		clr(dim, "HI"),
		clr(dim, "LO"),
		clr(dim, "WIND"),
		clr(dim, "RAIN"),
	)))
	fmt.Println(row(strings.Repeat("─", W-10)))

	for _, day := range info.Forecast {
		htc := tempColor(day.TempMax, info.TempUnit)
		ltc := tempColor(day.TempMin, info.TempUnit)
		hiStr := clr(htc, fmt.Sprintf("%4.0f%s", day.TempMax, info.TempUnit))
		loStr := clr(ltc, fmt.Sprintf("%4.0f%s", day.TempMin, info.TempUnit))
		wdStr := clr(blue, fmt.Sprintf("%5.0f %s", day.WindMax, info.WindUnit))

		pBars := day.PrecipProb / 10
		pBar := clr("\033[34m", strings.Repeat("█", pBars)) +
			clr(dim, strings.Repeat("░", 10-pBars))
		pctStr := clr("\033[34m", fmt.Sprintf("%3d%%", day.PrecipProb))

		cond := day.Description
		if len(cond) > 15 {
			cond = cond[:14] + "…"
		}

		fmt.Println(row(fmt.Sprintf("%-10s  %-15s  %s  %s  %s  %s %s",
			clr(bold, day.Date),
			clr(dim, cond),
			hiStr, loStr, wdStr,
			pBar, pctStr,
		)))
	}

	fmt.Println(botBar())
	fmt.Println()

	outfit := info.Outfit
	if len(outfit.Items) > 0 {
		fmt.Println(topBar("What to Wear"))
		fmt.Println(row(clr(dim+cyan, outfit.Headline)))
		fmt.Println(row(strings.Repeat("─", W-10)))
		for _, item := range outfit.Items {
			emoji := outfitEmoji(item.Icon)
			fmt.Println(row(fmt.Sprintf("%s  %-16s  %s",
				emoji,
				clr(bold, item.Label),
				clr(dim, item.Note),
			)))
		}
		fmt.Println(botBar())
		fmt.Println()
	}

	if info.Consensus != nil && info.Consensus.AvailCount > 0 {
		cons := info.Consensus
		fmt.Println(topBar("Model Consensus"))

		agreeClr := green
		if cons.AgreePct < 50 {
			agreeClr = red
		} else if cons.AgreePct < 80 {
			agreeClr = yellow
		}

		fmt.Println(row(fmt.Sprintf(
			"%s %s    %s %s    %s %s",
			clr(dim+cyan, "Agreement"),
			clr(agreeClr+bold, fmt.Sprintf("%s (%d%%)", cons.Agreement, cons.AgreePct)),
			clr(dim+cyan, "Avg"),
			clr(white, fmt.Sprintf("%.1f%s", cons.AvgTemp, info.TempUnit)),
			clr(dim+cyan, "Spread"),
			clr(agreeClr, fmt.Sprintf("%.1f%s", cons.Spread, info.TempUnit)),
		)))
		fmt.Println(blankRow())

		barW := 22
		for _, r := range cons.Models {
			name := fmt.Sprintf("%-14s", r.Model)
			if !r.Available {
				fmt.Println(row(clr(dim, name+"  unavailable")))
				continue
			}
			var filled int
			if cons.Spread > 0.1 {
				filled = int((r.Temp-cons.MinTemp)/cons.Spread*float64(barW-2)) + 1
			} else {
				filled = barW / 2
			}
			if filled < 1 {
				filled = 1
			}
			if filled > barW {
				filled = barW
			}
			bar := clr(green, strings.Repeat("█", filled)) +
				clr(dim, strings.Repeat("░", barW-filled))
			fmt.Println(row(fmt.Sprintf("  %s  [%s]  %s",
				clr(cyan, name),
				bar,
				clr(white, fmt.Sprintf("%.1f%s", r.Temp, info.TempUnit)),
			)))
		}

		fmt.Println(blankRow())
		fmt.Println(row(fmt.Sprintf("%s  %.1f%s — %.1f%s",
			clr(dim+cyan, "Range            "),
			cons.MinTemp, info.TempUnit,
			cons.MaxTemp, info.TempUnit,
		)))
		fmt.Println(botBar())
		fmt.Println()
	}

	fmt.Println("  " + clr(dim, "Data · open-meteo.com · free · no API key required"))
	fmt.Println()
}

func outfitEmoji(icon string) string {
	m := map[string]string{
		"thermal":    "[~~]",
		"sweater":    "[\\~/]",
		"longsleeve": "[>--]",
		"tshirt":     "[T]  ",
		"coat":       "[|||]",
		"jacket":     "[/|\\]",
		"windbreaker":"[>>>]",
		"umbrella":   " ( ) ",
		"raincoat":   "[:::]",
		"sunscreen":  "[SPF]",
		"sunglasses": "(_8_)",
		"beanie":     "[^^^]",
		"hat":        "[ ^ ]",
		"boots":      "[|||]",
		"sandals":    "[ _ ]",
	}
	if e, ok := m[icon]; ok {
		return e
	}
	return "[---]"
}

func moonPhaseIcon(phase float64) string {
switch {
case phase < 0.034 || phase >= 0.966:
return "( )" // new moon
case phase < 0.216:
return "()" // waxing crescent
case phase < 0.284:
return "(|" // first quarter
case phase < 0.466:
return "(@" // waxing gibbous
case phase < 0.534:
return "(@)" // full moon
case phase < 0.716:
return "@)" // waning gibbous
case phase < 0.784:
return "|)" // last quarter
default:
return "()" // waning crescent
}
}

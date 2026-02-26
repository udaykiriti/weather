package weather

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// forecastModels lists the free Open-Meteo models used for consensus.
var forecastModels = []struct {
	Name  string
	Param string
}{
	{"ECMWF", "ecmwf_ifs025"},
	{"ICON", "icon_seamless"},
	{"Météo-France", "meteofrance_seamless"},
	{"MET Norway", "metno_seamless"},
}

// ModelReading is current weather data from a single forecast model.
type ModelReading struct {
	Model       string
	Temp        float64
	Humidity    int
	WindSpeed   float64
	Pressure    float64
	WeatherCode int
	Available   bool
	Err         string
}

// ConsensusInfo holds per-model readings and derived consensus statistics.
type ConsensusInfo struct {
	Models     []ModelReading
	AvailCount int

	// Averages across all available models
	AvgTemp     float64
	AvgHumidity int
	AvgWind     float64
	AvgPressure float64

	// Temperature spread (max - min) as a disagreement measure
	MinTemp float64
	MaxTemp float64
	Spread  float64

	// Agreement score
	Agreement string // "High" / "Moderate" / "Low"
	AgreePct  int    // 0-100
}

// modelCurrentRaw uses pointers so null JSON fields don't cause decode errors.
type modelCurrentRaw struct {
	Temp     *float64 `json:"temperature_2m"`
	Humidity *int     `json:"relative_humidity_2m"`
	Wind     *float64 `json:"wind_speed_10m"`
	Pressure *float64 `json:"pressure_msl"`
	WMOCode  *int     `json:"weather_code"`
}
type modelResponse struct {
	Current modelCurrentRaw `json:"current"`
}

// consensusClient is a separate HTTP client with a shorter timeout so slow
// model fetches never block the main request beyond 6 seconds.
var consensusClient = &Client{HTTP: &http.Client{Timeout: 6 * time.Second}}

// fetchModel fetches current conditions from one Open-Meteo model.
// tempUnit must be "celsius" or "fahrenheit"; windUnit "kmh" or "mph".
func fetchModel(name, modelParam string, lat, lon float64, timezone, tempUnit, windUnit string) ModelReading {
	r := ModelReading{Model: name}

	u := fmt.Sprintf(
		"%s?latitude=%.4f&longitude=%.4f"+
			"&current=temperature_2m,relative_humidity_2m,wind_speed_10m,pressure_msl,weather_code"+
			"&temperature_unit=%s&wind_speed_unit=%s&timezone=%s&models=%s",
		forecastURL, lat, lon, tempUnit, windUnit,
		url.QueryEscape(timezone), modelParam,
	)

	var resp modelResponse
	if err := consensusClient.getJSON(u, &resp); err != nil {
		r.Err = err.Error()
		return r
	}

	cur := resp.Current
	if cur.Temp == nil || cur.Humidity == nil {
		r.Err = "incomplete data"
		return r
	}

	r.Temp = *cur.Temp
	r.Humidity = *cur.Humidity
	if cur.Wind != nil {
		r.WindSpeed = *cur.Wind
	}
	if cur.Pressure != nil {
		r.Pressure = *cur.Pressure
	}
	if cur.WMOCode != nil {
		r.WeatherCode = *cur.WMOCode
	}
	r.Available = true
	return r
}

// FetchConsensus fetches 4 weather models in parallel and computes agreement stats.
// tempUnit must be "celsius"/"fahrenheit"; windUnit "kmh"/"mph".
func (c *Client) FetchConsensus(lat, lon float64, timezone, tempUnit, windUnit string) *ConsensusInfo {
	readings := make([]ModelReading, len(forecastModels))
	var wg sync.WaitGroup

	for i, m := range forecastModels {
		wg.Add(1)
		go func(idx int, name, param string) {
			defer wg.Done()
			readings[idx] = fetchModel(name, param, lat, lon, timezone, tempUnit, windUnit)
		}(i, m.Name, m.Param)
	}
	wg.Wait()

	cons := &ConsensusInfo{Models: readings}

	var sumTemp, sumWind, sumPressure float64
	var sumHumidity, count int
	minTemp := math.MaxFloat64
	maxTemp := -math.MaxFloat64

	for _, r := range readings {
		if !r.Available {
			continue
		}
		count++
		sumTemp += r.Temp
		sumHumidity += r.Humidity
		sumWind += r.WindSpeed
		sumPressure += r.Pressure
		if r.Temp < minTemp {
			minTemp = r.Temp
		}
		if r.Temp > maxTemp {
			maxTemp = r.Temp
		}
	}

	if count == 0 {
		// No models responded; all stats stay zero
		return cons
	}

	cons.AvailCount = count
	cons.MinTemp = math.Round(minTemp*10) / 10
	cons.MaxTemp = math.Round(maxTemp*10) / 10
	cons.AvgTemp = math.Round(sumTemp/float64(count)*10) / 10
	cons.AvgHumidity = sumHumidity / count
	cons.AvgWind = math.Round(sumWind/float64(count)*10) / 10
	cons.AvgPressure = math.Round(sumPressure/float64(count)*10) / 10
	cons.Spread = math.Round((cons.MaxTemp-cons.MinTemp)*10) / 10

	// Agreement: 100% at 0°C spread, −12% per degree of spread
	pct := 100 - int(cons.Spread*12)
	if pct < 0 {
		pct = 0
	}
	cons.AgreePct = pct
	switch {
	case pct >= 80:
		cons.Agreement = "High"
	case pct >= 50:
		cons.Agreement = "Moderate"
	default:
		cons.Agreement = "Low"
	}

	return cons
}

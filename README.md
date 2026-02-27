# Weather App

A weather application built in Go with two interfaces: a command-line tool and a web server.
Uses the [Open-Meteo](https://open-meteo.com/) API - completely free, no API key or sign-up required.

---

> [!NOTE]
> This app requires outbound internet access to `geocoding-api.open-meteo.com` and
> `api.open-meteo.com`. If you are behind a corporate proxy or a restrictive firewall,
> requests may time out. See the [Network / Proxy](#network--proxy) section below.
---

## Features

- Real-time weather conditions (temperature, humidity, pressure, wind, UV index)
- 5-day forecast with precipitation probability
- Sunrise/sunset arc with daylight hours
- Weather alerts (heat, frost, storm, high UV, and more)
- Funny weather-matched quotes and feels-like advice
- Multi-model consensus (ECMWF, ICON, Meteo-France, MET Norway)
- Claymorphism + brutalism web UI with weather icons
- Full-colour ANSI CLI with animated spinner and box-drawing layout
- Responsive web design (mobile-friendly)
- 10-minute in-memory cache on the web server

---

## Project Structure

```
WeatherApp/
├── weather/
│   ├── client.go        # Open-Meteo API client, types, geocoding, forecast
│   ├── alerts.go        # Weather alert triggers (12 conditions, 3 severity levels)
│   ├── quotes.go        # Funny weather quotes and feels-like advice
│   ├── uv.go            # UV index level, advice, and colour helpers
│   └── consensus.go     # 4-model parallel forecast consensus
├── cmd/
│   └── cli/
│       └── main.go      # CLI application
├── templates/
│   └── index.html       # Web UI template (claymorphism + brutalism)
├── static/              # Static assets
├── main.go              # Web server
├── Makefile             # Build targets
├── run.sh               # Shell script build/run helper
├── .env.example         # Environment variable template
├── go.mod
└── go.sum
```

---

## Prerequisites

- Go 1.21 or higher
- No API key needed

---

## Setup

1. Clone or download the repository.

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Build using Make or the shell script:

   ```bash
   make build
   # or
   ./run.sh build
   ```

---

## Build Commands

### Using Make

```bash
make cli        # Build the CLI binary
make web        # Build the web server binary
make build      # Build both
make run-cli    # Build and run CLI (default city: London)
make run-web    # Build and run web server
make fmt        # Format all Go source files
make vet        # Run go vet
make clean      # Remove built binaries
```

Pass a city with `ARGS`:

```bash
make run-cli ARGS="Tokyo"
make run-cli ARGS="New York"
```

### Using run.sh

```bash
./run.sh cli
./run.sh web
./run.sh build
./run.sh run-cli London
./run.sh run-cli "New York"
./run.sh run-web
./run.sh fmt
./run.sh vet
./run.sh clean
```

---

## CLI Usage

```bash
./weather-cli [city]
./weather-cli -city <city> [-units metric|imperial]
```

| Flag     | Default | Description                                         |
|----------|---------|-----------------------------------------------------|
| `city`   | London  | City as a positional argument                       |
| `-city`  | London  | City name flag                                      |
| `-units` | metric  | Unit system: `metric` (C/km/h) or `imperial` (F/mph)|

### Examples

```bash
./weather-cli London
./weather-cli "New York"
./weather-cli -city Tokyo
./weather-cli -city Mumbai -units metric
./weather-cli -city "New York" -units imperial
```

### CLI Output Sections

- Animated spinner while fetching data
- Boxed header with city, country, and unit system
- Weather alerts (colour-coded by severity)
- Current conditions: temperature (colour by value), feels like, humidity, cloud cover, pressure, wind, UV index
- Daylight arc with sunrise, sunset, and current sun position
- 5-day forecast table with colour-coded temperatures and precipitation bars
- Multi-model consensus with per-model temperature bars

---

## Web App Usage

Start the server:

```bash
make run-web
# or
./run.sh run-web
# or directly
./weather-web
```

Then open [http://localhost:8080](http://localhost:8080) in your browser.

- Enter a city name in the search box.
- Choose Celsius or Fahrenheit.
- View current conditions, alerts, quotes, UV index, sunrise/sunset arc, 5-day forecast, and model consensus.

To use a custom port:

```bash
PORT=3000 ./weather-web
```

Or add it to a `.env` file:

```
PORT=3000
```

---

## API Reference

This app uses [Open-Meteo](https://open-meteo.com/) - free and open-source, no registration needed.

| API                    | Description                                      |
|------------------------|--------------------------------------------------|
| Geocoding API          | Resolves city name to coordinates and timezone   |
| Forecast API (current) | Temperature, wind, humidity, UV, cloud cover     |
| Forecast API (daily)   | 5-day high/low, wind, precipitation probability  |
| Forecast API (models)  | ECMWF, ICON, Meteo-France, MET Norway consensus  |

Weather conditions are decoded from [WMO Weather Codes](https://open-meteo.com/en/docs#weathervariables).

---

## Environment Variables

| Variable | Required | Default | Description     |
|----------|----------|---------|-----------------|
| `PORT`   | No       | `8080`  | Web server port |

---

## Network / Proxy

> [!WARNING]
> If the CLI or web app shows `dial tcp: lookup ... connection refused` or
> `context deadline exceeded`, your system's DNS resolver is unreachable or
> your network blocks outbound HTTPS.

The app includes a built-in DNS fallback: when the system resolver fails, it automatically
retries against Google DNS (`8.8.8.8`) and Cloudflare DNS (`1.1.1.1`).

If you are behind an HTTP/HTTPS proxy, set the standard Go proxy environment variables
before running:

```bash
export HTTPS_PROXY=http://proxy.example.com:8080
export HTTP_PROXY=http://proxy.example.com:8080
export NO_PROXY=localhost,127.0.0.1
make run-cli ARGS="London"
```

> [!TIP]
> Go's `net/http` package respects `HTTPS_PROXY` and `HTTP_PROXY` automatically —
> no code changes are needed.

---

## Tips

- City names with spaces must be quoted: `./weather-cli "New York"` or `make run-cli ARGS="New York"`.
- The web server caches results for 10 minutes per city. Use the unit toggle on the page to
  switch between Celsius and Fahrenheit.
- The geolocation button in the web UI calls a server-side proxy (`/api/reverse`) to avoid
  browser CORS restrictions when resolving GPS coordinates to a city name.
- Run `make vet` before committing to catch common Go mistakes.
- Use `make fmt` to auto-format all Go source files with `gofmt`.

---

## Known Limitations

- Forecast data is limited to what Open-Meteo exposes free of charge (no hourly history,
  no radar, no satellite imagery).
- Reverse geocoding uses [Nominatim](https://nominatim.openstreetmap.org/) (OpenStreetMap),
  which enforces a rate limit of 1 request/second. Repeated rapid geolocation lookups may
  be throttled.
- Weather alerts are rule-based (temperature, wind, and UV thresholds) and are not official
  government-issued alerts.
- The multi-model consensus fetches 4 separate API calls in parallel; on a slow connection
  the page load may be noticeably slower.

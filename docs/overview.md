# Overview

Weather App is a Go project with both CLI and web interfaces. It uses the free
Open-Meteo API and does not require an API key.

## Network Note

The app requires outbound internet access to:

- `geocoding-api.open-meteo.com`
- `api.open-meteo.com`

If you are behind a corporate proxy or restrictive firewall, requests may time
out. See [Operations and Reference](./operations.md) for proxy details.

## Features

### Weather Data

- Real-time conditions: temperature, humidity, pressure, wind, UV index
- 5-day forecast with precipitation probability
- Sun and moon details with sunrise/sunset arc and daylight hours
- Weather alerts for heat, frost, storm, high UV, and more

### Interfaces

- CLI tool with ANSI output, spinners, and box layouts
- Web server with weather-focused UI

### Internals

- Multi-model consensus using ECMWF, ICON, Meteo-France, and MET Norway
- 10-minute in-memory cache on the web server
- Responsive web layout
- Weather-matched quotes and feels-like advice

## Project Structure

```text
WeatherApp/
|-- weather/
|   |-- client.go
|   |-- alerts.go
|   |-- quotes.go
|   |-- uv.go
|   `-- consensus.go
|-- cmd/
|   `-- cli/
|       `-- main.go
|-- templates/
|   `-- index.html
|-- static/
|-- main.go
|-- Makefile
|-- run.sh
|-- .env.example
|-- go.mod
`-- go.sum
```

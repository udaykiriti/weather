# CLI Guide

## Usage

```bash
./weather-cli [city]
./weather-cli -city <city> [-units metric|imperial]
```

## Flags

| Flag | Default | Description |
| --- | --- | --- |
| `city` | `London` | City as a positional argument |
| `-city` | `London` | City name flag |
| `-units` | `metric` | `metric` (C/km/h) or `imperial` (F/mph) |

## Examples

```bash
./weather-cli London
./weather-cli "New York"
./weather-cli -city Tokyo
./weather-cli -city Mumbai -units metric
./weather-cli -city "New York" -units imperial
```

## Output Sections

- Animated spinner while fetching data
- Boxed header with city, country, and unit system
- Weather alerts with severity colors
- Current conditions including temperature, feels like, humidity, cloud cover,
  pressure, wind, and UV index
- Daylight arc with sunrise, sunset, and current sun position
- 5-day forecast table with temperature and precipitation details
- Multi-model consensus with per-model temperature bars

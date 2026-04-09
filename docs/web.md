# Web Guide

## Start the Server

```bash
make run-web
```

Or:

```bash
./run.sh run-web
./weather-web
```

Then open <http://localhost:8080>.

## Usage

- Enter a city name in the search box
- Choose Celsius or Fahrenheit
- View current conditions, alerts, quotes, UV index, sunrise/sunset arc,
  5-day forecast, and model consensus

## Custom Port

Run with:

```bash
PORT=3000 ./weather-web
```

Or add this to `.env`:

```env
PORT=3000
```

# Operations and Reference

## API Reference

This app uses [Open-Meteo](https://open-meteo.com/). No registration is needed.

| API | Description |
| --- | --- |
| Geocoding API | Resolves city name to coordinates and timezone |
| Forecast API (current) | Temperature, wind, humidity, UV, cloud cover |
| Forecast API (daily) | 5-day high/low, wind, precipitation probability |
| Forecast API (models) | ECMWF, ICON, Meteo-France, MET Norway consensus |

Weather conditions are decoded from
[WMO Weather Codes](https://open-meteo.com/en/docs#weathervariables).

## Environment Variables

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `PORT` | No | `8080` | Web server port |

## Network and Proxy

If the CLI or web app shows `dial tcp: lookup ... connection refused` or
`context deadline exceeded`, your system DNS resolver may be unreachable or
your network may block outbound HTTPS.

The app includes a built-in DNS fallback and retries against:

- `8.8.8.8`
- `1.1.1.1`

If you are behind an HTTP or HTTPS proxy, set:

```bash
export HTTPS_PROXY=http://proxy.example.com:8080
export HTTP_PROXY=http://proxy.example.com:8080
export NO_PROXY=localhost,127.0.0.1
make run-cli ARGS="London"
```

Go's `net/http` package respects `HTTPS_PROXY` and `HTTP_PROXY`
automatically.

## Tips

- Quote city names with spaces: `./weather-cli "New York"`
- You can also run `make run-cli ARGS="New York"`
- The web server caches results for 10 minutes per city
- The web UI uses `/api/reverse` as a server-side proxy for reverse geocoding
- Run `make vet` before committing
- Run `make fmt` to format Go files with `gofmt`

## Known Limitations

- Forecast data is limited to what Open-Meteo exposes for free
- Reverse geocoding uses Nominatim and may be throttled at 1 request per second
- Weather alerts are rule-based and not official government-issued alerts
- The multi-model consensus performs 4 parallel API calls and may be slower on
  slow networks

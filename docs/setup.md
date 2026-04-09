# Setup and Build

## Prerequisites

- Go 1.21 or higher
- No API key needed

## Initial Setup

1. Clone or download the repository.
2. Install dependencies:

```bash
go mod tidy
```

3. Build the project:

```bash
make build
```

Or:

```bash
./run.sh build
```

## Build Commands

### Using Make

```bash
make cli
make web
make build
make run-cli
make run-web
make fmt
make vet
make clean
```

Pass a city with `ARGS`:

```bash
make run-cli ARGS="Tokyo"
make run-cli ARGS="New York"
```

### Using `run.sh`

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

# kong-boilerplate

A small Go CLI starter using Kong for command parsing and slog + tint for structured logging.

## Features

- CLI entrypoint with a `hello` subcommand
- Structured logging with `slog` and `tint`
- Debug logging support via `--debug`
- Error handling that exits with a non-zero status

## Requirements

- Go 1.26+

## Getting Started

```bash
go mod download
```

## Usage

Run the CLI:

```bash
go run . --help
```

# AGENTS.md

This file provides guidance to AI coding agents (Claude Code, etc.) when working with code in this repository.

## Build & Development Commands

```bash
make build     # Build binary to ./bin/gsheet
make test      # Run tests
make vet       # Vet code
make fmt       # Format code
make lint      # Run golangci-lint (version managed by go.mod tool directive)
make tidy      # go mod tidy
make clean     # Remove build artifacts
make release type=patch dryrun=false  # Tag and push a release
```

The Makefile reads `.product_name` for the binary name and `go.mod` for the Go version. These are shared with the sibling `gml` project and should not contain app-specific hardcoding.

## Architecture

Google Sheets CLI built with Cobra/Viper, following the same architecture as the `gml` (Gmail CLI) project at `../gml`.

### Package Structure

- **`cmd/`** — Cobra commands. Each command creates a `gsheet.Service` via config, calls domain functions in `internal/gsheet/`, and formats output.
- **`internal/google/`** — Google API wrappers. `auth.go` defines the `Authenticator` interface with OAuth and Service Account implementations. `sheets.go` and `drive.go` wrap the respective API services. OAuth scopes are centralized in `Scopes()` in `auth.go`.
- **`internal/gsheet/`** — Domain logic. `service.go` composes `SheetsService` + `DriveService` into a single `Service`. `sheets.go` contains spreadsheet operations. `format.go` handles text (tablewriter) and JSON output. `config.go` manages TOML config via Viper.
- **`internal/version/`** — Version variables injected at build time via ldflags.

### Key Patterns

- **Authentication flow**: `config.go` → `service.go` calls `newAuthenticator()` → factory selects OAuth or Service Account → `Authenticator.GetClient()` returns `*http.Client` → passed to Google API service constructors.
- **Adding a new command**: Create `cmd/<name>.go`, call `rootCmd.AddCommand()` in `init()`, use `GetConfig()` → `gsheet.NewService()` → domain function → format function.
- **Adding a new OAuth scope**: Add to `Scopes()` in `internal/google/auth.go`. Users must re-run `gsheet auth`.

### Config

File: `~/.config/gsheet/config.toml`

```toml
auth_type = "oauth"
application_credentials = "/path/to/credentials.json"
user_credentials = "/path/to/token.json"
```

### Google APIs Used

- **Sheets API** (`sheets.SpreadsheetsScope`) — read/write spreadsheet data
- **Drive API** (`drive.DriveMetadataReadonlyScope`) — list user's spreadsheets

# gsheet

Google Sheets CLI client - A command-line tool for interacting with Google Spreadsheets.

## Installation

### Using Go

```bash
go install github.com/longkey1/gsheet@latest
```

### Using Homebrew (coming soon)

```bash
brew install longkey1/tap/gsheet
```

## Setup

### 1. Create Google Cloud Project and OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google Sheets API
4. Create OAuth 2.0 credentials (Desktop application type)
5. Download the credentials JSON file

### 2. Create Configuration File

Create `~/.config/gsheet/config.toml`:

```toml
auth_type = "oauth"
application_credentials = "/path/to/credentials.json"
user_credentials = "/path/to/token.json"
```

### 3. Authenticate

```bash
gsheet auth
```

This will open your browser for Google OAuth authentication.

## Usage

### Files - Spreadsheet File Operations

```bash
# List your spreadsheets
gsheet files list

# Search spreadsheets by name
gsheet files list -q "家計簿"

# Show only spreadsheets owned by me
gsheet files list --mine

# Output as JSON
gsheet files list --format json
```

### Sheets - Worksheet Tab Operations

```bash
# List all worksheet tabs in a spreadsheet
gsheet sheets list <spreadsheet-id>

# Search worksheet tabs by title
gsheet sheets list <spreadsheet-id> -q "2026-03"

# Search with regex
gsheet sheets list <spreadsheet-id> -q "2026-0[1-3]" --regex

# Output as JSON
gsheet sheets list <spreadsheet-id> --format json
```

### Cells - Cell Operations

```bash
# Search for cells by content
gsheet cells list <spreadsheet-id> --sheet "Sheet1" -q "検索文字列"

# Search with regex
gsheet cells list <spreadsheet-id> --sheet "Sheet1" -q "合計.*円" --regex

# Output as JSON
gsheet cells list <spreadsheet-id> --sheet "Sheet1" -q "検索文字列" --format json

# Update cells with inline values (rows separated by ";", cells by ",")
gsheet cells update <spreadsheet-id> --sheet "Sheet1" --range "C5" --values "15000"

# Update cells from stdin (CSV format)
cat data.csv | gsheet cells update <spreadsheet-id> --sheet "Sheet1" --range "A1:C2" --stdin
```

### Typical Workflow

```bash
# 1. Find the spreadsheet
gsheet files list -q "家計簿"
# → ID: abc123, NAME: "家計簿2026"

# 2. Find the worksheet tab
gsheet sheets list abc123 -q "2026-03"
# → TITLE: "2026-03月"

# 3. Find the cell coordinates
gsheet cells list abc123 --sheet "2026-03月" -q "25日"
# → CELL: A5
gsheet cells list abc123 --sheet "2026-03月" -q "売上"
# → CELL: C1

# 4. Update the cell at the intersection
gsheet cells update abc123 --sheet "2026-03月" --range "C5" --values "15000"
```

### Legacy Commands

The following commands are still available for backward compatibility:

```bash
# List spreadsheets (same as: gsheet files list)
gsheet list

# Get cell values
gsheet get <spreadsheet-id> --sheet "Sheet1" --range "A1:C10"

# Update cell values (same as: gsheet cells update)
gsheet update <spreadsheet-id> --sheet "Sheet1" --range "A1:C2" --values "a,b,c;d,e,f"

# Append rows
gsheet append <spreadsheet-id> --sheet "Sheet1" --values "a,b,c;d,e,f"

# Clear cell values
gsheet clear <spreadsheet-id> --sheet "Sheet1" --range "A1:C10"
```

### Version

```bash
gsheet version
```

## Configuration Options

| Option | Description |
|--------|-------------|
| `auth_type` | Authentication type: `oauth` or `service_account` |
| `application_credentials` | Path to OAuth client credentials JSON file |
| `user_credentials` | Path to store OAuth user token (for OAuth auth type) |

## License

Apache License 2.0

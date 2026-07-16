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
gsheet files list -q "budget"

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

# Add a new worksheet tab
gsheet sheets add <spreadsheet-id> --title "2026-04"

# Copy an existing worksheet tab
gsheet sheets add <spreadsheet-id> --title "2026-04" --from "2026-03"

# Rename a worksheet tab
gsheet sheets rename <spreadsheet-id> --sheet "2026-03" --title "March 2026"

# Insert 3 rows at row 5
gsheet sheets rows insert <spreadsheet-id> --sheet "Sheet1" --start 5 --count 3

# Insert rows inheriting formatting from the row before
gsheet sheets rows insert <spreadsheet-id> --sheet "Sheet1" --start 5 --after

# Delete rows 3 to 5
gsheet sheets rows delete <spreadsheet-id> --sheet "Sheet1" --start 3 --count 3

# Insert 2 columns at column 3
gsheet sheets cols insert <spreadsheet-id> --sheet "Sheet1" --start 3 --count 2

# Delete column 2
gsheet sheets cols delete <spreadsheet-id> --sheet "Sheet1" --start 2
```

### Cells - Cell Operations

```bash
# Get cell values
gsheet cells get <spreadsheet-id> --sheet "Sheet1" --range "A1:C10"

# Get cell values as CSV
gsheet cells get <spreadsheet-id> --sheet "Sheet1" --range "A1:C10" --format csv

# Search for cells by content
gsheet cells list <spreadsheet-id> --sheet "Sheet1" -q "keyword"

# Search with regex
gsheet cells list <spreadsheet-id> --sheet "Sheet1" -q "total.*usd" --regex

# Output as JSON
gsheet cells list <spreadsheet-id> --sheet "Sheet1" -q "keyword" --format json

# Update cells with inline values (rows separated by ";", cells by ",")
gsheet cells update <spreadsheet-id> --sheet "Sheet1" --range "C5" --values "15000"

# Update cells from a CSV file
gsheet cells update <spreadsheet-id> --sheet "Sheet1" --range "A1:C2" --file data.csv

# Write the entire stdin content into a single cell (preserves newlines, no CSV parsing)
echo "multi-line\nnote" | gsheet cells update <spreadsheet-id> --sheet "Sheet1" --range "B2" --stdin
cat note.md | gsheet cells update <spreadsheet-id> --sheet "Sheet1" --range "B2" --stdin

# Import CSV file into a worksheet
gsheet cells import <spreadsheet-id> --sheet "Sheet1" --file data.csv

# Import CSV from stdin
cat data.csv | gsheet cells import <spreadsheet-id> --sheet "Sheet1" --stdin

# Clear cell values
gsheet cells clear <spreadsheet-id> --sheet "Sheet1" --range "A1:C10"

# Export to CSV file
gsheet cells get <spreadsheet-id> --sheet "Sheet1" --format csv > data.csv
```

### Notes - Cell Note Operations

```bash
# Get the note on a cell
gsheet cells note get <spreadsheet-id> --sheet "Sheet1" --cell "B3"

# Set a note on a cell
gsheet cells note set <spreadsheet-id> --sheet "Sheet1" --cell "B3" --note "manually entered"

# Read note text from stdin
echo "supplemental comment" | gsheet cells note set <spreadsheet-id> --sheet "Sheet1" --cell "B3" --stdin
cat memo.txt | gsheet cells note set <spreadsheet-id> --sheet "Sheet1" --cell "B3" --stdin

# Clear a note (omit --note flag)
gsheet cells note set <spreadsheet-id> --sheet "Sheet1" --cell "B3"

# Clear a note (pass empty string)
gsheet cells note set <spreadsheet-id> --sheet "Sheet1" --cell "B3" --note ""
```

### Get - Quick Access via URL

Fetch cells directly from a Google Sheets URL. The spreadsheet ID, worksheet tab (`gid`), and range are all resolved from the URL, so no flags are required.

```bash
# Copy the URL from your browser (must be quoted — the shell would otherwise strip everything after #)
gsheet get 'https://docs.google.com/spreadsheets/d/abc123/edit#gid=123456&range=A1:C10'

# Without a range, the entire worksheet tab is fetched
gsheet get 'https://docs.google.com/spreadsheets/d/abc123/edit#gid=123456'

# Output as JSON or CSV
gsheet get 'https://docs.google.com/spreadsheets/d/abc123/edit#gid=123456&range=A1:C10' --format json
gsheet get 'https://docs.google.com/spreadsheets/d/abc123/edit#gid=123456&range=A1:C10' --format csv
```

### Typical Workflow

```bash
# 1. Find the spreadsheet
gsheet files list -q "budget"
# → ID: abc123, NAME: "Budget 2026"

# 2. Find the worksheet tab
gsheet sheets list abc123 -q "2026-03"
# → TITLE: "2026-03"

# 3. Find the cell coordinates
gsheet cells list abc123 --sheet "2026-03" -q "day 25"
# → CELL: A5
gsheet cells list abc123 --sheet "2026-03" -q "revenue"
# → CELL: C1

# 4. Update the cell at the intersection
gsheet cells update abc123 --sheet "2026-03" --range "C5" --values "15000"
```

### CSV Workflow Examples

```bash
# BigQuery → Spreadsheet
bq query --format=csv 'SELECT * FROM dataset.table' > data.csv
gsheet cells import <spreadsheet-id> --sheet "Sheet1" --file data.csv

# Spreadsheet → CSV → LLM / other tools
gsheet cells get <spreadsheet-id> --sheet "Sheet1" --format csv > data.csv
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
| `read_only` | When `true`, disables all write commands (add, update, clear, delete, etc.) |

Write commands can also be disabled via the `GSHEET_READ_ONLY` environment variable (e.g. `GSHEET_READ_ONLY=true`), which is useful when letting an LLM operate the CLI without config file changes.

## License

Apache License 2.0

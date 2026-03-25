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

### List Sheets

```bash
# List all sheets in a spreadsheet
gsheet list <spreadsheet-id>

# Output as JSON
gsheet list <spreadsheet-id> --format json
```

### Get Cell Values

```bash
# Get all values from Sheet1
gsheet get <spreadsheet-id>

# Get values from a specific sheet
gsheet get <spreadsheet-id> --sheet "Sheet1"

# Get values from a specific range
gsheet get <spreadsheet-id> --sheet "Sheet1" --range "A1:C10"

# Output as JSON
gsheet get <spreadsheet-id> --format json
```

### Update Cell Values

```bash
# Update cells with inline values (rows separated by ";", cells by ",")
gsheet update <spreadsheet-id> --range "A1:C2" --values "a,b,c;d,e,f"

# Update cells in a specific sheet
gsheet update <spreadsheet-id> --sheet "Sheet1" --range "A1:C2" --values "a,b,c;d,e,f"

# Update cells from stdin (CSV format)
cat data.csv | gsheet update <spreadsheet-id> --range "A1:C2" --stdin
```

### Append Rows

```bash
# Append rows with inline values
gsheet append <spreadsheet-id> --values "a,b,c;d,e,f"

# Append to a specific sheet
gsheet append <spreadsheet-id> --sheet "Sheet1" --values "a,b,c;d,e,f"

# Append from stdin (CSV format)
cat data.csv | gsheet append <spreadsheet-id> --stdin
```

### Clear Cell Values

```bash
# Clear a range
gsheet clear <spreadsheet-id> --range "A1:C10"

# Clear a range in a specific sheet
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

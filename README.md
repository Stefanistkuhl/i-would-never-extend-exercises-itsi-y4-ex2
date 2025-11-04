# pcapstore

CLI tool for managing pcap file storage and operations.

**Note: This is a personal quality-of-life tool, for a school exercise. It is likely unmaintained and provided as-is.**

## Client Configuration

Client configuration is stored in `~/.pcapstore` (server URL, password, port). Environment variables override config file:
- `PCAPSTORE_SERVER` - Server URL
- `PCAPSTORE_PASSWORD` - Authentication password
- `SORTER_PASSWORD` - Alternative password environment variable
- `PCAPSTORE_PORT` - Server port
- `PCAPSTORE_SOCKET` - Unix socket path

Command-line flags override environment variables and config file.

## Server Configuration

Server configuration is stored in `config.toml` and synchronized with the database. Configuration options:

- `watch_dir` - Directory to watch for incoming pcap files. Files placed here are automatically processed and organized.
- `organized_dir` - Directory where processed files are stored. Files are organized by hostname and datetime.
- `archive_dir` - Directory where archived files are moved. Files are archived after `archive_days` days.
- `expose_service` - Whether to expose the HTTP service (boolean).
- `port` - HTTP server port (default: 13173, must be between 1024-65536).
- `compression_enabled` - Whether automatic compression is enabled (boolean).
- `archive_days` - Number of days before files are automatically archived (default: 30).
- `max_retention_days` - Maximum retention period in days before files are deleted (default: 90).
- `log_level` - Logging level (e.g., "info", "debug", "error").

### Directory Workflow

1. Files are placed in `watch_dir` with naming format: `{hostname}_{scenario}_{YYYYMMDD_HHmmss}.pcap`
2. Files are validated, analyzed, and moved to `organized_dir` organized by hostname and datetime
3. After `archive_days`, files are moved from `organized_dir` to `archive_dir` maintaining the same structure
4. Files older than `max_retention_days` become cleanup candidates

Use `config get` to view current configuration and `config update` to modify it.

## Global Flags

- `--server, -s` - Server URL
- `--password, -p` - Authentication password
- `--port` - Server port
- `--socket` - Unix socket path
- `--raw` - Output raw JSON (no color, for piping to jq)

## Commands

### files

File operations.

- `files list` - List all capture files
- `files get <id>` - Get file details by ID
- `files download <id> [output]` - Download a file to specified path (or current directory)
- `files delete <id>` - Delete a file
- `files stats <id>` - Get statistics for a specific file
- `files by-hostname <hostname>` - List files filtered by hostname
- `files by-scenario <scenario>` - List files filtered by scenario

### stats

Statistics operations.

- `stats by-id <id>` - Get file statistics by ID
- `stats summary` - Get summary statistics across all files
- `stats by-hostname` - Get statistics grouped by hostname
- `stats by-scenario` - Get statistics grouped by scenario

### archive

Archive operations.

- `archive file <id>` - Archive a file
- `archive list` - List archived files
- `archive status` - Get archive status information

### compression

Compression operations.

- `compression file <id>` - Compress a file
- `compression trigger` - Trigger compression for pending files

### cleanup

Cleanup operations.

- `cleanup candidates` - Get cleanup candidates
- `cleanup execute` - Execute cleanup operations

### config

Configuration operations.

- `config get` - Get current server configuration
- `config update` - Update server configuration (opens in editor)

### search

- `search [query]` - Search for files (optional query string)

### export

- `export` - Export entire store (database and capture files) as tar.gz archive

### health

- `health` - Check server health status

### status

- `status` - Get server status information

### version

- `version` - Get server version information

### serve

- `serve` - Start the pcapstore server

## Examples

```bash
# List all files
pcapstore files list

# Get file details
pcapstore files get 1

# Download a file
pcapstore files download 1 /path/to/output.pcap

# Get statistics summary
pcapstore stats summary

# Archive a file
pcapstore archive file 1

# Search for files
pcapstore search "hostname1"

# Export store
pcapstore export > store_backup.tar.gz

# Use raw JSON output
pcapstore files list --raw | jq '.[] | .id'
```

## Unix Socket Support

When connecting via Unix socket (using `--socket` flag or `PCAPSTORE_SOCKET` env var), authentication is not required.

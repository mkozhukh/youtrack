# YouTrack MCP Server

MCP server and CLI for YouTrack. Gives AI assistants (Claude, etc.) direct access to your issues, tags, comments, worklogs, attachments, and links.

## Features

- Issue CRUD, search, and command execution
- Tags, comments, attachments, worklogs
- Issue linking (depends on, relates to, subtask, etc.)
- Project and user lookups
- CLI tool (`yt`) for terminal workflows
- STDIO (default) and Streaming HTTP modes

## Install

```bash
git clone <repository-url>
cd youtrack-mcp
go build -o youtrack-mcp ./cmd/youtrack-mcp
go build -o yt ./cmd/yt
```

## Quick Start

### MCP Server

Create a `config.toml`:

```toml
[youtrack]
base_url = "https://youtrack.example.com"
api_key = "perm:your-token-here"
default_project = "PROJ"
```

Run:

```bash
./youtrack-mcp
```

Or with environment variables:

```bash
export YOUTRACK_YOUTRACK_BASE_URL="https://youtrack.example.com"
export YOUTRACK_YOUTRACK_API_KEY="perm:your-token-here"
./youtrack-mcp
```

For HTTP mode: `./youtrack-mcp --http` (health check at `/health`).

### CLI

```bash
./yt login   # prompts for URL and token
./yt tickets list -p PROJ
./yt tickets show PROJ-123
./yt tickets create -p PROJ -t "Fix the login bug"
```

### Claude Desktop

```json
{
  "mcpServers": {
    "youtrack": {
      "command": "/path/to/youtrack-mcp",
      "env": {
        "YOUTRACK_YOUTRACK_BASE_URL": "https://youtrack.example.com",
        "YOUTRACK_YOUTRACK_API_KEY": "perm:your-token-here"
      }
    }
  }
}
```

## API Key

1. Log in to YouTrack
2. Profile > Authentication > New permanent token
3. Copy the token into your config

The token needs read/write access to issues, comments, tags, and projects.

## Documentation

- [MCP tools reference](spec/mcp.md) -- all available MCP tools and parameters
- [CLI commands](spec/yt.md) -- `yt` command-line interface
- [REST client API](spec/rest.md) -- Go client methods used internally
- [Usage scenarios](spec/usage.md) -- common workflow examples

## Built With

Go, [mcp-go](https://github.com/mark3labs/mcp-go), [cobra](https://github.com/spf13/cobra), [koanf](https://github.com/knadh/koanf), [charmbracelet/log](https://github.com/charmbracelet/log), [chi](https://github.com/go-chi/chi)

## License

MIT

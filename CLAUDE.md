# MCP server and CLI tools for youtrack

## Technical stack

- cobra for cli
- charmbracelet/log for logs
- chi for web routing
- koanf for configuration ( toml file )
- use mcp-go for mcp server

## Extra information

- [yt CLI specification](spec/yt.md) - CLI tool commands and options
- [MCP tools specification](spec/mcp.md) - MCP server tools reference
- [REST client API](spec/rest.md) - YouTrack REST client methods
- [Usage scenarios](spec/usage.md) - Common workflow examples

## Workflow

For each iterations

- plan what need to be done
- implement feature
- ensure that build passes
    - `go build -o yt ./cmd/yt`
    - `go build -o youtrack-mcp ./cmd/youtrack-mcp`
- run `go fmt ./...`

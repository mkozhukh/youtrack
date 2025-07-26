# MCP server and CLI tools for youtrack

## Technical stack

- cobra for cli
- charmbracelet/log for logs
- chi for web routing
- koanf for configuration ( toml file )
- use [mcp-go library](spec/go/mcp-go.md) for mcp server

## Extra information

[mcp tasks specification](spec/go/mcp-go.md)
[youtrack rest client](youtrack/README.md)
[yt cli tool](youtrack/README.md)

## Workflow

For each iterations

- plan what need to be done
- implement feature
- ensure that build passes
- run `go fmt`

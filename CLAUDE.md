# MCP server for youtrack

## Technical stack

use [mcp-go library](spec/go/mcp-go.md)

- cobra for cli
- charmbracelet/log for logs
- chi for web routing
- koanf for configuration ( toml file )

## Extra information

[mcp tasks specification](spec/go/mcp-go.md)
[youtrack rest client](youtrack/README.md)

## Workflow

For each iterations

- plan what need to be done
- implement feature
- ensure that build passes
- run `go fmt`
- make commit ( use dense commit message without mentioning claude )

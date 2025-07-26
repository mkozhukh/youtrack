# YouTrack MCP Server

A Model Context Protocol (MCP) server that provides YouTrack integration capabilities for AI assistants like Claude. This server allows AI assistants to interact with YouTrack's issue tracking system through a standardized interface.

## Features

- **Issue Management**: Create, read, update, and search issues with smart query defaults
- **Tag Management**: Add tags to issues with automatic tag creation
- **Comment Management**: Add comments to issues

## Supported modes

- STDIO - default
- Streaming HTTP - run like `youtrack-mcp --http`

## Available Tools

### Issue Operations

#### `get_issue_list`
Search for issues in a project with optional filtering.

**Parameters:**
- `project_id` (required): The YouTrack project ID (e.g., "PROJ")
- `query` (optional): YouTrack search query (e.g., "assignee: me")
- `max_results` (optional): Maximum number of issues to return (default: 10)

**Example:**
```json
{
  "name": "get_issue_list",
  "arguments": {
    "project_id": "PROJ",
    "query": "state: Open assignee: john.doe",
    "max_results": 20
  }
}
```

#### `get_issue_details`
Get detailed information about a specific issue including comments.

**Parameters:**
- `issue_id` (required): The issue ID (e.g., "PROJ-123")

**Example:**
```json
{
  "name": "get_issue_details",
  "arguments": {
    "issue_id": "PROJ-123"
  }
}
```

#### `create_issue`
Create a new issue in a project.

**Parameters:**
- `project_id` (required): The YouTrack project ID
- `summary` (required): Issue summary/title
- `description` (optional): Issue description

**Example:**
```json
{
  "name": "create_issue",
  "arguments": {
    "project_id": "PROJ",
    "summary": "Fix authentication bug",
    "description": "Users are unable to log in with SSO credentials"
  }
}
```

#### `update_issue`
Update an existing issue.

**Parameters:**
- `issue_id` (required): The issue ID
- `summary` (optional): New issue summary
- `description` (optional): New issue description
- `state` (optional): New issue state (e.g., "Open", "In Progress", "Fixed")
- `assignee` (optional): New assignee username (supports fuzzy matching)

**Example:**
```json
{
  "name": "update_issue",
  "arguments": {
    "issue_id": "PROJ-123",
    "state": "In Progress",
    "assignee": "john.doe"
  }
}
```

### Tag Operations

#### `tag_issue`
Add a tag to an issue. Creates the tag if it doesn't exist.

**Parameters:**
- `issue_id` (required): The issue ID
- `tag` (required): Tag name

**Example:**
```json
{
  "name": "tag_issue",
  "arguments": {
    "issue_id": "PROJ-123",
    "tag": "urgent"
  }
}
```

### Comment Operations

#### `add_comment`
Add a comment to an issue.

**Parameters:**
- `issue_id` (required): The issue ID
- `comment` (required): Comment text

**Example:**
```json
{
  "name": "add_comment",
  "arguments": {
    "issue_id": "PROJ-123",
    "comment": "I've reproduced this issue and will work on a fix."
  }
}
```

## Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd youtrack-mcp
   ```

2. **Build the server:**
   ```bash
   go build -o youtrack-mcp
   ```

3. **Configure the server** (see Configuration section below)

4. **Run the server:**
   ```bash
   ./youtrack-mcp
   ```

## Configuration

The server supports multiple configuration methods:

### 1. Configuration File

Create a `config.toml` file in the project root:

```toml
[server]
# MCP server configuration
port = 3204
name = "YouTrack MCP Server"

[logging]
# Enable tool call logging
log_tool_calls = true
tool_calls_file = "tool_calls.log"

[youtrack]
# YouTrack instance base URL
base_url = "https://youtrack.example.com"

# YouTrack API key (can be set via environment variable)
api_key = "your_api_key_here"

# Default project ID for operations
default_project = "PROJ"

# API timeout in seconds
timeout = 30

# Default query for issue listing
default_query = "updated: -7d"

# Default max results for issue listing
max_results = 10
```

### 2. Environment Variables

All configuration options can be set via environment variables with the `YOUTRACK_` prefix:

```bash
export YOUTRACK_YOUTRACK_BASE_URL="https://youtrack.example.com"
export YOUTRACK_YOUTRACK_API_KEY="your_api_key_here"
export YOUTRACK_YOUTRACK_DEFAULT_PROJECT="PROJ"
export YOUTRACK_LOGGING_LOG_TOOL_CALLS="true"
```

### 3. Custom Configuration File Path

Set a custom configuration file path:

```bash
export YOUTRACK_CONFIG_PATH="/path/to/your/config.toml"
./youtrack-mcp
```

## YouTrack API Key Setup

1. **Log in to your YouTrack instance**
2. **Go to your profile settings**
3. **Navigate to "Authentication"**
4. **Create a new permanent token**
5. **Copy the token and use it as your API key**

## Required YouTrack Permissions

The API key must have the following permissions:

- **Read Issues**: To retrieve issue details and search issues
- **Update Issues**: To modify issue properties and assignees
- **Create Issues**: To create new issues
- **Read/Write Comments**: To retrieve and add comments
- **Read/Write Tags**: To manage issue tags
- **Read Projects**: To validate project access
- **Read Users**: To resolve user assignments

## Error Handling

The server provides comprehensive error handling:

- **Parameter Validation**: Ensures all required parameters are provided and valid
- **API Error Translation**: Converts YouTrack API errors into user-friendly messages
- **Authentication Errors**: Clear messages for invalid API keys or permissions
- **Rate Limiting**: Handles YouTrack rate limits gracefully
- **Server Errors**: Appropriate handling of temporary server issues

## Logging

### Standard Logging

The server logs all operations with structured fields:

```
INFO Tool call executed tool=get_issue_list args=map[project_id:PROJ max_results:10] timestamp=2023-12-01 10:30:00
```

### Tool Call Logging

When enabled, detailed tool calls are logged to a file in JSON format:

```json
{"timestamp":"2023-12-01 10:30:00","tool":"get_issue_list","args":{"project_id":"PROJ","max_results":10}}
```

## Usage with AI Assistants

### With Claude Desktop

Add the server to your Claude Desktop configuration:

```json
{
  "mcpServers": {
    "youtrack": {
      "command": "/path/to/youtrack-mcp",
      "env": {
        "YOUTRACK_YOUTRACK_BASE_URL": "https://youtrack.example.com",
        "YOUTRACK_YOUTRACK_API_KEY": "your_api_key_here"
      }
    }
  }
}
```

### Example Queries

Once configured, you can ask your AI assistant:

- "Show me all open issues in the PROJ project"
- "Create a new issue for the login bug I found"
- "Add a comment to PROJ-123 saying I'm working on it"
- "Update PROJ-456 to assign it to john.doe and mark it as In Progress"
- "Tag PROJ-789 with 'urgent' and 'needs-review'"
- "Check the health status of the YouTrack server"

## Troubleshooting

### Common Issues

1. **Authentication Failed**
   - Verify your API key is correct
   - Check that the YouTrack URL is accessible
   - Ensure the API key has sufficient permissions

2. **Project Not Found**
   - Verify the project ID exists in YouTrack
   - Check that your API key has access to the project

3. **Connection Issues**
   - Verify the YouTrack URL is correct and accessible
   - Check firewall settings
   - Ensure YouTrack is running and responsive

4. **Permission Denied**
   - Verify your API key has the required permissions
   - Check YouTrack user roles and permissions

### Debug Mode

Enable debug logging by setting the log level:

```bash
export YOUTRACK_LOG_LEVEL="debug"
./youtrack-mcp
```

### Testing Connection

Test your configuration by running:

```bash
./youtrack-mcp
```

The server will attempt to connect to YouTrack during startup and report any issues.

### Health Check

In http mode, you can check the `/health` url to get the server status


## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:

1. Check the troubleshooting section above
2. Review YouTrack API documentation
3. Open an issue in the repository
4. Check YouTrack server logs for API errors

## Technical Details

- **Built with**: Go 1.19+
- **MCP Library**: [mcp-go](https://github.com/mark3labs/mcp-go)
- **Configuration**: [koanf](https://github.com/knadh/koanf)
- **Logging**: [charmbracelet/log](https://github.com/charmbracelet/log)
- **HTTP Routing**: [chi](https://github.com/go-chi/chi)
- **CLI Framework**: [cobra](https://github.com/spf13/cobra)

## Version

Current version: 1.0.0

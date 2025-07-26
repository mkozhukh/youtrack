# yt - YouTrack CLI Specification

`yt` is a command-line interface (CLI) tool for interacting with a remote YouTrack instance. It allows users to perform common YouTrack operations directly from their terminal.

## 1. Configuration

The `yt` tool loads its configuration from the following sources, in order of precedence (highest to lowest):

1.  **Command-line arguments:** Parameters passed directly with a command (e.g., `--token`).
2.  **Environment variables:** Variables prefixed with `YT_` (e.g., `YT_TOKEN`).
3.  **Configuration file:** A `config.toml` file.

### 1.1. Configuration File

The tool reads a TOML configuration file.

-   **Default Location:** `~/.config/yt/config.toml`
-   **Custom Location:** A path can be specified using the global `--config` or `-c` flag.

**Example `config.toml`:**

```toml
[server]
url = "https://youtrack.example.com"
token = "your-permanent-token"

[defaults]
project = "DEFAULT_PROJECT_ID"
user_id = "your-user-id" # Optional: Used as the default for --user flags
```

### 1.2. Configuration Parameters

-   `url`: The base URL of the YouTrack instance.
    -   CLI: `--url <URL>`
    -   Env: `YT_URL`
    -   File: `server.url`
-   `token`: The permanent token for API authentication.
    -   CLI: `--token <TOKEN>`
    -   Env: `YT_TOKEN`
    -   File: `server.token`
-   `user_id`: The YouTrack ID of the current user.
    -   Env: `YT_USER_ID`
    -   File: `defaults.user_id`

## 2. Commands

### Global Options

-   `--config <PATH>`, `-c <PATH>`: Path to the configuration file.
-   `--output <FORMAT>`, `-o <FORMAT>`: Output format (e.g., `text`, `json`). Default: `text`.
-   `--verbose`, `-v`: Enable verbose output (changes log level to INFO, default is Warn).
-   `--help`, `-h`: Show help message.

### `yt login`

Interactively prompts the user for the YouTrack URL and a permanent token, then saves them to the configuration file. It will also attempt to automatically determine and save the user's own YouTrack user ID, which enables commands to default to the current user.

### `yt completion <shell>`

Generates a shell completion script for the specified shell (e.g., `bash`, `zsh`).

### `yt projects`

Manages projects.

#### `yt projects list`

Shows a list of all available projects.

-   **Alias:** `yt projects`
-   **Options:**
    -   `--query <QUERY>`, `-q <QUERY>`: Filter projects by a search query.

#### `yt projects describe <project_id>`

Shows detailed information for a specific project, including available custom fields, statuses, and types.

-   **Arguments:**
    -   `<project_id>`: The ID of the project to describe (e.g., "PRJ"). (Required)

### `yt tickets`

Manages tickets (issues).

#### `yt tickets list`

Shows the latest tickets in a project.

-   **Alias:** `yt tickets`
-   **Options:**
    -   `--project <PROJECT_ID>`, `-p <PROJECT_ID>`: The project ID. If not provided, uses the default project from the config.
    -   `--limit <NUMBER>`: Number of tickets to show. Default: 20.
    -   `--query <QUERY>`, `-q <QUERY>`: Filter tickets with a YouTrack search query.
    -   `--user <USER>`, `-u <USER>`: Filter tickets by assignee. If not provided, defaults to the current user's ID stored in the config.

#### `yt tickets show <ticket_id>`

Shows detailed information for a specific ticket.

-   **Arguments:**
    -   `<ticket_id>`: The full ID of the ticket (e.g., "PRJ-123"). (Required)

#### `yt tickets create`

Creates a new ticket in a project.

-   **Options:**
    -   `--project <PROJECT_ID>`, `-p <PROJECT_ID>`: The project ID. If not provided, uses the default project from the config. (Required)
    -   `--title <TITLE>`, `-t <TITLE>`: The title of the new ticket. (Required)
    -   `--description <DESC>`, `-d <DESC>`: The description for the ticket.
    -   `--assignee <USER>`: Assign the ticket to a user.
    -   `--field "<KEY>=<VALUE>"`: Set a custom field. Can be specified multiple times.

#### `yt tickets update <ticket_id>`

Updates fields of a specific ticket.

-   **Arguments:**
    -   `<ticket_id>`: The full ID of the ticket to update. (Required)
-   **Options:**
    -   `--status <STATUS>`: Change the ticket's status.
    -   `--assignee <USER>`: Change the assignee.
    -   `--field "<KEY>=<VALUE>"`: Set a custom field. Can be specified multiple times.

#### `yt tickets tag <ticket_id> <tag_name...>`

Adds one or more tags to a ticket.

-   **Arguments:**
    -   `<ticket_id>`: The full ID of the ticket. (Required)
    -   `<tag_name...>`: One or more tag names to add. (Required)

#### `yt tickets untag <ticket_id> <tag_name...>`

Removes one or more tags from a ticket.

-   **Arguments:**
    -   `<ticket_id>`: The full ID of the ticket. (Required)
    -   `<tag_name...>`: One or more tag names to remove. (Required)

### `yt tickets comments`

Manages comments on a ticket.

#### `yt tickets comments list <ticket_id>`

Lists all comments for a specific ticket.

-   **Arguments:**
    -   `<ticket_id>`: The full ID of the ticket. (Required)

#### `yt tickets comments add <ticket_id>`

Adds a comment to a ticket.

-   **Arguments:**
    -   `<ticket_id>`: The full ID of the ticket. (Required)
-   **Options:**
    -   `--message <MESSAGE>`, `-m <MESSAGE>`: The comment message. (Required)

### `yt tickets attachments`

Manages attachments on a ticket.

#### `yt tickets attachments list <ticket_id>`

Lists all attachments for a specific ticket.

-   **Arguments:**
    -   `<ticket_id>`: The full ID of the ticket. (Required)

#### `yt tickets attachments add <ticket_id> <file_path>`

Attaches a file to a ticket.

-   **Arguments:**
    -   `<ticket_id>`: The full ID of the ticket. (Required)
    -   `<file_path>`: The path to the file to attach. (Required)

### `yt tickets worklogs`

Manages worklogs on a ticket.

#### `yt tickets worklogs list <ticket_id>`

Lists all worklog entries for a specific ticket.

-   **Arguments:**
    -   `<ticket_id>`: The full ID of the ticket. (Required)

#### `yt tickets worklogs add <ticket_id>`

Adds a worklog entry to a ticket.

-   **Arguments:**
    -   `<ticket_id>`: The full ID of the ticket. (Required)
-   **Options:**
    -   `--duration <DURATION>`: The duration of the work (e.g., "1h 30m"). (Required)
    -   `--description <DESC>`: An optional description for the worklog entry.

### `yt tickets links`

Manages links between tickets.

#### `yt tickets links add <ticket_id> <other_ticket_id>`

Links two tickets together.

-   **Arguments:**
    -   `<ticket_id>`: The full ID of the source ticket. (Required)
    -   `<other_ticket_id>`: The full ID of the target ticket. (Required)
-   **Options:**
    -   `--relation <RELATION>`, `-r <RELATION>`: The relationship type (e.g., "relates to", "is duplicated by"). (Required)

### `yt tickets history`

Shows the activity stream for a ticket.

#### `yt tickets history <ticket_id>`

Shows the activity stream for a ticket (field changes, comments, etc.).

-   **Arguments:**
    -   `<ticket_id>`: The full ID of the ticket. (Required)

### `yt users`

Manages users.

#### `yt users list`

Shows all users associated with a project (the project team).

-   **Alias:** `yt users`
-   **Options:**
    -   `--project <PROJECT_ID>`, `-p <PROJECT_ID>`: The project ID. If not provided, uses the default project from the config. (Required)

#### `yt users worklogs <user>`

Lists all worklog entries for a specific user.

-   **Arguments:**
    -   `<user>`: The username to show worklogs for. (Required)
-   **Options:**
    -   `--project <PROJECT_ID>`, `-p <PROJECT_ID>`: Filter worklogs by project.
    -   `--since <DATE>`: Show worklogs since a specific date (e.g., "2025-07-01").
    -   `--until <DATE>`: Show worklogs until a specific date.

## 3. Implementation Details

### 3.1. Authentication
- The `yt login` command handles authentication setup by prompting for YouTrack URL and permanent token
- During login, the tool verifies the connection and automatically determines the user's ID
- No separate verification command is needed

### 3.2. Error Handling
- Log errors and exit on failure
- For user input errors: provide readable error messages to help users fix the issue
- For network/server errors: display the raw error without custom messages
- Server error responses are logged at INFO level (visible with `--verbose`)

### 3.3. Output Formats
- **Text format**: Use custom data structures for formatting human-readable output
- **JSON format**: If custom structures are used for text output, dump those structures as JSON; otherwise, use raw API responses

### 3.4. Field Handling
- Custom fields use simple key=value format
- More complex field types will be addressed as practical use cases arise

### 3.5. User Identification
- Accept username or email for user-related parameters
- Implement partial matching for user lookups

### 3.6. Ticket ID Validation
- Validate ticket ID format (e.g., "PRJ-123") before making API calls

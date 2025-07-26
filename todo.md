# YouTrack CLI Implementation Plan

Full specification is at the spec/yt.md
Existing REST client spec is at the spec/rest.md

## Overview

This document outlines the implementation plan for the `yt` CLI tool that interfaces with YouTrack's REST API.

## Architecture

```
youtrack/
├── cmd/yt/                    # CLI entry point
│   └── main.go               # Main executable
├── internal/yt/              # CLI implementation
│   ├── config/              # Configuration management
│   │   └── config.go        # Config loading with koanf
│   ├── commands/            # Command implementations
│   │   ├── root.go         # Root command setup
│   │   ├── login.go        # Login command
│   │   ├── completion.go   # Shell completion
│   │   ├── projects.go     # Projects subcommands
│   │   ├── tickets/        # Tickets subcommands
│   │   │   ├── tickets.go  # Main tickets commands
│   │   │   ├── comments.go # Comments management
│   │   │   ├── attachments.go # Attachments management
│   │   │   ├── worklogs.go # Worklogs management
│   │   │   ├── links.go    # Links management
│   │   │   └── history.go  # History/activity stream
│   │   └── users.go        # Users subcommands
│   └── output/              # Output formatting
│       └── formatter.go     # Text/JSON formatters
└── pkg/youtrack/           # REST API client
    ├── client.go           # HTTP client
    ├── types.go            # Data structures
    ├── projects.go         # Projects API
    ├── issues.go           # Issues/Tickets API
    ├── comments.go         # Comments API
    ├── attachments.go      # Attachments API
    ├── worklogs.go         # Worklogs API
    ├── users.go            # Users API
    └── tags.go             # Tags API
```

## Implementation Steps

### Phase 1: Foundation (High Priority)

#### 1. Directory Structure **done**
**Expected Result:** Basic project structure created with all necessary directories.

#### 2. Configuration Management **done**
**Implementation:** 
- Use koanf for configuration loading
- Support TOML files, environment variables, and CLI flags
- Default config path: `~/.config/yt/config.toml`
- Environment variables prefixed with `YT_`
- Config includes: server.url, server.token, defaults.project, defaults.user_id

**Expected Result:** 
- Config struct defined with server URL, token, and defaults
- Config loading from file, env, and CLI with proper precedence
- Global `--config` flag working

#### 3. Login Command **done**
**Command:** `yt login`
**Implementation:**
- Interactive prompts for YouTrack URL and permanent token
- Verify connection and authenticate
- Automatically determine and save user's ID
- Save credentials to config file

**Expected Result:**
- User can configure authentication interactively
- Connection is verified during login
- User ID is auto-detected and saved

### Phase 2: Core Commands (Medium Priority)

#### 4. Projects Commands **done**
**Commands:** `yt projects list`, `yt projects describe <id>`
**Expected Result:**
- List all projects with optional query filter
- Show detailed project information including custom fields, statuses, and types
- Proper error handling for missing projects
- Support alias `yt projects` for list command

#### 5. Tickets List & Show **done**
**Commands:** `yt tickets list`, `yt tickets show <id>`
**Expected Result:**
- List tickets with project filter, limit, and query
- Support `--user` filter (defaults to current user from config)
- Show detailed ticket information
- Default to configured project when not specified
- Support alias `yt tickets` for list command
- Validate ticket ID format (e.g., "PRJ-123")

#### 6. Tickets Create **done**
**Command:** `yt tickets create`
**Expected Result:**
- Create tickets with title, description, assignee
- Support custom fields via `--field` flag
- Return created ticket ID

#### 7. Tickets Update **done**
**Command:** `yt tickets update <id>`
**Expected Result:**
- Update status, assignee, and custom fields
- Partial updates (only specified fields)
- Confirmation of changes

#### 8. Tickets Tag Management **done**
**Commands:** `yt tickets tag <id> <tags...>`, `yt tickets untag <id> <tags...>`
**Expected Result:**
- Add multiple tags in one command
- Remove specific tags
- Handle non-existent tags gracefully

#### 9. Users Commands **done**
**Commands:** `yt users list`, `yt users worklogs <user>`
**Expected Result:**
- List project team members
- Show worklogs with date filtering
- Support project filter for worklogs
- Support alias `yt users` for list command
- Accept username or email for user identification
- Implement partial matching for user lookups

### Phase 3: Advanced Commands (Low Priority)

#### 10. Tickets Comments **done**
**Commands:** `yt tickets comments list <id>`, `yt tickets comments add <id>`
**Expected Result:**
- List all comments for a ticket
- Add new comment with `--message` flag

#### 11. Tickets Attachments **done**
**Commands:** `yt tickets attachments list <id>`, `yt tickets attachments add <id> <file>`
**Expected Result:**
- List all attachments for a ticket
- Upload file as attachment

#### 12. Tickets Worklogs **done**
**Commands:** `yt tickets worklogs list <id>`, `yt tickets worklogs add <id>`
**Expected Result:**
- List worklog entries for a ticket
- Add worklog with duration and optional description

#### 13. Tickets Links **done**
**Command:** `yt tickets links add <id> <other_id>`
**Expected Result:**
- Link two tickets with specified relation type

#### 14. Tickets History **done**
**Command:** `yt tickets history <id>`
**Expected Result:**
- Show activity stream (field changes, comments, etc.)

#### 15. Shell Completion **done**
**Command:** `yt completion <shell>`
**Expected Result:**
- Generate completion scripts for bash/zsh

## Testing Strategy

- do not implement tests

## Output format

For all commands be sure to support the global `--output` flag, so each command need to provide text ( human redable ) and json ( just serialization of result ) outputs

Prefer readable output with consistent light formatting and color usage

## Error Handling

- Log errors and exit on failure
- For user input errors: provide readable error messages to help users fix the issue
- For network/server errors: display the raw error without custom messages
- Server error responses are logged at INFO level (visible with `--verbose`)

## Logging

do not overuse logging, user expect to receive the command output, not the process log

- log all errors with log.Error 
- if there is some minor error, which doesn't prevent normal command output - use log.Warn
- add log.Info for all rest API calls and server error responses

set default log level to Warn

## Implementation Notes

### User Identification
- Accept username or email for user-related parameters
- Implement partial matching for user lookups

### Ticket ID Validation
- Validate ticket ID format (e.g., "PRJ-123") before making API calls

### Field Handling
- Custom fields use simple key=value format
- More complex field types will be addressed as practical use cases arise
# YouTrack MCP Implementation Plan

## Overview
Implementation of MCP operations for YouTrack integration, structured in modular phases. Each phase builds upon the previous one with clear deliverables and acceptance criteria.

## Technical Architecture

### Module Structure
```
mcp/
├── handlers/          # MCP tool handlers
│   ├── get_issue_list.go      # separate file for each tool, contains tool description and handler
│   ├── get_issue_details.go   # same for other tool
│   └── ...
└── main.go            # MCP server initialization
```

### Configuration Extensions
- YouTrack API URL
- YouTrack API key
- Default project ID
- Default query parameters
- Timeout settings

---

## Phase 1: Foundation & Configuration
**Goal**: Establish core infrastructure and YouTrack client integration

### Tasks
1. **Create mcp package structure**
   - Create subdirectory: `handlers/`
   - Create `mcp/server.go` for MCP server initialization

2. **Extend configuration**
   - Add YouTrack section to `Config` struct
   - Add required fields: `base_url`, `api_key`, `default_project`, `timeout`
   - Update `config.toml` with YouTrack configuration section
   - Update environment variable handling

3. **Create YouTrack client wrapper**
   - Initialize YouTrack client from configuration
   - Create context wrapper for API calls
   - Add error handling and validation

### Acceptance Criteria
- [ ] MCP package structure exists
- [ ] Configuration loads YouTrack settings
- [ ] YouTrack client initializes successfully
- [ ] Basic error handling implemented
- [ ] Code compiles and runs without errors

---

## Phase 2: Issue Operations (Core)
**Goal**: Implement primary issue management operations

### Tasks
1. **get_issue_list tool**
   - Create `mcp/tools/issue_tools.go` with tool definition
   - Create `mcp/handlers/issues.go` with handler implementation
   - Parameters: `project_id`, `query` (optional), `max_results` (optional)
   - Use `client.SearchIssues()` for implementation
   - Handle pagination if needed

2. **get_issue_details tool**
   - Add tool definition to `issue_tools.go`
   - Add handler to `issues.go`
   - Parameters: `issue_id`
   - Use `client.GetIssue()` and `client.GetIssueComments()`
   - Format response with issue details and comments

3. **create_issue tool**
   - Add tool definition to `issue_tools.go`
   - Add handler to `issues.go`
   - Parameters: `project_id`, `summary`, `description` (optional)
   - Use `client.CreateIssue()` for implementation
   - Return created issue details

4. **update_issue tool**
   - Add tool definition to `issue_tools.go`
   - Add handler to `issues.go`
   - Parameters: `issue_id`, `state` (optional), `assignee` (optional)
   - Use `client.UpdateIssue()` and `client.UpdateIssueAssigneeByProject()` for implementation
   - Handle state and assignee updates separately

### Acceptance Criteria
- [ ] All 4 issue tools are defined and registered
- [ ] Handlers process parameters correctly
- [ ] YouTrack API calls work with real data
- [ ] Error handling covers common scenarios
- [ ] Tools return properly formatted responses
- [ ] Integration tests pass with real YouTrack instance

---

## Phase 3: Tag Operations
**Goal**: Implement tag management functionality

### Tasks
1. **tag_issue tool**
   - Create `mcp/tools/tag_tools.go` with tool definition
   - Create `mcp/handlers/tags.go` with handler implementation
   - Parameters: `issue_id`, `tag`
   - Use `client.EnsureTag()` to create tag if needed
   - Use `client.AddIssueTag()` to apply tag to issue

### Acceptance Criteria
- [ ] Tag tool is defined and registered
- [ ] Handler creates tags if they don't exist
- [ ] Tags are successfully applied to issues
- [ ] Error handling covers tag creation failures
- [ ] Integration tests pass

---

## Phase 4: Comment Operations
**Goal**: Implement comment management functionality

### Tasks
1. **add_comment tool**
   - Create `mcp/tools/comment_tools.go` with tool definition
   - Create `mcp/handlers/comments.go` with handler implementation
   - Parameters: `issue_id`, `comment`
   - Use `client.AddIssueComment()` for implementation
   - Return created comment details

### Acceptance Criteria
- [ ] Comment tool is defined and registered
- [ ] Handler adds comments successfully
- [ ] Error handling covers comment creation failures
- [ ] Integration tests pass

---

## Phase 5: Integration & Testing
**Goal**: Complete integration and comprehensive testing

### Tasks
1. **Main.go integration**
   - Remove example hello_world tool
   - Register all YouTrack tools

2. **Error handling enhancement**
   - Implement consistent error formatting
   - Add validation for required parameters
   - Handle YouTrack API errors gracefully

3. **Tool call logging**
   - Integrate tool call logging with existing logging system
   - Add structured logging for all operations

4. **Documentation update**
   - Update README with tool usage examples
   - Add configuration examples
   - Document required YouTrack permissions

### Acceptance Criteria
- [ ] All tools are registered and working
- [ ] Error handling is consistent across all tools
- [ ] Tool call logging works correctly
- [ ] Documentation is complete and accurate
- [ ] Manual testing passes for all operations

---

## Phase 6: Advanced Features & Optimization
**Goal**: Add advanced features and optimize performance

### Tasks
1. **Query optimization**
   - Implement smart defaults for issue queries

2. **Response formatting**
   - Implement consistent response formatting
   - Include relevant metadata in responses

3. **Configuration validation**
   - Add startup validation for YouTrack connectivity

4. **Performance monitoring**
   - Add health check endpoint

### Acceptance Criteria
- [ ] Responses are well-formatted and readable
- [ ] Configuration validation prevents startup issues
- [ ] System is production-ready

---

## Implementation Notes

### Configuration Example
```toml
[youtrack]
# YouTrack instance base URL
base_url = "https://youtrack.example.com"

# YouTrack API key (can be set via YOUTRACK_API_KEY env var)
api_key = "your_api_key_here"

# Default project ID for operations
default_project = "PROJ"

# API timeout in seconds
timeout = 30

# Default query for issue listing
default_query = "updated: -7d"

# Default max results for issue listing
default_max_results = 10
```

### Error Handling Strategy
- Validate all input parameters
- Handle YouTrack API errors gracefully
- Provide meaningful error messages
- Log errors for debugging
- Return consistent error format

### Testing Strategy
- Skip tests for now

---

## Success Metrics
- All 6 MCP tools implemented and working
- Code follows Go best practices and project conventions
- Documentation is complete and helpful
- Error handling is robust and user-friendly

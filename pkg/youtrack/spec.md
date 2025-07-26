# YouTrack API Client Specification

This document provides implementation details for LLM agents extending the YouTrack API client submodule.

## Architecture Overview

The YouTrack API client is designed as a stateless HTTP client where authentication context is passed with each request. This design ensures thread-safety and allows for multiple concurrent API operations with different credentials.

## File Structure

```
youtrack/
├── context.go    # Authentication context management
├── client.go     # Core HTTP client with request/response handling
├── errors.go     # Error types and handling
├── types.go      # Data structures for API entities
├── issues.go     # Issue-related API operations
├── projects.go   # Project-related API operations
└── comments.go   # Comment-related API operations
└── tags.go   # Tags-related API operations
└── users.go   # Users-related API operations
```

## Implementation Details

### Context (context.go)
- `YouTrackContext`: Wraps standard Go context with API key
- Stateless design - new context created for each operation
- Allows cancellation and timeout propagation

### Client (client.go)
- `Client`: Main HTTP client struct
- `doRequest`: Central request handler with:
  - Bearer token authentication
  - JSON marshaling/unmarshaling
  - Error response handling
  - Proper header management
- HTTP methods: `Get`, `Post`, `Put`, `Delete`

### Error Handling (errors.go)
- `APIError`: Custom error type for API errors
- Includes HTTP status code and error message
- Implements standard error interface

### Types (types.go)
- Core entities: `Issue`, `User`, `Project`, `IssueComment`
- Request/Response DTOs for API operations
- Uses pointer fields for optional values
- Time fields use `time.Time` type

## API Design Patterns

### Stateless Operations
Every API method follows this pattern:
```go
func (c *Client) MethodName(ctx *YouTrackContext, params...) (returnType, error)
```

### Field Selection
Most GET operations include field selection via query parameters:
```go
query.Add("fields", "id,summary,description,...")
```

### Pagination
List operations support pagination:
```go
query.Add("$skip", fmt.Sprintf("%d", skip))
query.Add("$top", fmt.Sprintf("%d", top))
```

## Extension Guidelines

### Adding New Endpoints

1. **Create new file** for logical grouping (e.g., `attachments.go`)

2. **Define types** in `types.go`:
```go
type Attachment struct {
    ID       string    `json:"id"`
    Name     string    `json:"name"`
    Size     int64     `json:"size"`
    MimeType string    `json:"mimeType"`
    Created  time.Time `json:"created"`
}
```

3. **Implement methods** following existing patterns:
```go
func (c *Client) GetAttachment(ctx *YouTrackContext, issueID, attachmentID string) (*Attachment, error) {
    path := fmt.Sprintf("/api/issues/%s/attachments/%s", issueID, attachmentID)

    resp, err := c.Get(ctx, path, nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var attachment Attachment
    if err := json.NewDecoder(resp.Body).Decode(&attachment); err != nil {
        return nil, fmt.Errorf("failed to decode attachment: %w", err)
    }

    return &attachment, nil
}
```

### Adding Custom Fields Support

For complex custom fields, extend the `Fields` map in types:
```go
// In types.go
type CustomField struct {
    Name  string      `json:"name"`
    Value interface{} `json:"value"`
}

// Usage in API methods
fields := []CustomField{
    {Name: "Priority", Value: "Critical"},
    {Name: "Assignee", Value: map[string]string{"login": "john.doe"}},
}
```

### Error Handling Extensions

For specific error types:
```go
type ValidationError struct {
    APIError
    Fields map[string]string `json:"error_fields"`
}

// In doRequest method
if resp.StatusCode == 400 {
    var valErr ValidationError
    // ... decode and return
}
```

## Testing Considerations

When adding new functionality:
1. Mock HTTP responses for unit tests
2. Use test context with test API keys
3. Validate request construction (headers, body, query params)
4. Test error scenarios (4xx, 5xx responses)

## Best Practices

1. **Always use contexts** - Never ignore the context parameter
2. **Close response bodies** - Use `defer resp.Body.Close()`
3. **Wrap errors** - Use `fmt.Errorf` with `%w` for error context
4. **Field selection** - Always specify fields to minimize response size
5. **Null handling** - Use pointers for optional fields
6. **Consistent naming** - Follow Go conventions (e.g., ID not Id)

## Common Patterns

### Batch Operations
```go
func (c *Client) BatchUpdateIssues(ctx *YouTrackContext, updates []IssueUpdate) error {
    // Use goroutines with error group for concurrent updates
}
```

### Streaming Results
```go
func (c *Client) StreamIssues(ctx *YouTrackContext, query string) (<-chan *Issue, <-chan error) {
    // Return channels for streaming large result sets
}
```

### Retry Logic
```go
func (c *Client) doRequestWithRetry(ctx *YouTrackContext, method, path string, maxRetries int) (*http.Response, error) {
    // Implement exponential backoff for transient errors
}
```

## Security Considerations

1. **Never log API keys** - Sanitize logs
2. **Use HTTPS only** - Validate base URL scheme
3. **Timeout handling** - Set reasonable timeouts
4. **Rate limiting** - Implement client-side rate limiting if needed

## Future Enhancements

Potential areas for extension:
- WebSocket support for real-time updates
- Caching layer for frequently accessed data
- Bulk operations optimization
- GraphQL API support (if YouTrack adds it)
- Metrics and instrumentation hooks

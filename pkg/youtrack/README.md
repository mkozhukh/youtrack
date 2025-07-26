# YouTrack API Client

A stateless Go client library for the YouTrack REST API. This client provides a clean, idiomatic Go interface for interacting with YouTrack's issue tracking system.

## Features

- Stateless design with per-request authentication context
- Full CRUD operations for issues, projects, and comments
- Type-safe request and response structures
- Comprehensive error handling
- Built-in pagination support
- Configurable HTTP client with timeouts

## Installation

```go
import "mkozhukh/youtrack/pkg/youtrack"
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "mkozhukh/youtrack/pkg/youtrack"
)

func main() {
    // Create a new client
    client := youtrack.NewClient("https://youtrack.example.com")
    
    // Create context with API key
    ctx := youtrack.NewYouTrackContext(context.Background(), "your-api-key")
    
    // Get an issue
    issue, err := client.GetIssue(ctx, "PROJ-123")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Issue: %s - %s\n", issue.ID, issue.Summary)
}
```

## Authentication

The client uses Bearer token authentication. Create a context with your API key for each request:

```go
ctx := youtrack.NewYouTrackContext(context.Background(), "your-api-key")
```

You can obtain an API key from your YouTrack profile settings.

## Public API Reference

### Client Creation

```go
func NewClient(baseURL string) *Client
```

Creates a new YouTrack API client with the specified base URL.

### Context Management

```go
func NewYouTrackContext(ctx context.Context, apiKey string) *YouTrackContext
```

Creates a new context with authentication information. The context can be used to cancel operations or set timeouts.

### Issues API

#### Get Issue
```go
func (c *Client) GetIssue(ctx *YouTrackContext, issueID string) (*Issue, error)
```

Retrieves a single issue by ID.

#### Create Issue
```go
func (c *Client) CreateIssue(ctx *YouTrackContext, req *CreateIssueRequest) (*Issue, error)
```

Creates a new issue. Example:
```go
req := &youtrack.CreateIssueRequest{
    Project:     "PROJ",
    Summary:     "New feature request",
    Description: "Detailed description here",
}
issue, err := client.CreateIssue(ctx, req)
```

#### Update Issue
```go
func (c *Client) UpdateIssue(ctx *YouTrackContext, issueID string, req *UpdateIssueRequest) (*Issue, error)
```

Updates an existing issue. Example:
```go
summary := "Updated summary"
assignee := "user123"  // User ID, not login
req := &youtrack.UpdateIssueRequest{
    Summary: &summary,
    Assignee: &assignee,
}
issue, err := client.UpdateIssue(ctx, "PROJ-123", req)
```

#### Update Issue Assignee
```go
func (c *Client) UpdateIssueAssignee(ctx *YouTrackContext, issueID string, assigneeLogin string) (*Issue, error)
```

Updates issue assignee by exact login. Example:
```go
issue, err := client.UpdateIssueAssignee(ctx, "PROJ-123", "john.doe")
```

#### Update Issue Assignee by Project
```go
func (c *Client) UpdateIssueAssigneeByProject(ctx *YouTrackContext, issueID string, projectID string, username string) (*Issue, error)
```

Updates issue assignee using fuzzy matching within project users. Example:
```go
// Will find and assign the first user matching "john" in project "PROJ"
issue, err := client.UpdateIssueAssigneeByProject(ctx, "PROJ-123", "PROJ", "john")
```

#### Delete Issue
```go
func (c *Client) DeleteIssue(ctx *YouTrackContext, issueID string) error
```

Deletes an issue.

#### Search Issues
```go
func (c *Client) SearchIssues(ctx *YouTrackContext, query string, skip, top int) ([]*Issue, error)
```

Searches for issues using YouTrack query language. Example:
```go
issues, err := client.SearchIssues(ctx, "project: PROJ state: Open", 0, 10)
```

### Projects API

#### Get Project
```go
func (c *Client) GetProject(ctx *YouTrackContext, projectID string) (*Project, error)
```

Retrieves project details by ID.

#### List Projects
```go
func (c *Client) ListProjects(ctx *YouTrackContext, skip, top int) ([]*Project, error)
```

Lists all accessible projects with pagination.

### Comments API

#### Get Issue Comments
```go
func (c *Client) GetIssueComments(ctx *YouTrackContext, issueID string) ([]*IssueComment, error)
```

Retrieves all comments for an issue.

#### Add Comment
```go
func (c *Client) AddIssueComment(ctx *YouTrackContext, issueID string, text string) (*IssueComment, error)
```

Adds a new comment to an issue.

#### Update Comment
```go
func (c *Client) UpdateIssueComment(ctx *YouTrackContext, issueID, commentID string, text string) (*IssueComment, error)
```

Updates an existing comment.

#### Delete Comment
```go
func (c *Client) DeleteIssueComment(ctx *YouTrackContext, issueID, commentID string) error
```

Deletes a comment from an issue.

### Users API

#### Get User
```go
func (c *Client) GetUser(ctx *YouTrackContext, userID string) (*User, error)
```

Retrieves user details by ID.

#### Search Users
```go
func (c *Client) SearchUsers(ctx *YouTrackContext, query string, skip, top int) ([]*User, error)
```

Searches for users. Example:
```go
users, err := client.SearchUsers(ctx, "john", 0, 10)
```

#### Get User by Login
```go
func (c *Client) GetUserByLogin(ctx *YouTrackContext, login string) (*User, error)
```

Finds user by login name. Example:
```go
user, err := client.GetUserByLogin(ctx, "john.doe")
```

#### Get Project Users
```go
func (c *Client) GetProjectUsers(ctx *YouTrackContext, projectID string, skip, top int) ([]*User, error)
```

Retrieves users who have access to a specific project. Example:
```go
users, err := client.GetProjectUsers(ctx, "PROJ", 0, 50)
```

#### Suggest User by Project
```go
func (c *Client) SuggestUserByProject(ctx *YouTrackContext, projectID string, username string) (*User, error)
```

Finds the first user in a project whose login, full name, or email contains the given username (case-insensitive). Perfect for fuzzy matching when you remember part of someone's name. Example:
```go
// Will match users like "john.doe", "John Smith", "johnsmith@example.com"
user, err := client.SuggestUserByProject(ctx, "PROJ", "john")
```

### Tags API

#### Get Issue Tags
```go
func (c *Client) GetIssueTags(ctx *YouTrackContext, issueID string) ([]*IssueTag, error)
```

Retrieves all tags for an issue.

#### Add Issue Tag
```go
func (c *Client) AddIssueTag(ctx *YouTrackContext, issueID string, tagName string) error
```

Adds a tag to an issue. Example:
```go
err := client.AddIssueTag(ctx, "PROJ-123", "needs-triage")
```

#### Remove Issue Tag
```go
func (c *Client) RemoveIssueTag(ctx *YouTrackContext, issueID string, tagName string) error
```

Removes a tag from an issue.

#### List Tags
```go
func (c *Client) ListTags(ctx *YouTrackContext, skip, top int) ([]*Tag, error)
```

Lists all available tags with pagination.

#### Create Tag
```go
func (c *Client) CreateTag(ctx *YouTrackContext, name string, color string) (*Tag, error)
```

Creates a new tag. Example:
```go
tag, err := client.CreateTag(ctx, "urgent", "#ff0000")
```

#### Get Tag by Name
```go
func (c *Client) GetTagByName(ctx *YouTrackContext, name string) (*Tag, error)
```

Finds a tag by its name. Example:
```go
tag, err := client.GetTagByName(ctx, "needs-triage")
```

#### Ensure Tag
```go
func (c *Client) EnsureTag(ctx *YouTrackContext, name string, color string) (string, error)
```

Returns tag ID if tag exists, otherwise creates new tag and returns its ID. Example:
```go
tagID, err := client.EnsureTag(ctx, "needs-triage", "#ffa500")
```

## Data Types

### Issue
```go
type Issue struct {
    ID          string
    Summary     string
    Description string
    Created     time.Time
    Updated     time.Time
    Resolved    *time.Time
    Reporter    *User
    UpdatedBy   *User
    Assignee    *User
    Tags        []*IssueTag
    Fields      map[string]interface{} // Custom fields
}
```

### User
```go
type User struct {
    ID       string
    Login    string
    FullName string
    Email    string
}
```

### Project
```go
type Project struct {
    ID          string
    Name        string
    ShortName   string
    Description string
}
```

### IssueComment
```go
type IssueComment struct {
    ID      string
    Author  *User
    Text    string
    Created time.Time
    Updated time.Time
}
```

### Tag
```go
type Tag struct {
    ID    string
    Name  string
    Color string
}
```

### IssueTag
```go
type IssueTag struct {
    ID    string
    Name  string
    Color string
}
```

## Error Handling

The client returns an `APIError` for YouTrack API errors:

```go
issue, err := client.GetIssue(ctx, "INVALID-ID")
if err != nil {
    if apiErr, ok := err.(*youtrack.APIError); ok {
        fmt.Printf("API error: %d - %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

## Advanced Usage

### Custom Timeouts

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

ytCtx := youtrack.NewYouTrackContext(ctx, "your-api-key")
issue, err := client.GetIssue(ytCtx, "PROJ-123")
```

### Pagination

```go
// Get issues in batches
skip := 0
top := 50

for {
    issues, err := client.SearchIssues(ctx, "project: PROJ", skip, top)
    if err != nil {
        return err
    }
    
    if len(issues) == 0 {
        break
    }
    
    // Process issues
    for _, issue := range issues {
        fmt.Printf("%s: %s\n", issue.ID, issue.Summary)
    }
    
    skip += len(issues)
}
```

### Custom Fields

Access custom fields through the `Fields` map:

```go
issue, err := client.GetIssue(ctx, "PROJ-123")
if err != nil {
    return err
}

if priority, ok := issue.Fields["Priority"]; ok {
    fmt.Printf("Priority: %v\n", priority)
}
```

## Best Practices

1. **Reuse clients**: Create one client instance and reuse it for multiple requests
2. **Handle errors**: Always check for `APIError` to get detailed error information
3. **Use contexts**: Leverage contexts for cancellation and timeouts
4. **Pagination**: Use pagination for large result sets to avoid timeouts
5. **Field selection**: The client automatically selects commonly used fields to minimize response size

## Requirements

- Go 1.19 or higher
- Valid YouTrack instance URL
- YouTrack API key with appropriate permissions

## License

This client is part of the YouTrack MCP project. See the parent project for license information.
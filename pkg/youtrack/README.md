# YouTrack API Client

A stateless Go client library for the YouTrack REST API. Provides an idiomatic Go interface for issue tracking, project management, time tracking, and more.

## Installation

```go
import "mkozhukh/youtrack/pkg/youtrack"
```

## Quick Start

```go
client := youtrack.NewClient("https://youtrack.example.com")
ctx := youtrack.NewYouTrackContext(context.Background(), "your-api-key")

issue, err := client.GetIssue(ctx, "PROJ-123")
```

## Authentication

Bearer token via `YouTrackContext`. Obtain an API key from your YouTrack profile settings.

```go
ctx := youtrack.NewYouTrackContext(context.Background(), "your-api-key")
```

## API Reference

### Issues

| Method | Signature | Description |
|---|---|---|
| GetIssue | `(issueID) -> Issue` | Get issue by readable ID |
| CreateIssue | `(req) -> Issue` | Create issue with project, summary, description, custom fields |
| UpdateIssue | `(issueID, req) -> Issue` | Update summary, description, or custom fields |
| UpdateIssueAssignee | `(issueID, login) -> Issue` | Set assignee by exact login |
| UpdateIssueAssigneeByProject | `(issueID, projectID, username) -> Issue` | Set assignee by fuzzy match within project members |
| DeleteIssue | `(issueID) -> error` | Delete an issue |
| SearchIssues | `(query, skip, top) -> []Issue` | Search using YouTrack query language, paginated |
| SearchIssuesSorted | `(query, skip, top, sortBy, sortOrder) -> []Issue` | Search with `sort by:` clause appended |
| ApplyCommand | `(issueID, command) -> error` | Apply a YouTrack command (e.g. `"State Open"`, `"Priority Critical"`) |
| GetIssueCustomFields | `(issueID) -> []CustomFieldValue` | Get all custom field values for an issue |
| GetAvailableLinkTypes | `() -> []LinkType` | List all link types (e.g. "Depends on", "Subtask of") |
| CreateIssueLink | `(sourceID, targetID, linkType) -> error` | Link two issues via command |
| GetIssueLinks | `(issueID) -> []IssueLink` | Get all links for an issue |
| GetIssueActivities | `(issueID) -> []ActivityItem` | Get full activity/history log |

### Comments

| Method | Signature | Description |
|---|---|---|
| GetIssueComments | `(issueID) -> []IssueComment` | List all comments |
| AddIssueComment | `(issueID, text) -> IssueComment` | Add a comment |
| UpdateIssueComment | `(issueID, commentID, text) -> IssueComment` | Update a comment |
| DeleteIssueComment | `(issueID, commentID) -> error` | Delete a comment |

### Tags

| Method | Signature | Description |
|---|---|---|
| GetIssueTags | `(issueID) -> []IssueTag` | Get tags on an issue |
| AddIssueTag | `(issueID, tagName) -> error` | Add a tag to an issue |
| RemoveIssueTag | `(issueID, tagName) -> error` | Remove a tag from an issue |
| ListTags | `(skip, top) -> []Tag` | List all tags, paginated |
| CreateTag | `(name, color) -> Tag` | Create a new tag |
| GetTagByName | `(name) -> Tag` | Find tag by exact name |
| EnsureTag | `(name, color) -> tagID` | Get or create tag, return ID |

### Attachments

| Method | Signature | Description |
|---|---|---|
| GetIssueAttachments | `(issueID) -> []Attachment` | List attachment metadata |
| AddIssueAttachment | `(issueID, filePath) -> Attachment` | Upload a file (multipart) |
| GetIssueAttachmentContent | `(issueID, attachmentID) -> []byte` | Download raw attachment bytes |

### Worklogs

| Method | Signature | Description |
|---|---|---|
| GetIssueWorklogs | `(issueID) -> []WorkItem` | List work items for an issue |
| AddIssueWorklog | `(issueID, req) -> WorkItem` | Add work item (duration in minutes) |
| GetUserWorklogs | `(userID, projectID, start, end, skip, top) -> []WorkItem` | User's work items, filtered by project/dates |

### Projects

| Method | Signature | Description |
|---|---|---|
| GetProject | `(projectID) -> Project` | Get project by internal ID |
| GetProjectByName | `(name) -> Project` | Find by name or short name (case-insensitive) |
| ListProjects | `(skip, top) -> []Project` | List all projects, paginated |
| GetProjectIssues | `(projectID, skip, top) -> []Issue` | Issues in a project, paginated |
| GetProjectCustomFields | `(projectID) -> []CustomField` | Custom field definitions for a project |
| GetCustomFieldAllowedValues | `(projectID, fieldName) -> []AllowedValue` | Allowed values for a bundle-backed field |
| AddCustomFieldEnumValue | `(projectID, fieldName, value, color) -> error` | Add enum value to a field's bundle |

### Users

| Method | Signature | Description |
|---|---|---|
| GetCurrentUser | `() -> User` | Authenticated user's profile |
| GetUser | `(userID) -> User` | Get user by internal ID |
| SearchUsers | `(query, skip, top) -> []User` | Search users, paginated |
| GetUserByLogin | `(login) -> User` | Find by exact login |
| GetProjectUsers | `(projectID, skip, top) -> []User` | Project members, paginated |
| SuggestUserByProject | `(projectID, username) -> User` | Fuzzy match user in project (login/name/email) |

## Data Types

```go
type Issue struct {
    ID          string        // Readable ID, e.g. "PROJ-123"
    Summary     string
    Description string
    Created     YouTrackTime
    Updated     YouTrackTime
    Resolved    *YouTrackTime
    Reporter    *User
    UpdatedBy   *User
    Assignee    *User
    Tags        []*IssueTag
}

type User struct {
    ID       string
    Login    string
    FullName string
    Email    string
}

type Project struct {
    ID          string
    Name        string
    ShortName   string
    Description string
}

type IssueComment struct {
    ID      string
    Author  *User
    Text    string
    Created YouTrackTime
    Updated YouTrackTime
}

type Tag struct {
    ID    string
    Name  string
    Color YouTrackColor
}

type WorkItem struct {
    ID          string
    Author      *User
    Date        YouTrackTime
    Duration    int          // minutes
    Description string
    Type        *WorkType
    Issue       *Issue
}

type Attachment struct {
    ID       string
    Name     string
    Size     int64
    Created  YouTrackTime
    Author   *User
    MimeType string
    URL      string
}

type IssueLink struct {
    ID        string
    Direction string
    LinkType  *LinkType
    Issues    []*Issue
}

type CustomFieldValue struct {
    Name  string
    Type  string       // $type
    Value interface{}  // nested object with name, id, $type
}

type AllowedValue struct {
    ID   string
    Name string
}

type ActivityItem struct {
    ID        string
    Category  Category
    Author    *User
    Timestamp YouTrackTime
    Field     *Field
    Added     *FieldValue
    Removed   *FieldValue
}
```

## Error Handling

API errors are returned as `*APIError`:

```go
issue, err := client.GetIssue(ctx, "INVALID-ID")
if err != nil {
    if apiErr, ok := err.(*youtrack.APIError); ok {
        fmt.Printf("status %d: %s\n", apiErr.StatusCode, apiErr.Message)
    }
}
```

## Pagination

All list methods support `skip`/`top` parameters:

```go
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
    for _, issue := range issues {
        fmt.Printf("%s: %s\n", issue.ID, issue.Summary)
    }
    skip += len(issues)
}
```

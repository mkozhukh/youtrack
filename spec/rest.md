# YouTrack REST Client API

High-level reference for `pkg/youtrack` — the Go REST client used by MCP handlers and CLI.

## Client Setup

`NewClient(baseURL string) *Client` — creates a client with the YouTrack instance base URL (e.g. `https://myteam.youtrack.cloud`). All methods require a `*YouTrackContext` carrying a `context.Context` and an API key (Bearer token).

Errors from the API are returned as `*APIError{StatusCode, Message}`.

## Data Types

| Type | Key Fields |
|---|---|
| `Issue` | `ID` (readable, e.g. `PROJ-123`), `Summary`, `Description`, `Created`, `Updated`, `Resolved`, `Reporter`, `UpdatedBy`, `Assignee`, `Tags` |
| `User` | `ID`, `Login`, `FullName`, `Email` |
| `Project` | `ID`, `Name`, `ShortName`, `Description` |
| `IssueComment` | `ID`, `Author`, `Text`, `Created`, `Updated` |
| `Tag` / `IssueTag` | `ID`, `Name`, `Color` |
| `WorkItem` | `ID`, `Author`, `Date`, `Duration` (minutes), `Description`, `Type`, `Issue` |
| `Attachment` | `ID`, `Name`, `Size`, `Created`, `Author`, `MimeType`, `URL` |
| `IssueLink` | `ID`, `Direction`, `LinkType`, `Issues` |
| `LinkType` | `ID`, `Name` |
| `CustomField` | `Name`, `Type` (`$type`), `Value` |
| `CustomFieldValue` | `Name`, `Type` (`$type`), `Value` (with nested `name`, `id`, `$type`) |
| `AllowedValue` | `ID`, `Name` |
| `ActivityItem` | `ID`, `Category`, `Author`, `Timestamp`, `Field`, `Added`, `Removed` |

## Issues

### GetIssue(issueID) -> Issue
Get a single issue by its readable ID (e.g. `PROJ-123`). Returns full issue with reporter, assignee, tags.

### CreateIssue(req) -> Issue
Create an issue. Request includes project (by `ShortName`), summary, description, and optional custom fields.

### UpdateIssue(issueID, req) -> Issue
Update summary, description, or custom fields of an existing issue.

### UpdateIssueAssignee(issueID, assigneeLogin) -> Issue
Set assignee by exact login. Resolves user first via `GetUserByLogin`.

### UpdateIssueAssigneeByProject(issueID, projectID, username) -> Issue
Set assignee by fuzzy match within project members. Uses `SuggestUserByProject`.

### DeleteIssue(issueID) -> error
Delete an issue.

### SearchIssues(query, skip, top) -> []Issue
Search issues using YouTrack query language. Paginated with `skip`/`top`.

### SearchIssuesSorted(query, skip, top, sortBy, sortOrder) -> []Issue
Same as `SearchIssues` but appends `sort by: {sortBy} {sortOrder}` to the query string.

### ApplyCommand(issueID, command) -> error
Apply a YouTrack command to an issue (e.g. `"State Open"`, `"Priority Critical"`, `"assignee me"`). Uses the commands API.

### GetIssueCustomFields(issueID) -> []CustomFieldValue
Get all custom field values for an issue, including field name, type, and value (with nested name/id).

### GetAvailableLinkTypes() -> []LinkType
List all available issue link types in the YouTrack instance (e.g. "Depends on", "Subtask of").

### CreateIssueLink(sourceIssueID, targetIssueID, linkType) -> error
Create a link between two issues using a command (e.g. `"Depends on PROJ-456"`).

### GetIssueLinks(issueID) -> []IssueLink
Get all links for an issue, including direction, link type, and linked issues.

### GetIssueActivities(issueID) -> []ActivityItem
Get the full activity/history log of an issue: field changes, comments added/removed, etc.

## Comments

### GetIssueComments(issueID) -> []IssueComment
List all comments on an issue.

### AddIssueComment(issueID, text) -> IssueComment
Add a comment to an issue.

### UpdateIssueComment(issueID, commentID, text) -> IssueComment
Update an existing comment.

### DeleteIssueComment(issueID, commentID) -> error
Delete a comment.

## Tags

### GetIssueTags(issueID) -> []IssueTag
Get all tags on an issue.

### AddIssueTag(issueID, tagName) -> error
Add a tag to an issue by name. Creates association if the tag exists.

### RemoveIssueTag(issueID, tagName) -> error
Remove a tag from an issue.

### ListTags(skip, top) -> []Tag
List all tags in the instance. Paginated.

### CreateTag(name, color) -> Tag
Create a new tag with optional color.

### GetTagByName(name) -> Tag
Find a tag by exact name.

### EnsureTag(name, color) -> tagID
Get or create a tag by name. Returns the tag ID.

## Attachments

### GetIssueAttachments(issueID) -> []Attachment
List all attachments on an issue (metadata only: name, size, mime type, etc.).

### AddIssueAttachment(issueID, filePath) -> Attachment
Upload a local file as an attachment to an issue. Uses multipart form upload.

### AddIssueAttachmentFromBytes(issueID, content, filename) -> Attachment
Upload in-memory bytes as an attachment to an issue. Uses multipart form upload.

### GetIssueAttachmentContent(issueID, attachmentID) -> []byte
Download the raw binary content of an attachment.

## Worklogs

### GetIssueWorklogs(issueID) -> []WorkItem
List all work items (time entries) for an issue.

### AddIssueWorklog(issueID, req) -> WorkItem
Add a work item to an issue. Request includes duration (minutes) and optional description.

### GetUserWorklogs(userID, projectID, startDate, endDate, skip, top) -> []WorkItem
Get work items for a specific user, optionally filtered by project and date range. Paginated.

## Projects

### GetProject(projectID) -> Project
Get project details by internal ID.

### GetProjectByName(name) -> Project
Find a project by name or short name (case-insensitive). Iterates all projects with pagination.

### ListProjects(skip, top) -> []Project
List all projects. Paginated.

### GetProjectIssues(projectID, skip, top) -> []Issue
Get issues belonging to a project. Paginated. Uses query `project:{projectID}`.

### GetProjectCustomFields(projectID) -> []CustomField
Get the custom field definitions configured for a project (field name, type).

### GetCustomFieldAllowedValues(projectID, fieldName) -> []AllowedValue
Get allowed values for a bundle-backed custom field (enum, state, version, build, owned).
Resolves the bundle type automatically from the field's `$type`.

### AddCustomFieldEnumValue(projectID, fieldName, valueName, color) -> error
Add a new value to an enum-type custom field's bundle. Resolves the bundle ID from the project's field configuration.

## Users

### GetCurrentUser() -> User
Get the authenticated user's profile (login, name, email).

### GetUser(userID) -> User
Get a user by internal ID.

### SearchUsers(query, skip, top) -> []User
Search users by query string. Paginated.

### GetUserByLogin(login) -> User
Find a user by exact login.

### GetProjectUsers(projectID, skip, top) -> []User
List users that are members of a project. Paginated.

### SuggestUserByProject(projectID, username) -> User
Fuzzy-find a user within a project's members. Matches against login, full name, and email (case-insensitive substring match). Iterates all members with pagination.

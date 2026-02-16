# MCP Server Specification

This server provides MCP tools for interacting with YouTrack.

## Tools

### Issues

- `get_issue_list`: Retrieve a list of issues from YouTrack with optional filtering and sorting.
  - `project_id` (string, required): Project ID to search issues in.
  - `query` (string, optional): YouTrack query string for filtering issues.
  - `max_results` (number, optional): Maximum number of results to return.
  - `sort_by` (string, optional): Field to sort by (e.g., 'created', 'updated', 'priority').
  - `sort_order` (string, optional): Sort order: 'asc' or 'desc' (defaults to 'desc').

- `get_issue_details`: Get detailed information about a specific issue including comments.
  - `issue_id` (string, required): Issue ID to retrieve details for.

- `create_issue`: Create a new issue in YouTrack.
  - `project_id` (string, required): Project ID where the issue should be created.
  - `summary` (string, required): Issue summary/title.
  - `description` (string, optional): Issue description.

- `update_issue`: Update an existing issue in YouTrack.
  - `issue_id` (string, required): Issue ID to update.
  - `state` (string, optional): New state for the issue.
  - `assignee` (string, optional): New assignee login/username for the issue.
  - `summary` (string, optional): New summary for the issue.
  - `description` (string, optional): New description for the issue.

- `delete_issue`: Delete an issue from YouTrack.
  - `issue_id` (string, required): Issue ID to delete.

- `apply_command`: Execute a YouTrack command on an issue (e.g., 'State Open', 'Priority Critical').
  - `issue_id` (string, required): Issue ID to apply the command to.
  - `command` (string, required): YouTrack command string to execute.

### Tags

- `tag_issue`: Add a tag to an issue. Creates the tag if it doesn't exist.
  - `issue_id` (string, required): Issue ID to add the tag to.
  - `tag` (string, required): Tag name to add to the issue.

- `untag_issue`: Remove a tag from an issue.
  - `issue_id` (string, required): Issue ID to remove the tag from.
  - `tag` (string, required): Tag name to remove from the issue.

- `search_tags`: Search for tags by partial name match (case-insensitive).
  - `query` (string, required): Partial tag name to search for.

### Comments

- `add_comment`: Add a comment to an issue.
  - `issue_id` (string, required): Issue ID to add the comment to.
  - `comment` (string, required): Comment text to add to the issue.

### Links

- `get_issue_links`: Get all links for a specific issue, grouped by link type and direction.
  - `issue_id` (string, required): Issue ID to retrieve links for.

- `create_issue_link`: Create a link between two issues.
  - `source_issue_id` (string, required): Source issue ID.
  - `target_issue_id` (string, required): Target issue ID.
  - `link_type` (string, required): Link type name (e.g., 'depends on', 'relates to', 'parent for').

### Attachments

- `get_issue_attachments`: List all attachments for a specific issue with metadata.
  - `issue_id` (string, required): Issue ID to retrieve attachments for.

- `get_issue_attachment_content`: Download the content of a specific attachment.
  - `issue_id` (string, required): Issue ID the attachment belongs to.
  - `attachment_id` (string, required): Attachment ID to download.

- `upload_attachment`: Upload an attachment to an issue. Content must be base64-encoded. Max 10MB.
  - `issue_id` (string, required): Issue ID to attach the file to.
  - `content` (string, required): Base64-encoded file content.
  - `filename` (string, required): Name of the file to create.

### Worklogs

- `add_worklog`: Log work time on an issue.
  - `issue_id` (string, required): Issue ID to log work on.
  - `duration` (number, required): Duration in minutes.
  - `text` (string, optional): Description of the work performed.
  - `date` (string, optional): Date in YYYY-MM-DD format (defaults to today).
  - `work_type` (string, optional): Type of work (e.g., 'Development', 'Testing').

- `get_issue_worklogs`: Get all work items logged on a specific issue.
  - `issue_id` (string, required): Issue ID to retrieve worklogs for.

- `get_user_worklogs`: Get work items logged by a specific user.
  - `user_id` (string, optional): User ID (defaults to current user).
  - `project_id` (string, optional): Filter by project ID.
  - `start_date` (string, optional): Start date in YYYY-MM-DD format.
  - `end_date` (string, optional): End date in YYYY-MM-DD format.

### Projects

- `get_project_info`: Get project schema including custom fields with allowed values and link types.
  - `project_id` (string, required): Project ID (short name) to retrieve info for.

- `list_projects`: List available YouTrack projects.
  - `query` (string, optional): Project name to search for (case-insensitive).

### Users

- `get_current_user`: Get the authenticated user's profile information.

- `get_project_users`: List all users who are members of a specific project.
  - `project_id` (string, required): Project ID (short name) to retrieve users for.

### Cache

- `drop_cache`: Drop cached project metadata (custom fields, users) to force refresh.
  - `project_id` (string, optional): Project ID to drop cache for. If empty, drops all.

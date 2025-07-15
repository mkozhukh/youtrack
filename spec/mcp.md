### MCP Server Specification

This server will provide a command-line interface for interacting with YouTrack.

**Tools:**

1.  **`get_issue_list`**: Retrieves a list of issues from a specific YouTrack project.
    *   **`project_id`** (string, required): The ID of the YouTrack project (e.g., "TEST").
    *   **`query`** (string, optional): A search query to filter issues (e.g., "for: me #unresolved"). Defaults to showing issues updated in the last 7 days.
    *   **`max_results`** (integer, optional): The maximum number of issues to return. Defaults to 10.

2.  **`get_issue_details`**: Retrieves a single issue with all its comments.
    *   **`issue_id`** (string, required): The ID of the issue to retrieve (e.g., "TEST-123").

3.  **`create_issue`**: Creates a new issue.
    *   **`project_id`** (string, required): The ID of the project to create the issue in.
    *   **`summary`** (string, required): The summary or title of the issue.
    *   **`description`** (string, optional): The full description of the issue.

4.  **`update_issue`**: Updates the state or assignee of an issue.
    *   **`issue_id`** (string, required): The ID of the issue to update.
    *   **`state`** (string, optional): The new state for the issue (e.g., "In Progress", "Done").
    *   **`assignee`** (string, optional): The login name of the user to assign the issue to.

5.  **`tag_issue`**: Adds a tag to a specific issue. If the tag doesn't exist, it will be created.
    *   **`issue_id`** (string, required): The ID of the issue to tag (e.g., "TEST-123").
    *   **`tag`** (string, required): The tag to add (e.g., "needs-review").

6.  **`add_comment`**: Adds a comment to a specific issue.
    *   **`issue_id`** (string, required): The ID of the issue to comment on (e.g., "TEST-123").
    *   **`comment`** (string, required): The content of the comment.

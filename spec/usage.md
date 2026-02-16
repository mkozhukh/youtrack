### YouTrack MCP Usage Scenarios

This document outlines typical workflows using the YouTrack MCP (Model Context Protocol) server.

---

## Scenario 1: Support Engineer - Bug Report Triage

**Context:** A customer reports a bug where the "Export to PDF" feature is failing.

#### Step 1: Check for Existing Issues

*   **Action:** Search for existing issues.
*   **Tool Used:** `get_issue_list`
*   **Command:**
    ```
    get_issue_list(project_id="METRICO", query="PDF #unresolved")
    ```
*   **Expected Outcome:** Returns a list of matching issues, or empty if none found.

#### Step 2: Create a New Issue

*   **Action:** Create a new bug report.
*   **Tool Used:** `create_issue`
*   **Command:**
    ```
    create_issue(
      project_id="METRICO",
      summary="Export to PDF feature fails on large reports",
      description="When a user tries to export a report with over 500 rows, the process times out."
    )
    ```
*   **Expected Outcome:** Issue created, returns issue ID (e.g., `METRICO-5821`).

#### Step 3: Assign the Issue

*   **Action:** Assign the bug to a developer.
*   **Tool Used:** `update_issue`
*   **Command:**
    ```
    update_issue(issue_id="METRICO-5821", assignee="john.doe")
    ```
*   **Expected Outcome:** Issue assigned. Partial name match works (e.g., "john" matches "john.doe").

#### Step 4: Add Comment

*   **Action:** Add context from customer.
*   **Tool Used:** `add_comment`
*   **Command:**
    ```
    add_comment(issue_id="METRICO-5821", comment="Customer provided logs. See attachment.")
    ```
*   **Expected Outcome:** Comment added to the issue.

#### Step 5: Tag for Prioritization

*   **Action:** Tag for triage meeting.
*   **Tool Used:** `tag_issue`
*   **Command:**
    ```
    tag_issue(issue_id="METRICO-5821", tag="needs-triage")
    ```
*   **Expected Outcome:** Tag created (if new) and applied to issue.

#### Step 6: Verify

*   **Action:** Verify all updates.
*   **Tool Used:** `get_issue_details`
*   **Command:**
    ```
    get_issue_details(issue_id="METRICO-5821")
    ```
*   **Expected Outcome:** Returns full issue details including assignee, comments, and tags.

---

## Scenario 2: Developer - Issue Investigation

**Context:** Developer investigates an assigned issue.

#### Step 1: Get Issue Details

*   **Action:** Read issue details and comments.
*   **Tool Used:** `get_issue_details`
*   **Command:**
    ```
    get_issue_details(issue_id="METRICO-5821")
    ```
*   **Expected Outcome:** Returns summary, description, comments, custom fields.

#### Step 2: Check Related Issues

*   **Action:** Find linked/blocking issues.
*   **Tool Used:** `get_issue_links`
*   **Command:**
    ```
    get_issue_links(issue_id="METRICO-5821")
    ```
*   **Expected Outcome:** Returns all linked issues grouped by link type.

#### Step 3: Get Attachments

*   **Action:** List attached files.
*   **Tool Used:** `get_issue_attachments`
*   **Command:**
    ```
    get_issue_attachments(issue_id="METRICO-5821")
    ```
*   **Expected Outcome:** Returns list of attachments with IDs, names, sizes.

#### Step 4: Download Attachment

*   **Action:** Read attachment content.
*   **Tool Used:** `get_issue_attachment_content`
*   **Command:**
    ```
    get_issue_attachment_content(issue_id="METRICO-5821", attachment_id="123-456")
    ```
*   **Expected Outcome:** Returns file content (text or base64 for binary).

---

## Scenario 3: Developer - Status Update Workflow

**Context:** Developer updates issue status as work progresses.

#### Step 1: Start Work

*   **Action:** Change state to In Progress.
*   **Tool Used:** `update_issue`
*   **Command:**
    ```
    update_issue(issue_id="METRICO-5821", state="In Progress")
    ```
*   **Expected Outcome:** State updated. Partial match works ("progress" → "In Progress").

#### Step 2: Add Progress Comment

*   **Action:** Document progress.
*   **Tool Used:** `add_comment`
*   **Command:**
    ```
    add_comment(issue_id="METRICO-5821", comment="Root cause identified: memory limit exceeded.")
    ```
*   **Expected Outcome:** Comment added.

#### Step 3: Complete Work

*   **Action:** Mark as fixed.
*   **Tool Used:** `update_issue`
*   **Command:**
    ```
    update_issue(issue_id="METRICO-5821", state="Fixed")
    ```
*   **Expected Outcome:** State updated to Fixed.

---

## Scenario 4: Link Related Issues

**Context:** Developer finds a related/duplicate issue.

#### Step 1: Search for Related Issues

*   **Action:** Find potentially related issues.
*   **Tool Used:** `get_issue_list`
*   **Command:**
    ```
    get_issue_list(project_id="METRICO", query="PDF export timeout")
    ```
*   **Expected Outcome:** Returns list of matching issues.

#### Step 2: Create Link

*   **Action:** Link issues together.
*   **Tool Used:** `create_issue_link`
*   **Command:**
    ```
    create_issue_link(
      source_issue_id="METRICO-5821",
      target_issue_id="METRICO-5100",
      link_type="relates to"
    )
    ```
*   **Expected Outcome:** Link created between the two issues.

Link types: `"depends on"`, `"is required for"`, `"relates to"`, `"duplicates"`, `"parent for"`, `"subtask of"`

---

## Scenario 5: Log Work Time

**Context:** Developer logs time spent on an issue.

#### Step 1: Add Worklog

*   **Action:** Log 2 hours of work.
*   **Tool Used:** `add_worklog`
*   **Command:**
    ```
    add_worklog(
      issue_id="METRICO-5821",
      duration=120,
      text="Investigated and fixed memory leak"
    )
    ```
*   **Expected Outcome:** Worklog entry created (duration in minutes).

#### Step 2: View Issue Worklogs

*   **Action:** Check all logged time on issue.
*   **Tool Used:** `get_issue_worklogs`
*   **Command:**
    ```
    get_issue_worklogs(issue_id="METRICO-5821")
    ```
*   **Expected Outcome:** Returns all work items with author, date, duration.

---

## Scenario 6: Find User's Project

**Context:** Agent needs to determine which project to use.

#### Step 1: Get Current User

*   **Action:** Get user info (may include default project).
*   **Tool Used:** `get_current_user`
*   **Command:**
    ```
    get_current_user()
    ```
*   **Expected Outcome:** Returns user profile. If no default project, proceed to step 2.

#### Step 2: List Projects (if needed)

*   **Action:** List available projects.
*   **Tool Used:** `list_projects`
*   **Command:**
    ```
    list_projects()
    ```
*   **Expected Outcome:** Returns list of projects. Ask user to select one.

---

## Scenario 7: Get Correct Field Values

**Context:** Agent needs to know valid values for State, Priority, etc.

#### Step 1: Get Project Info

*   **Action:** Retrieve project schema with field options.
*   **Tool Used:** `get_project_info`
*   **Command:**
    ```
    get_project_info(project_id="METRICO")
    ```
*   **Expected Outcome:** Returns custom fields with allowed values (e.g., State: [Open, In Progress, Fixed]).

#### Step 2: Use Exact Value

*   **Action:** Apply field change with correct value.
*   **Tool Used:** `apply_command`
*   **Command:**
    ```
    apply_command(issue_id="METRICO-5821", command="Priority Critical")
    ```
*   **Expected Outcome:** Field updated using exact value from allowed list.

---

## Scenario 8: Find Correct Username

**Context:** Agent needs to assign issue but unsure of exact username.

#### Step 1: Try Partial Match

*   **Action:** Assign with partial name (auto-resolved).
*   **Tool Used:** `update_issue`
*   **Command:**
    ```
    update_issue(issue_id="METRICO-5821", assignee="john")
    ```
*   **Expected Outcome:** If single match, assigns automatically. If multiple/none, returns error.

#### Step 2: List Project Users (if needed)

*   **Action:** Get list of valid usernames.
*   **Tool Used:** `get_project_users`
*   **Command:**
    ```
    get_project_users(project_id="METRICO")
    ```
*   **Expected Outcome:** Returns all project members with login names.

### YouTrack MCP Usage Scenario: Support Engineer Workflow

This document outlines a typical workflow for a support engineer using the YouTrack MCP (Machine-to-Machine Command Protocol) to manage issues.

**Scenario:** A customer reports a bug where the "Export to PDF" feature is failing in the "Metrico" application.

---

#### Step 1: Triage - Check for Existing Issues

The first step is to see if this bug has already been reported. The support engineer searches for open issues in the "Metrico" project containing the "PDF" keyword.

*   **Action:** Search for existing issues.
*   **Tool Used:** `get_issue_list`
*   **Command:**
    ```
    get_issue_list(project_id="METRICO", query="PDF #unresolved")
    ```
*   **Expected Outcome:** The tool returns a list of issues. In this case, let's assume it returns an empty list, indicating this is a new bug.

---

#### Step 2: Create a New Issue

Since no existing issue was found, the support engineer creates a new one.

*   **Action:** Create a new bug report.
*   **Tool Used:** `create_issue`
*   **Command:**
    ```
    create_issue(
      project_id="METRICO",
      summary="Export to PDF feature fails on large reports",
      description="When a user tries to export a report with over 500 rows, the process times out and results in a 504 Gateway Timeout error. This has been reproduced on staging environment."
    )
    ```
*   **Expected Outcome:** The tool successfully creates the issue in YouTrack and returns the new issue ID, for example, `{"issue_id": "METRICO-5821"}`.

---

#### Step 3: Assign the Issue

The support engineer now assigns the newly created issue to the appropriate developer, "john.doe", for investigation.

*   **Action:** Assign the bug to a developer.
*   **Tool Used:** `update_issue`
*   **Command:**
    ```
    update_issue(issue_id="METRICO-5821", assignee="john.doe")
    ```
*   **Expected Outcome:** The tool finds the user "john.doe", assigns the issue to them, and confirms the update was successful. The underlying REST client would first call the `GET /api/users` endpoint to find the correct user object for "john.doe" before making the update call.

---

#### Step 4: Add Contextual Information

The customer provided a log file. The support engineer adds a comment to the issue with a link to the log.

*   **Action:** Add a comment with a link to logs.
*   **Tool Used:** `add_comment`
*   **Command:**
    ```
    add_comment(
      issue_id="METRICO-5821",
      comment="Customer has provided logs. See attached file: [link_to_logs.txt]"
    )
    ```
*   **Expected Outcome:** The comment is successfully added to the issue `METRICO-5821`.

---

#### Step 5: Tag the Issue for Prioritization

To ensure the issue is reviewed in the next triage meeting, the support engineer adds a "needs-triage" tag.

*   **Action:** Tag the issue for the next triage meeting.
*   **Tool Used:** `tag_issue`
*   **Command:**
    ```
    tag_issue(issue_id="METRICO-5821", tag="needs-triage")
    ```
*   **Expected Outcome:** The "needs-triage" tag is created if it doesn't exist and is then successfully applied to the issue.

---

#### Step 6: Final Verification

The support engineer does a final check to ensure all the details—assignee, comment, and tag—are correctly reflected on the issue.

*   **Action:** Verify all updates on the issue.
*   **Tool Used:** `get_issue_details`
*   **Command:**
    ```
    get_issue_details(issue_id="METRICO-5821")
    ```
*   **Expected Outcome:** The tool returns a detailed view of issue `METRICO-5821`, including its summary, description, current state, the assignee ("john.doe"), all comments, and the "needs-triage" tag. This confirms the issue is correctly filed and routed.

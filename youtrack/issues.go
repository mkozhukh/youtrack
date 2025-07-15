package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) GetIssue(ctx *YouTrackContext, issueID string) (*Issue, error) {
	path := fmt.Sprintf("/api/issues/%s", issueID)
	
	query := url.Values{}
	query.Add("fields", "id,summary,description,created,updated,resolved,reporter(id,login,fullName,email),updatedBy(id,login,fullName,email),assignee(id,login,fullName,email),tags(id,name,color),customFields(name,value)")
	
	resp, err := c.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode issue: %w", err)
	}

	return &issue, nil
}

func (c *Client) CreateIssue(ctx *YouTrackContext, req *CreateIssueRequest) (*Issue, error) {
	resp, err := c.Post(ctx, "/api/issues", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode issue: %w", err)
	}

	return &issue, nil
}

func (c *Client) UpdateIssue(ctx *YouTrackContext, issueID string, req *UpdateIssueRequest) (*Issue, error) {
	path := fmt.Sprintf("/api/issues/%s", issueID)
	
	resp, err := c.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode issue: %w", err)
	}

	return &issue, nil
}

func (c *Client) UpdateIssueAssignee(ctx *YouTrackContext, issueID string, assigneeLogin string) (*Issue, error) {
	user, err := c.GetUserByLogin(ctx, assigneeLogin)
	if err != nil {
		return nil, fmt.Errorf("failed to find user '%s': %w", assigneeLogin, err)
	}

	req := &UpdateIssueRequest{
		Assignee: &user.ID,
	}

	return c.UpdateIssue(ctx, issueID, req)
}

func (c *Client) UpdateIssueAssigneeByProject(ctx *YouTrackContext, issueID string, projectID string, username string) (*Issue, error) {
	user, err := c.SuggestUserByProject(ctx, projectID, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user matching '%s' in project '%s': %w", username, projectID, err)
	}

	req := &UpdateIssueRequest{
		Assignee: &user.ID,
	}

	return c.UpdateIssue(ctx, issueID, req)
}

func (c *Client) DeleteIssue(ctx *YouTrackContext, issueID string) error {
	path := fmt.Sprintf("/api/issues/%s", issueID)
	
	resp, err := c.Delete(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) SearchIssues(ctx *YouTrackContext, query string, skip, top int) ([]*Issue, error) {
	params := url.Values{}
	params.Add("query", query)
	params.Add("$skip", fmt.Sprintf("%d", skip))
	params.Add("$top", fmt.Sprintf("%d", top))
	params.Add("fields", "id,summary,description,created,updated,resolved,reporter(id,login,fullName,email),updatedBy(id,login,fullName,email),assignee(id,login,fullName,email),tags(id,name,color),customFields(name,value)")

	resp, err := c.Get(ctx, "/api/issues", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issues []*Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("failed to decode issues: %w", err)
	}

	return issues, nil
}
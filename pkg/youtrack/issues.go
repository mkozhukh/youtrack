package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) GetIssue(ctx *YouTrackContext, issueID string) (*Issue, error) {
	path := fmt.Sprintf("/api/issues/%s", issueID)

	query := url.Values{}
	query.Add("fields", "idReadable,summary,description,created,updated,resolved,reporter(id,login,fullName,email),updatedBy(id,login,fullName,email),assignee(id,login,fullName,email),tags(id,name,color)")

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
	// Add fields parameter to get the full issue details in response
	query := url.Values{}
	query.Add("fields", "idReadable,summary,description,created,updated,resolved,reporter(id,login,fullName,email),updatedBy(id,login,fullName,email),assignee(id,login,fullName,email),tags(id,name,color)")

	resp, err := c.PostWithQuery(ctx, "/api/issues", query, req)
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
	params.Add("fields", "idReadable,summary,description,created,updated,resolved,reporter(id,login,fullName,email),updatedBy(id,login,fullName,email),assignee(id,login,fullName,email),tags(id,name,color)")

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

func (c *Client) GetIssueWorklogs(ctx *YouTrackContext, issueID string) ([]*WorkItem, error) {
	path := fmt.Sprintf("/api/issues/%s/timeTracking/workItems", issueID)

	query := url.Values{}
	query.Add("fields", "id,date,duration,text,author(id,login,fullName,email),type(id,name)")

	resp, err := c.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var workItems []*WorkItem
	if err := json.NewDecoder(resp.Body).Decode(&workItems); err != nil {
		return nil, fmt.Errorf("failed to decode work items: %w", err)
	}

	return workItems, nil
}

func (c *Client) AddIssueWorklog(ctx *YouTrackContext, issueID string, req *CreateWorklogRequest) (*WorkItem, error) {
	path := fmt.Sprintf("/api/issues/%s/timeTracking/workItems", issueID)

	query := url.Values{}
	query.Add("fields", "id,date,duration,text,author(id,login,fullName,email),type(id,name)")

	resp, err := c.PostWithQuery(ctx, path, query, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var workItem WorkItem
	if err := json.NewDecoder(resp.Body).Decode(&workItem); err != nil {
		return nil, fmt.Errorf("failed to decode work item: %w", err)
	}

	return &workItem, nil
}

func (c *Client) CreateIssueLink(ctx *YouTrackContext, sourceIssueID, targetIssueID, linkType string) error {
	req := &CreateIssueLinkRequest{
		Query: fmt.Sprintf("%s %s", linkType, targetIssueID),
		Issues: []*IssueRef{
			{ID: sourceIssueID},
		},
	}

	resp, err := c.Post(ctx, "/api/commands", req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) GetIssueLinks(ctx *YouTrackContext, issueID string) ([]*IssueLink, error) {
	path := fmt.Sprintf("/api/issues/%s/links", issueID)

	query := url.Values{}
	query.Add("fields", "id,direction,linkType(id,name),issues(idReadable,summary)")

	resp, err := c.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var links []*IssueLink
	if err := json.NewDecoder(resp.Body).Decode(&links); err != nil {
		return nil, fmt.Errorf("failed to decode issue links: %w", err)
	}

	return links, nil
}

func (c *Client) GetIssueActivities(ctx *YouTrackContext, issueID string) ([]*ActivityItem, error) {
	path := fmt.Sprintf("/api/issues/%s/activities", issueID)

	query := url.Values{}
	query.Add("fields", "id,category(id),author(id,login,fullName,email),timestamp,targetMember,field(id,name),removed(id,name,text,fullName,login,markdown),added(id,name,text,fullName,login,markdown)")

	resp, err := c.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var activities []*ActivityItem
	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		return nil, fmt.Errorf("failed to decode activities: %w", err)
	}

	return activities, nil
}

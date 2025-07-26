package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) GetIssueComments(ctx *YouTrackContext, issueID string) ([]*IssueComment, error) {
	path := fmt.Sprintf("/api/issues/%s/comments", issueID)

	query := url.Values{}
	query.Add("fields", "id,text,created,updated,author(id,login,fullName,email)")

	resp, err := c.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var comments []*IssueComment
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, fmt.Errorf("failed to decode comments: %w", err)
	}

	return comments, nil
}

func (c *Client) AddIssueComment(ctx *YouTrackContext, issueID string, text string) (*IssueComment, error) {
	path := fmt.Sprintf("/api/issues/%s/comments", issueID)

	req := map[string]string{
		"text": text,
	}

	resp, err := c.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var comment IssueComment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return nil, fmt.Errorf("failed to decode comment: %w", err)
	}

	return &comment, nil
}

func (c *Client) UpdateIssueComment(ctx *YouTrackContext, issueID, commentID string, text string) (*IssueComment, error) {
	path := fmt.Sprintf("/api/issues/%s/comments/%s", issueID, commentID)

	req := map[string]string{
		"text": text,
	}

	resp, err := c.Post(ctx, path, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var comment IssueComment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return nil, fmt.Errorf("failed to decode comment: %w", err)
	}

	return &comment, nil
}

func (c *Client) DeleteIssueComment(ctx *YouTrackContext, issueID, commentID string) error {
	path := fmt.Sprintf("/api/issues/%s/comments/%s", issueID, commentID)

	resp, err := c.Delete(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

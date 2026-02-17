package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) GetIssueTags(ctx *YouTrackContext, issueID string) ([]*IssueTag, error) {
	path := fmt.Sprintf("/api/issues/%s/tags", issueID)

	query := url.Values{}
	query.Add("fields", "id,name,color")

	resp, err := c.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tags []*IssueTag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode tags: %w", err)
	}

	return tags, nil
}

func (c *Client) AddIssueTag(ctx *YouTrackContext, issueID string, tagID string) error {
	path := fmt.Sprintf("/api/issues/%s/tags", issueID)

	req := map[string]interface{}{
		"id": tagID,
	}

	resp, err := c.Post(ctx, path, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) RemoveIssueTag(ctx *YouTrackContext, issueID string, tagID string) error {
	path := fmt.Sprintf("/api/issues/%s/tags/%s", issueID, tagID)

	resp, err := c.Delete(ctx, path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) ListTags(ctx *YouTrackContext, skip, top int) ([]*Tag, error) {
	query := url.Values{}
	query.Add("$skip", fmt.Sprintf("%d", skip))
	query.Add("$top", fmt.Sprintf("%d", top))
	query.Add("fields", "id,name,color")

	resp, err := c.Get(ctx, "/api/tags", query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tags []*Tag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode tags: %w", err)
	}

	return tags, nil
}

func (c *Client) CreateTag(ctx *YouTrackContext, name string, color string) (*Tag, error) {
	req := map[string]interface{}{
		"name": name,
	}
	if color != "" {
		req["color"] = color
	}

	resp, err := c.Post(ctx, "/api/tags", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tag Tag
	if err := json.NewDecoder(resp.Body).Decode(&tag); err != nil {
		return nil, fmt.Errorf("failed to decode tag: %w", err)
	}

	return &tag, nil
}

func (c *Client) GetTagByName(ctx *YouTrackContext, name string) (*Tag, error) {
	query := url.Values{}
	query.Add("fields", "id,name,color")
	query.Add("query", name)

	resp, err := c.Get(ctx, "/api/tags", query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tags []*Tag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to decode tags: %w", err)
	}

	// Exact match since query may return partial matches
	lowercaseName := strings.ToLower(name)
	for _, tag := range tags {
		if strings.ToLower(tag.Name) == lowercaseName {
			return tag, nil
		}
	}

	return nil, fmt.Errorf("tag with name '%s' not found", name)
}

func (c *Client) EnsureTag(ctx *YouTrackContext, name string, color string) (string, error) {
	tag, err := c.GetTagByName(ctx, name)
	if err == nil {
		return tag.ID, nil
	}

	newTag, err := c.CreateTag(ctx, name, color)
	if err != nil {
		return "", fmt.Errorf("failed to create tag '%s': %w", name, err)
	}

	return newTag.ID, nil
}

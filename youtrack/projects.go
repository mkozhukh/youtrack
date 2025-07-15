package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) GetProject(ctx *YouTrackContext, projectID string) (*Project, error) {
	path := fmt.Sprintf("/api/admin/projects/%s", projectID)
	
	query := url.Values{}
	query.Add("fields", "id,name,shortName,description")
	
	resp, err := c.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var project Project
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode project: %w", err)
	}

	return &project, nil
}

func (c *Client) ListProjects(ctx *YouTrackContext, skip, top int) ([]*Project, error) {
	query := url.Values{}
	query.Add("$skip", fmt.Sprintf("%d", skip))
	query.Add("$top", fmt.Sprintf("%d", top))
	query.Add("fields", "id,name,shortName,description")

	resp, err := c.Get(ctx, "/api/admin/projects", query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var projects []*Project
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to decode projects: %w", err)
	}

	return projects, nil
}
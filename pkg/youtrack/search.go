package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func (c *Client) SearchIssuesSorted(ctx *YouTrackContext, query string, skip, top int, sortBy, sortOrder string) ([]*Issue, error) {
	fullQuery := query
	if sortBy != "" {
		fullQuery = fmt.Sprintf("%s sort by: %s %s", query, sortBy, sortOrder)
	}

	params := url.Values{}
	params.Add("query", fullQuery)
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

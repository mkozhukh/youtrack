package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) GetCurrentUser(ctx *YouTrackContext) (*User, error) {
	query := url.Values{}
	query.Add("fields", "id,login,fullName,email")

	resp, err := c.Get(ctx, "/api/users/me", query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode current user: %w", err)
	}

	return &user, nil
}

func (c *Client) GetUser(ctx *YouTrackContext, userID string) (*User, error) {
	path := fmt.Sprintf("/api/users/%s", userID)

	query := url.Values{}
	query.Add("fields", "id,login,fullName,email")

	resp, err := c.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

func (c *Client) SearchUsers(ctx *YouTrackContext, query string, skip, top int) ([]*User, error) {
	params := url.Values{}
	params.Add("query", query)
	params.Add("$skip", fmt.Sprintf("%d", skip))
	params.Add("$top", fmt.Sprintf("%d", top))
	params.Add("fields", "id,login,fullName,email")

	resp, err := c.Get(ctx, "/api/users", params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var users []*User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

func (c *Client) GetUserByLogin(ctx *YouTrackContext, login string) (*User, error) {
	users, err := c.SearchUsers(ctx, fmt.Sprintf("login:%s", login), 0, 1)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user with login '%s' not found", login)
	}

	return users[0], nil
}

func (c *Client) GetProjectUsers(ctx *YouTrackContext, projectID string, skip, top int) ([]*User, error) {
	path := fmt.Sprintf("/api/admin/projects/%s/users", projectID)

	params := url.Values{}
	params.Add("$skip", fmt.Sprintf("%d", skip))
	params.Add("$top", fmt.Sprintf("%d", top))
	params.Add("fields", "id,login,fullName,email")

	resp, err := c.Get(ctx, path, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var users []*User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode project users: %w", err)
	}

	return users, nil
}

func (c *Client) SuggestUserByProject(ctx *YouTrackContext, projectID string, username string) (*User, error) {
	if username == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	lowercaseUsername := strings.ToLower(username)

	// Get all users for the project with pagination
	skip := 0
	top := 100

	for {
		users, err := c.GetProjectUsers(ctx, projectID, skip, top)
		if err != nil {
			return nil, fmt.Errorf("failed to get project users: %w", err)
		}

		if len(users) == 0 {
			break
		}

		// Search for matching user
		for _, user := range users {
			// Check if username matches any of the user fields (case-insensitive)
			if strings.Contains(strings.ToLower(user.Login), lowercaseUsername) ||
				strings.Contains(strings.ToLower(user.FullName), lowercaseUsername) ||
				strings.Contains(strings.ToLower(user.Email), lowercaseUsername) {
				return user, nil
			}
		}

		// If we got fewer results than requested, we've reached the end
		if len(users) < top {
			break
		}

		skip += len(users)
	}

	return nil, fmt.Errorf("no user found matching '%s' in project '%s'", username, projectID)
}

func (c *Client) GetUserWorklogs(ctx *YouTrackContext, userID string, projectID string, startDate, endDate string, skip, top int) ([]*WorkItem, error) {
	path := fmt.Sprintf("/api/users/%s/timeTracking/workItems", userID)

	params := url.Values{}
	params.Add("$skip", fmt.Sprintf("%d", skip))
	params.Add("$top", fmt.Sprintf("%d", top))
	params.Add("fields", "id,date,duration,text,author(id,login,fullName,email),type(id,name),issue(idReadable,summary)")

	if projectID != "" {
		params.Add("project", projectID)
	}
	if startDate != "" {
		params.Add("start", startDate)
	}
	if endDate != "" {
		params.Add("end", endDate)
	}

	resp, err := c.Get(ctx, path, params)
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

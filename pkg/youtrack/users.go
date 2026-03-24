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
	// Step 1: Get the project's ringId (Hub entity ID)
	ringID, err := c.getProjectRingID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project ringId: %w", err)
	}

	// Step 2: Fetch team users via Hub REST API
	path := fmt.Sprintf("/hub/api/rest/projects/%s/team/users", ringID)

	params := url.Values{}
	params.Add("$skip", fmt.Sprintf("%d", skip))
	params.Add("$top", fmt.Sprintf("%d", top))
	params.Add("fields", "id,login,name,profile(email(email))")

	resp, err := c.hubGet(ctx, path, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var hubResp struct {
		Users []struct {
			ID      string `json:"id"`
			Login   string `json:"login"`
			Name    string `json:"name"`
			Profile struct {
				Email struct {
					Email string `json:"email"`
				} `json:"email"`
			} `json:"profile"`
		} `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&hubResp); err != nil {
		return nil, fmt.Errorf("failed to decode project users: %w", err)
	}

	users := make([]*User, len(hubResp.Users))
	for i, hu := range hubResp.Users {
		users[i] = &User{
			ID:       hu.ID,
			Login:    hu.Login,
			FullName: hu.Name,
			Email:    hu.Profile.Email.Email,
		}
	}

	return users, nil
}

// getProjectRingID retrieves the Hub entity ID (ringId) for a YouTrack project.
func (c *Client) getProjectRingID(ctx *YouTrackContext, projectID string) (string, error) {
	path := fmt.Sprintf("/api/admin/projects/%s", projectID)

	params := url.Values{}
	params.Add("fields", "ringId")

	resp, err := c.Get(ctx, path, params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		RingID string `json:"ringId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode project ringId: %w", err)
	}

	if result.RingID == "" {
		return "", fmt.Errorf("project '%s' has no ringId", projectID)
	}

	return result.RingID, nil
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
	params := url.Values{}
	params.Add("$skip", fmt.Sprintf("%d", skip))
	params.Add("$top", fmt.Sprintf("%d", top))
	params.Add("fields", "id,date,duration(minutes,presentation),text,author(id,login,fullName,email),type(id,name),issue(idReadable,summary)")
	params.Add("author", userID)

	if projectID != "" {
		params.Add("query", fmt.Sprintf("project:{%s}", projectID))
	}
	if startDate != "" {
		params.Add("startDate", startDate)
	}
	if endDate != "" {
		params.Add("endDate", endDate)
	}

	resp, err := c.Get(ctx, "/api/workItems", params)
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

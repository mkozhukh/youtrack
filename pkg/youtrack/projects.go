package youtrack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
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

func (c *Client) GetProjectByName(ctx *YouTrackContext, name string) (*Project, error) {
	lowercaseName := strings.ToLower(name)

	skip := 0
	top := 100

	for {
		projects, err := c.ListProjects(ctx, skip, top)
		if err != nil {
			return nil, fmt.Errorf("failed to list projects: %w", err)
		}

		if len(projects) == 0 {
			break
		}

		for _, project := range projects {
			if strings.ToLower(project.Name) == lowercaseName || strings.ToLower(project.ShortName) == lowercaseName {
				return project, nil
			}
		}

		if len(projects) < top {
			break
		}

		skip += len(projects)
	}

	return nil, fmt.Errorf("project with name '%s' not found", name)
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

func (c *Client) GetProjectIssues(ctx *YouTrackContext, projectID string, skip, top int) ([]*Issue, error) {
	params := url.Values{}
	params.Add("query", fmt.Sprintf("project:{%s}", projectID))
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
		return nil, fmt.Errorf("failed to decode project issues: %w", err)
	}

	return issues, nil
}

func (c *Client) GetProjectCustomFields(ctx *YouTrackContext, projectID string) ([]*CustomField, error) {
	path := fmt.Sprintf("/api/admin/projects/%s/customFields", projectID)

	query := url.Values{}
	query.Add("fields", "field(id,name,$type)")

	resp, err := c.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// The API returns objects with a nested "field" property
	var rawFields []struct {
		Field *CustomField `json:"field"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rawFields); err != nil {
		return nil, fmt.Errorf("failed to decode project custom fields: %w", err)
	}

	var fields []*CustomField
	for _, rf := range rawFields {
		if rf.Field != nil {
			fields = append(fields, rf.Field)
		}
	}

	return fields, nil
}

func (c *Client) GetCustomFieldAllowedValues(ctx *YouTrackContext, projectID string, fieldName string) ([]AllowedValue, error) {
	path := fmt.Sprintf("/api/admin/projects/%s/customFields", projectID)

	query := url.Values{}
	query.Add("fields", "field(id,name,$type),bundle(id)")

	resp, err := c.Get(ctx, path, query)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rawFields []struct {
		Field  *CustomField `json:"field"`
		Bundle *struct {
			ID string `json:"id"`
		} `json:"bundle"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rawFields); err != nil {
		return nil, fmt.Errorf("failed to decode project custom fields: %w", err)
	}

	// Find the matching field
	for _, rf := range rawFields {
		if rf.Field == nil || !strings.EqualFold(rf.Field.Name, fieldName) {
			continue
		}

		if rf.Bundle == nil {
			return nil, fmt.Errorf("field '%s' has no associated bundle", fieldName)
		}

		// Determine bundle type from field type
		bundleType := "enum"
		fieldType := rf.Field.Type
		switch {
		case strings.Contains(fieldType, "State"):
			bundleType = "state"
		case strings.Contains(fieldType, "Owned"):
			bundleType = "ownedField"
		case strings.Contains(fieldType, "Enum"):
			bundleType = "enum"
		case strings.Contains(fieldType, "Version"):
			bundleType = "version"
		case strings.Contains(fieldType, "Build"):
			bundleType = "build"
		}

		// GET bundle values
		bundlePath := fmt.Sprintf("/api/admin/customFieldSettings/bundles/%s/%s/values", bundleType, rf.Bundle.ID)

		bundleQuery := url.Values{}
		bundleQuery.Add("fields", "id,name")

		bundleResp, err := c.Get(ctx, bundlePath, bundleQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to get bundle values: %w", err)
		}
		defer bundleResp.Body.Close()

		var values []AllowedValue
		if err := json.NewDecoder(bundleResp.Body).Decode(&values); err != nil {
			return nil, fmt.Errorf("failed to decode allowed values: %w", err)
		}

		return values, nil
	}

	return nil, fmt.Errorf("custom field '%s' not found in project '%s'", fieldName, projectID)
}

func (c *Client) AddCustomFieldEnumValue(ctx *YouTrackContext, projectID string, fieldName string, valueName string, color string) error {
	path := fmt.Sprintf("/api/admin/projects/%s/customFields", projectID)

	query := url.Values{}
	query.Add("fields", "field(id,name,$type),bundle(id)")

	resp, err := c.Get(ctx, path, query)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var rawFields []struct {
		Field  *CustomField `json:"field"`
		Bundle *struct {
			ID string `json:"id"`
		} `json:"bundle"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rawFields); err != nil {
		return fmt.Errorf("failed to decode project custom fields: %w", err)
	}

	for _, rf := range rawFields {
		if rf.Field == nil || !strings.EqualFold(rf.Field.Name, fieldName) {
			continue
		}

		if rf.Bundle == nil {
			return fmt.Errorf("field '%s' has no associated bundle", fieldName)
		}

		postPath := fmt.Sprintf("/api/admin/customFieldSettings/bundles/enum/%s/values", rf.Bundle.ID)

		body := map[string]interface{}{
			"name": valueName,
		}
		if color != "" {
			body["color"] = map[string]string{"id": color}
		}

		postResp, err := c.Post(ctx, postPath, body)
		if err != nil {
			return fmt.Errorf("failed to add enum value: %w", err)
		}
		defer postResp.Body.Close()

		return nil
	}

	return fmt.Errorf("custom field '%s' not found in project '%s'", fieldName, projectID)
}

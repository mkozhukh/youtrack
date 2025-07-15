package youtrack

import "time"

type Issue struct {
	ID          string                 `json:"id"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description,omitempty"`
	Created     time.Time              `json:"created"`
	Updated     time.Time              `json:"updated"`
	Resolved    *time.Time             `json:"resolved,omitempty"`
	Reporter    *User                  `json:"reporter,omitempty"`
	UpdatedBy   *User                  `json:"updatedBy,omitempty"`
	Assignee    *User                  `json:"assignee,omitempty"`
	Tags        []*IssueTag            `json:"tags,omitempty"`
	Fields      map[string]interface{} `json:"customFields,omitempty"`
}

type User struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	FullName string `json:"fullName,omitempty"`
	Email    string `json:"email,omitempty"`
}

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ShortName   string `json:"shortName"`
	Description string `json:"description,omitempty"`
}

type IssueComment struct {
	ID      string    `json:"id"`
	Author  *User     `json:"author,omitempty"`
	Text    string    `json:"text"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type CreateIssueRequest struct {
	Project     string                 `json:"project"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description,omitempty"`
	Fields      map[string]interface{} `json:"customFields,omitempty"`
}

type UpdateIssueRequest struct {
	Summary     *string                `json:"summary,omitempty"`
	Description *string                `json:"description,omitempty"`
	Assignee    *string                `json:"assignee,omitempty"`
	Fields      map[string]interface{} `json:"customFields,omitempty"`
}

type SearchIssuesRequest struct {
	Query  string   `json:"query"`
	Fields []string `json:"fields,omitempty"`
	Skip   int      `json:"$skip,omitempty"`
	Top    int      `json:"$top,omitempty"`
}

type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

type IssueTag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}
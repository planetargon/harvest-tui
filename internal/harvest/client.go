package harvest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

const (
	DefaultBaseURL = "https://api.harvestapp.com"
	UserAgent      = "harvest-tui (github.com/planetargon/harvest-tui)"
)

// User represents a Harvest user returned by the API.
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// ProjectClient represents a client associated with a project.
type ProjectClient struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Project represents a Harvest project returned by the API.
type Project struct {
	ID     int           `json:"id"`
	Name   string        `json:"name"`
	Client ProjectClient `json:"client"`
}

// projectsResponse represents the paginated response from GET /v2/projects.
type projectsResponse struct {
	Projects     []Project `json:"projects"`
	PerPage      int       `json:"per_page"`
	TotalPages   int       `json:"total_pages"`
	TotalEntries int       `json:"total_entries"`
	Page         int       `json:"page"`
	NextPage     *int      `json:"next_page"`
}

// TaskAssignmentProject represents project info within a task assignment.
type TaskAssignmentProject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TaskAssignmentTask represents task info within a task assignment.
type TaskAssignmentTask struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TaskAssignment represents a task assignment from the Harvest API.
type TaskAssignment struct {
	ID       int                   `json:"id"`
	Project  TaskAssignmentProject `json:"project"`
	Task     TaskAssignmentTask    `json:"task"`
	IsActive bool                  `json:"is_active"`
	Billable bool                  `json:"billable"`
}

// taskAssignmentsResponse represents the paginated response from GET /v2/task_assignments.
type taskAssignmentsResponse struct {
	TaskAssignments []TaskAssignment `json:"task_assignments"`
	PerPage         int              `json:"per_page"`
	TotalPages      int              `json:"total_pages"`
	TotalEntries    int              `json:"total_entries"`
	Page            int              `json:"page"`
	NextPage        *int             `json:"next_page"`
}

// TimeEntryClient represents client info within a time entry.
type TimeEntryClient struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TimeEntryProject represents project info within a time entry.
type TimeEntryProject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TimeEntryTask represents task info within a time entry.
type TimeEntryTask struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TimeEntry represents a time entry from the Harvest API.
type TimeEntry struct {
	ID         int              `json:"id"`
	SpentDate  string           `json:"spent_date"`
	Hours      float64          `json:"hours"`
	Notes      string           `json:"notes"`
	IsRunning  bool             `json:"is_running"`
	IsLocked   bool             `json:"is_locked"`
	IsBillable bool             `json:"billable"`
	Client     TimeEntryClient  `json:"client"`
	Project    TimeEntryProject `json:"project"`
	Task       TimeEntryTask    `json:"task"`
}

// timeEntriesResponse represents the paginated response from GET /v2/time_entries.
type timeEntriesResponse struct {
	TimeEntries  []TimeEntry `json:"time_entries"`
	PerPage      int         `json:"per_page"`
	TotalPages   int         `json:"total_pages"`
	TotalEntries int         `json:"total_entries"`
	Page         int         `json:"page"`
	NextPage     *int        `json:"next_page"`
}

// Task represents a task available for time tracking.
type Task struct {
	ID   int
	Name string
}

// ProjectWithTasks combines a project with its available tasks.
type ProjectWithTasks struct {
	Project Project
	Tasks   []Task
}

// CreateTimeEntryRequest represents the request payload for creating a time entry.
type CreateTimeEntryRequest struct {
	ProjectID  int     `json:"project_id"`
	TaskID     int     `json:"task_id"`
	SpentDate  string  `json:"spent_date"`
	Hours      float64 `json:"hours"`
	Notes      string  `json:"notes"`
	IsBillable *bool   `json:"billable,omitempty"`
}

// UpdateTimeEntryRequest represents the request payload for updating a time entry.
type UpdateTimeEntryRequest struct {
	ProjectID  *int     `json:"project_id,omitempty"`
	TaskID     *int     `json:"task_id,omitempty"`
	SpentDate  *string  `json:"spent_date,omitempty"`
	Hours      *float64 `json:"hours,omitempty"`
	Notes      *string  `json:"notes,omitempty"`
	IsBillable *bool    `json:"billable,omitempty"`
}

// AggregateProjectsWithTasks combines projects and task assignments into a sorted list.
// Projects without tasks are excluded. Results are sorted by client name, then project name.
func AggregateProjectsWithTasks(projects []Project, taskAssignments []TaskAssignment) []ProjectWithTasks {
	// Build a map of project ID to tasks
	tasksByProject := make(map[int][]Task)
	for _, ta := range taskAssignments {
		tasksByProject[ta.Project.ID] = append(tasksByProject[ta.Project.ID], Task{
			ID:   ta.Task.ID,
			Name: ta.Task.Name,
		})
	}

	// Build result with only projects that have tasks
	var result []ProjectWithTasks
	for _, p := range projects {
		if tasks, ok := tasksByProject[p.ID]; ok && len(tasks) > 0 {
			result = append(result, ProjectWithTasks{
				Project: p,
				Tasks:   tasks,
			})
		}
	}

	// Sort by client name, then project name
	sort.Slice(result, func(i, j int) bool {
		if result[i].Project.Client.Name != result[j].Project.Client.Name {
			return result[i].Project.Client.Name < result[j].Project.Client.Name
		}
		return result[i].Project.Name < result[j].Project.Name
	})

	return result
}

// Client is an HTTP client for the Harvest API v2.
type Client struct {
	baseURL     string
	accountID   string
	accessToken string
	httpClient  *http.Client
	userID      int // ID of the authenticated user
}

// NewClient creates a new Harvest API client with the given credentials.
func NewClient(accountID, accessToken string) *Client {
	return &Client{
		baseURL:     DefaultBaseURL,
		accountID:   accountID,
		accessToken: accessToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userID: 0, // Will be set when ValidateAuth is called
	}
}

// SetBaseURL sets a custom base URL for the client (useful for testing).
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// SetHTTPClient sets a custom HTTP client (useful for testing).
func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// SetUserID sets the user ID (useful for testing).
func (c *Client) SetUserID(userID int) {
	c.userID = userID
}

// GetUserID returns the current user ID.
func (c *Client) GetUserID() int {
	return c.userID
}

// Get performs a GET request to the specified path.
func (c *Client) Get(path string) (*http.Response, error) {
	return c.doRequest(http.MethodGet, path, nil)
}

// Post performs a POST request to the specified path with the given body.
func (c *Client) Post(path string, body interface{}) (*http.Response, error) {
	return c.doRequest(http.MethodPost, path, body)
}

// Patch performs a PATCH request to the specified path with the given body.
func (c *Client) Patch(path string, body interface{}) (*http.Response, error) {
	return c.doRequest(http.MethodPatch, path, body)
}

// Delete performs a DELETE request to the specified path.
func (c *Client) Delete(path string) (*http.Response, error) {
	return c.doRequest(http.MethodDelete, path, nil)
}

// doRequest performs an HTTP request with the appropriate headers.
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req, body != nil)

	return c.httpClient.Do(req)
}

// setHeaders sets the required headers for Harvest API requests.
func (c *Client) setHeaders(req *http.Request, hasBody bool) {
	req.Header.Set("Harvest-Account-Id", c.accountID)
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("User-Agent", UserAgent)

	if hasBody {
		req.Header.Set("Content-Type", "application/json")
	}
}

// ValidateAuth validates the credentials by calling GET /v2/users/me.
// Returns the authenticated user on success and stores the user ID.
// API Reference: https://help.getharvest.com/api-v2/users-api/users/users/#retrieve-the-currently-authenticated-user
func (c *Client) ValidateAuth() (*User, error) {
	resp, err := c.Get("/v2/users/me")
	if err != nil {
		return nil, fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed with status %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Store the user ID for filtering time entries
	c.userID = user.ID

	return &user, nil
}

// FetchProjects retrieves all active projects the user has access to.
// Handles pagination automatically.
// API Reference: https://help.getharvest.com/api-v2/projects-api/projects/projects/
func (c *Client) FetchProjects() ([]Project, error) {
	var allProjects []Project
	page := 1

	for {
		path := fmt.Sprintf("/v2/projects?is_active=true&page=%d", page)
		resp, err := c.Get(path)
		if err != nil {
			return nil, fmt.Errorf("network request failed: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to fetch projects with status %d", resp.StatusCode)
		}

		var projectsResp projectsResponse
		if err := json.NewDecoder(resp.Body).Decode(&projectsResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		resp.Body.Close()

		allProjects = append(allProjects, projectsResp.Projects...)

		// Check for more pages
		if projectsResp.NextPage == nil {
			break
		}
		page = *projectsResp.NextPage
	}

	return allProjects, nil
}

// FetchTaskAssignments retrieves all active task assignments.
// Handles pagination automatically.
// API Reference: https://help.getharvest.com/api-v2/projects-api/projects/task-assignments/
func (c *Client) FetchTaskAssignments() ([]TaskAssignment, error) {
	var allTaskAssignments []TaskAssignment
	page := 1

	for {
		path := fmt.Sprintf("/v2/task_assignments?is_active=true&page=%d", page)
		resp, err := c.Get(path)
		if err != nil {
			return nil, fmt.Errorf("network request failed: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to fetch task assignments with status %d", resp.StatusCode)
		}

		var taskAssignmentsResp taskAssignmentsResponse
		if err := json.NewDecoder(resp.Body).Decode(&taskAssignmentsResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		resp.Body.Close()

		allTaskAssignments = append(allTaskAssignments, taskAssignmentsResp.TaskAssignments...)

		// Check for more pages
		if taskAssignmentsResp.NextPage == nil {
			break
		}
		page = *taskAssignmentsResp.NextPage
	}

	return allTaskAssignments, nil
}

// FetchTimeEntries retrieves all time entries for a specific date.
// The date parameter should be in YYYY-MM-DD format.
// Handles pagination automatically.
// API Reference: https://help.getharvest.com/api-v2/timesheets-api/timesheets/time-entries/
func (c *Client) FetchTimeEntries(date string) ([]TimeEntry, error) {
	var allTimeEntries []TimeEntry
	page := 1

	for {
		// Filter by user_id to only get current user's entries
		path := fmt.Sprintf("/v2/time_entries?from=%s&to=%s&user_id=%d&page=%d", date, date, c.userID, page)
		resp, err := c.Get(path)
		if err != nil {
			return nil, fmt.Errorf("network request failed: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to fetch time entries with status %d", resp.StatusCode)
		}

		var timeEntriesResp timeEntriesResponse
		if err := json.NewDecoder(resp.Body).Decode(&timeEntriesResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		resp.Body.Close()

		allTimeEntries = append(allTimeEntries, timeEntriesResp.TimeEntries...)

		// Check for more pages
		if timeEntriesResp.NextPage == nil {
			break
		}
		page = *timeEntriesResp.NextPage
	}

	return allTimeEntries, nil
}

// CreateTimeEntry creates a new time entry in Harvest.
// API Reference: https://help.getharvest.com/api-v2/timesheets-api/timesheets/time-entries/
func (c *Client) CreateTimeEntry(request CreateTimeEntryRequest) (*TimeEntry, error) {
	resp, err := c.Post("/v2/time_entries", request)
	if err != nil {
		return nil, fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create time entry with status %d", resp.StatusCode)
	}

	var timeEntry TimeEntry
	if err := json.NewDecoder(resp.Body).Decode(&timeEntry); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &timeEntry, nil
}

// UpdateTimeEntry updates an existing time entry in Harvest.
// API Reference: https://help.getharvest.com/api-v2/timesheets-api/timesheets/time-entries/
func (c *Client) UpdateTimeEntry(id int, request UpdateTimeEntryRequest) (*TimeEntry, error) {
	path := fmt.Sprintf("/v2/time_entries/%d", id)
	resp, err := c.Patch(path, request)
	if err != nil {
		return nil, fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update time entry with status %d", resp.StatusCode)
	}

	var timeEntry TimeEntry
	if err := json.NewDecoder(resp.Body).Decode(&timeEntry); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &timeEntry, nil
}

// DeleteTimeEntry deletes an existing time entry in Harvest.
// API Reference: https://help.getharvest.com/api-v2/timesheets-api/timesheets/time-entries/
func (c *Client) DeleteTimeEntry(id int) error {
	path := fmt.Sprintf("/v2/time_entries/%d", id)
	resp, err := c.Delete(path)
	if err != nil {
		return fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete time entry with status %d", resp.StatusCode)
	}

	return nil
}

// RestartTimeEntry restarts (starts the timer for) an existing time entry in Harvest.
// API Reference: https://help.getharvest.com/api-v2/timesheets-api/timesheets/time-entries/
func (c *Client) RestartTimeEntry(id int) (*TimeEntry, error) {
	path := fmt.Sprintf("/v2/time_entries/%d/restart", id)
	resp, err := c.Patch(path, nil)
	if err != nil {
		return nil, fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to restart time entry with status %d", resp.StatusCode)
	}

	var timeEntry TimeEntry
	if err := json.NewDecoder(resp.Body).Decode(&timeEntry); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &timeEntry, nil
}

// StopTimeEntry stops the timer for an existing time entry in Harvest.
// API Reference: https://help.getharvest.com/api-v2/timesheets-api/timesheets/time-entries/
func (c *Client) StopTimeEntry(id int) (*TimeEntry, error) {
	path := fmt.Sprintf("/v2/time_entries/%d/stop", id)
	resp, err := c.Patch(path, nil)
	if err != nil {
		return nil, fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to stop time entry with status %d", resp.StatusCode)
	}

	var timeEntry TimeEntry
	if err := json.NewDecoder(resp.Body).Decode(&timeEntry); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &timeEntry, nil
}

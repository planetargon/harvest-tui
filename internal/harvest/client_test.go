package harvest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestValidateAuth(t *testing.T) {
	t.Run("given valid credentials when ValidateAuth called then returns user info without error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v2/users/me" {
				t.Errorf("expected path /v2/users/me, got %s", r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("expected method GET, got %s", r.Method)
			}

			// Verify headers
			if r.Header.Get("Harvest-Account-Id") != "12345" {
				t.Errorf("expected Harvest-Account-Id header 12345, got %s", r.Header.Get("Harvest-Account-Id"))
			}
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Errorf("expected Authorization header Bearer test-token, got %s", r.Header.Get("Authorization"))
			}
			if r.Header.Get("User-Agent") == "" {
				t.Error("expected User-Agent header to be set")
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         1,
				"first_name": "Test",
				"last_name":  "User",
				"email":      "test@example.com",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		user, err := client.ValidateAuth()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if user.ID != 1 {
			t.Errorf("expected user ID 1, got %d", user.ID)
		}
		if user.FirstName != "Test" {
			t.Errorf("expected first name Test, got %s", user.FirstName)
		}
		if user.LastName != "User" {
			t.Errorf("expected last name User, got %s", user.LastName)
		}
		if user.Email != "test@example.com" {
			t.Errorf("expected email test@example.com, got %s", user.Email)
		}
	})

	t.Run("given invalid credentials when ValidateAuth called then returns authentication error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":             "invalid_token",
				"error_description": "The access token is invalid",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "invalid-token")
		client.SetBaseURL(server.URL)

		user, err := client.ValidateAuth()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if user != nil {
			t.Errorf("expected nil user, got %v", user)
		}

		// Check that error indicates authentication failure
		if !strings.Contains(err.Error(), "authentication failed") && !strings.Contains(err.Error(), "Authentication failed") {
			t.Errorf("expected authentication failure error, got: %s", err.Error())
		}
	})

	t.Run("given rate limited response when ValidateAuth called then returns rate limit error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		user, err := client.ValidateAuth()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if user != nil {
			t.Errorf("expected nil user, got %v", user)
		}

		// Check that error indicates rate limiting
		if !strings.Contains(err.Error(), "429") && !strings.Contains(err.Error(), "rate") {
			t.Errorf("expected rate limit error, got: %s", err.Error())
		}
	})

	t.Run("given malformed JSON response when ValidateAuth called then returns parse error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{invalid json"))
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		user, err := client.ValidateAuth()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if user != nil {
			t.Errorf("expected nil user, got %v", user)
		}

		// Check that error indicates parsing failure
		if !strings.Contains(err.Error(), "parse") && !strings.Contains(err.Error(), "Parse") {
			t.Errorf("expected parse error, got: %s", err.Error())
		}
	})

	t.Run("given timeout when ValidateAuth called then returns network error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)
		client.SetHTTPClient(&http.Client{
			Timeout: 10 * time.Millisecond,
		})

		user, err := client.ValidateAuth()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if user != nil {
			t.Errorf("expected nil user, got %v", user)
		}

		// Check that error indicates network failure
		if !strings.Contains(err.Error(), "network") && !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "Timeout") {
			t.Errorf("expected network/timeout error, got: %s", err.Error())
		}
	})
}

func TestFetchProjects(t *testing.T) {
	t.Run("given valid response when FetchProjects called then returns projects with client data", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v2/projects" {
				t.Errorf("expected path /v2/projects, got %s", r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("expected method GET, got %s", r.Method)
			}
			// Verify is_active query param is set
			if r.URL.Query().Get("is_active") != "true" {
				t.Errorf("expected is_active=true query param, got %s", r.URL.Query().Get("is_active"))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"projects": []map[string]interface{}{
					{
						"id":   1,
						"name": "API Development",
						"client": map[string]interface{}{
							"id":   100,
							"name": "Acme Corp",
						},
					},
					{
						"id":   2,
						"name": "Mobile App",
						"client": map[string]interface{}{
							"id":   100,
							"name": "Acme Corp",
						},
					},
					{
						"id":   3,
						"name": "Consulting",
						"client": map[string]interface{}{
							"id":   200,
							"name": "BigCo Industries",
						},
					},
				},
				"per_page":      100,
				"total_pages":   1,
				"total_entries": 3,
				"page":          1,
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		projects, err := client.FetchProjects()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(projects) != 3 {
			t.Fatalf("expected 3 projects, got %d", len(projects))
		}

		// Verify first project
		if projects[0].ID != 1 {
			t.Errorf("expected project ID 1, got %d", projects[0].ID)
		}
		if projects[0].Name != "API Development" {
			t.Errorf("expected project name 'API Development', got '%s'", projects[0].Name)
		}
		if projects[0].Client.ID != 100 {
			t.Errorf("expected client ID 100, got %d", projects[0].Client.ID)
		}
		if projects[0].Client.Name != "Acme Corp" {
			t.Errorf("expected client name 'Acme Corp', got '%s'", projects[0].Client.Name)
		}

		// Verify third project has different client
		if projects[2].Client.ID != 200 {
			t.Errorf("expected client ID 200, got %d", projects[2].Client.ID)
		}
		if projects[2].Client.Name != "BigCo Industries" {
			t.Errorf("expected client name 'BigCo Industries', got '%s'", projects[2].Client.Name)
		}
	})

	t.Run("given empty projects response when FetchProjects called then returns empty slice", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"projects":      []interface{}{},
				"per_page":      100,
				"total_pages":   1,
				"total_entries": 0,
				"page":          1,
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		projects, err := client.FetchProjects()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(projects) != 0 {
			t.Errorf("expected 0 projects, got %d", len(projects))
		}
	})

	t.Run("given paginated response when FetchProjects called then fetches all pages", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			page := r.URL.Query().Get("page")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			if page == "" || page == "1" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"projects": []map[string]interface{}{
						{"id": 1, "name": "Project 1", "client": map[string]interface{}{"id": 1, "name": "Client 1"}},
						{"id": 2, "name": "Project 2", "client": map[string]interface{}{"id": 1, "name": "Client 1"}},
					},
					"per_page":      2,
					"total_pages":   2,
					"total_entries": 3,
					"page":          1,
					"next_page":     2,
				})
			} else {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"projects": []map[string]interface{}{
						{"id": 3, "name": "Project 3", "client": map[string]interface{}{"id": 2, "name": "Client 2"}},
					},
					"per_page":      2,
					"total_pages":   2,
					"total_entries": 3,
					"page":          2,
				})
			}
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		projects, err := client.FetchProjects()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(projects) != 3 {
			t.Errorf("expected 3 projects from pagination, got %d", len(projects))
		}

		if requestCount != 2 {
			t.Errorf("expected 2 requests for pagination, got %d", requestCount)
		}
	})
}

func TestFetchTaskAssignments(t *testing.T) {
	t.Run("given valid response when FetchTaskAssignments called then returns task assignments with project and task data", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v2/task_assignments" {
				t.Errorf("expected path /v2/task_assignments, got %s", r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("expected method GET, got %s", r.Method)
			}
			// Verify is_active query param is set
			if r.URL.Query().Get("is_active") != "true" {
				t.Errorf("expected is_active=true query param, got %s", r.URL.Query().Get("is_active"))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"task_assignments": []map[string]interface{}{
					{
						"id": 1,
						"project": map[string]interface{}{
							"id":   100,
							"name": "API Development",
						},
						"task": map[string]interface{}{
							"id":   1000,
							"name": "Code Review",
						},
						"is_active": true,
						"billable":  true,
					},
					{
						"id": 2,
						"project": map[string]interface{}{
							"id":   100,
							"name": "API Development",
						},
						"task": map[string]interface{}{
							"id":   1001,
							"name": "Development",
						},
						"is_active": true,
						"billable":  true,
					},
					{
						"id": 3,
						"project": map[string]interface{}{
							"id":   200,
							"name": "Mobile App",
						},
						"task": map[string]interface{}{
							"id":   1002,
							"name": "Testing",
						},
						"is_active": true,
						"billable":  false,
					},
				},
				"per_page":      100,
				"total_pages":   1,
				"total_entries": 3,
				"page":          1,
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		taskAssignments, err := client.FetchTaskAssignments()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(taskAssignments) != 3 {
			t.Fatalf("expected 3 task assignments, got %d", len(taskAssignments))
		}

		// Verify first task assignment
		if taskAssignments[0].ID != 1 {
			t.Errorf("expected task assignment ID 1, got %d", taskAssignments[0].ID)
		}
		if taskAssignments[0].Project.ID != 100 {
			t.Errorf("expected project ID 100, got %d", taskAssignments[0].Project.ID)
		}
		if taskAssignments[0].Project.Name != "API Development" {
			t.Errorf("expected project name 'API Development', got '%s'", taskAssignments[0].Project.Name)
		}
		if taskAssignments[0].Task.ID != 1000 {
			t.Errorf("expected task ID 1000, got %d", taskAssignments[0].Task.ID)
		}
		if taskAssignments[0].Task.Name != "Code Review" {
			t.Errorf("expected task name 'Code Review', got '%s'", taskAssignments[0].Task.Name)
		}
	})

	t.Run("given empty task assignments response when FetchTaskAssignments called then returns empty slice", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"task_assignments": []interface{}{},
				"per_page":         100,
				"total_pages":      1,
				"total_entries":    0,
				"page":             1,
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		taskAssignments, err := client.FetchTaskAssignments()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(taskAssignments) != 0 {
			t.Errorf("expected 0 task assignments, got %d", len(taskAssignments))
		}
	})

	t.Run("given paginated response when FetchTaskAssignments called then fetches all pages", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			page := r.URL.Query().Get("page")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			if page == "" || page == "1" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"task_assignments": []map[string]interface{}{
						{"id": 1, "project": map[string]interface{}{"id": 1, "name": "P1"}, "task": map[string]interface{}{"id": 1, "name": "T1"}},
						{"id": 2, "project": map[string]interface{}{"id": 1, "name": "P1"}, "task": map[string]interface{}{"id": 2, "name": "T2"}},
					},
					"per_page":      2,
					"total_pages":   2,
					"total_entries": 3,
					"page":          1,
					"next_page":     2,
				})
			} else {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"task_assignments": []map[string]interface{}{
						{"id": 3, "project": map[string]interface{}{"id": 2, "name": "P2"}, "task": map[string]interface{}{"id": 3, "name": "T3"}},
					},
					"per_page":      2,
					"total_pages":   2,
					"total_entries": 3,
					"page":          2,
				})
			}
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		taskAssignments, err := client.FetchTaskAssignments()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(taskAssignments) != 3 {
			t.Errorf("expected 3 task assignments from pagination, got %d", len(taskAssignments))
		}

		if requestCount != 2 {
			t.Errorf("expected 2 requests for pagination, got %d", requestCount)
		}
	})
}

func TestFetchTimeEntries(t *testing.T) {
	t.Run("given valid response when FetchTimeEntries called then returns time entries for date", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v2/time_entries" {
				t.Errorf("expected path /v2/time_entries, got %s", r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("expected method GET, got %s", r.Method)
			}
			// Verify from and to query params are set to the same date
			if r.URL.Query().Get("from") != "2025-01-15" {
				t.Errorf("expected from=2025-01-15, got %s", r.URL.Query().Get("from"))
			}
			if r.URL.Query().Get("to") != "2025-01-15" {
				t.Errorf("expected to=2025-01-15, got %s", r.URL.Query().Get("to"))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"time_entries": []map[string]interface{}{
					{
						"id":         1,
						"spent_date": "2025-01-15",
						"hours":      1.5,
						"notes":      "Code review",
						"is_running": false,
						"is_locked":  false,
						"billable":   true,
						"client": map[string]interface{}{
							"id":   100,
							"name": "Acme Corp",
						},
						"project": map[string]interface{}{
							"id":   200,
							"name": "API Development",
						},
						"task": map[string]interface{}{
							"id":   300,
							"name": "Code Review",
						},
					},
					{
						"id":         2,
						"spent_date": "2025-01-15",
						"hours":      2.0,
						"notes":      "Feature development",
						"is_running": true,
						"is_locked":  false,
						"billable":   true,
						"client": map[string]interface{}{
							"id":   100,
							"name": "Acme Corp",
						},
						"project": map[string]interface{}{
							"id":   201,
							"name": "Mobile App",
						},
						"task": map[string]interface{}{
							"id":   301,
							"name": "Development",
						},
					},
					{
						"id":         3,
						"spent_date": "2025-01-15",
						"hours":      0.75,
						"notes":      "Weekly sync",
						"is_running": false,
						"is_locked":  true,
						"billable":   false,
						"client": map[string]interface{}{
							"id":   101,
							"name": "BigCo Industries",
						},
						"project": map[string]interface{}{
							"id":   202,
							"name": "Consulting",
						},
						"task": map[string]interface{}{
							"id":   302,
							"name": "Meetings",
						},
					},
				},
				"per_page":      100,
				"total_pages":   1,
				"total_entries": 3,
				"page":          1,
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		entries, err := client.FetchTimeEntries("2025-01-15")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(entries) != 3 {
			t.Fatalf("expected 3 time entries, got %d", len(entries))
		}

		// Verify first entry
		if entries[0].ID != 1 {
			t.Errorf("expected entry ID 1, got %d", entries[0].ID)
		}
		if entries[0].Hours != 1.5 {
			t.Errorf("expected hours 1.5, got %f", entries[0].Hours)
		}
		if entries[0].Notes != "Code review" {
			t.Errorf("expected notes 'Code review', got '%s'", entries[0].Notes)
		}
		if entries[0].IsRunning != false {
			t.Errorf("expected IsRunning false, got true")
		}
		if entries[0].Client.Name != "Acme Corp" {
			t.Errorf("expected client name 'Acme Corp', got '%s'", entries[0].Client.Name)
		}
		if entries[0].Project.Name != "API Development" {
			t.Errorf("expected project name 'API Development', got '%s'", entries[0].Project.Name)
		}
		if entries[0].Task.Name != "Code Review" {
			t.Errorf("expected task name 'Code Review', got '%s'", entries[0].Task.Name)
		}

		// Verify second entry is running
		if entries[1].IsRunning != true {
			t.Errorf("expected entry 2 to be running")
		}

		// Verify third entry is locked
		if entries[2].IsLocked != true {
			t.Errorf("expected entry 3 to be locked")
		}
	})

	t.Run("given empty time entries response when FetchTimeEntries called then returns empty slice", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"time_entries":  []interface{}{},
				"per_page":      100,
				"total_pages":   1,
				"total_entries": 0,
				"page":          1,
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		entries, err := client.FetchTimeEntries("2025-01-15")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(entries) != 0 {
			t.Errorf("expected 0 time entries, got %d", len(entries))
		}
	})

	t.Run("given paginated response when FetchTimeEntries called then fetches all pages", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			page := r.URL.Query().Get("page")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			if page == "" || page == "1" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"time_entries": []map[string]interface{}{
						{"id": 1, "spent_date": "2025-01-15", "hours": 1.0, "client": map[string]interface{}{"id": 1, "name": "C1"}, "project": map[string]interface{}{"id": 1, "name": "P1"}, "task": map[string]interface{}{"id": 1, "name": "T1"}},
						{"id": 2, "spent_date": "2025-01-15", "hours": 1.0, "client": map[string]interface{}{"id": 1, "name": "C1"}, "project": map[string]interface{}{"id": 1, "name": "P1"}, "task": map[string]interface{}{"id": 1, "name": "T1"}},
					},
					"per_page":      2,
					"total_pages":   2,
					"total_entries": 3,
					"page":          1,
					"next_page":     2,
				})
			} else {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"time_entries": []map[string]interface{}{
						{"id": 3, "spent_date": "2025-01-15", "hours": 1.0, "client": map[string]interface{}{"id": 1, "name": "C1"}, "project": map[string]interface{}{"id": 1, "name": "P1"}, "task": map[string]interface{}{"id": 1, "name": "T1"}},
					},
					"per_page":      2,
					"total_pages":   2,
					"total_entries": 3,
					"page":          2,
				})
			}
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		entries, err := client.FetchTimeEntries("2025-01-15")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(entries) != 3 {
			t.Errorf("expected 3 time entries from pagination, got %d", len(entries))
		}

		if requestCount != 2 {
			t.Errorf("expected 2 requests for pagination, got %d", requestCount)
		}
	})
}

func TestCreateTimeEntry(t *testing.T) {
	t.Run("given valid time entry data when CreateTimeEntry called then creates entry and returns it", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v2/time_entries" {
				t.Errorf("expected path /v2/time_entries, got %s", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("expected method POST, got %s", r.Method)
			}

			// Verify Content-Type header
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			// Verify request body
			var reqData map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}

			expectedFields := map[string]interface{}{
				"project_id": float64(100),
				"task_id":    float64(200),
				"spent_date": "2025-01-15",
				"hours":      1.5,
				"notes":      "Code review session",
			}

			for field, expectedValue := range expectedFields {
				if reqData[field] != expectedValue {
					t.Errorf("expected %s=%v, got %v", field, expectedValue, reqData[field])
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         1001,
				"spent_date": "2025-01-15",
				"hours":      1.5,
				"notes":      "Code review session",
				"is_running": false,
				"is_locked":  false,
				"billable":   true,
				"client": map[string]interface{}{
					"id":   50,
					"name": "Test Client",
				},
				"project": map[string]interface{}{
					"id":   100,
					"name": "Test Project",
				},
				"task": map[string]interface{}{
					"id":   200,
					"name": "Development",
				},
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		entry := CreateTimeEntryRequest{
			ProjectID: 100,
			TaskID:    200,
			SpentDate: "2025-01-15",
			Hours:     1.5,
			Notes:     "Code review session",
		}

		created, err := client.CreateTimeEntry(entry)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if created.ID != 1001 {
			t.Errorf("expected ID 1001, got %d", created.ID)
		}
		if created.SpentDate != "2025-01-15" {
			t.Errorf("expected spent_date 2025-01-15, got %s", created.SpentDate)
		}
		if created.Hours != 1.5 {
			t.Errorf("expected hours 1.5, got %f", created.Hours)
		}
		if created.Notes != "Code review session" {
			t.Errorf("expected notes 'Code review session', got '%s'", created.Notes)
		}
		if created.Project.ID != 100 {
			t.Errorf("expected project ID 100, got %d", created.Project.ID)
		}
		if created.Task.ID != 200 {
			t.Errorf("expected task ID 200, got %d", created.Task.ID)
		}
	})

	t.Run("given time entry with billable field when CreateTimeEntry called then creates entry with correct billable flag", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqData map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}

			if reqData["billable"] != false {
				t.Errorf("expected billable=false, got %v", reqData["billable"])
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         1002,
				"spent_date": "2025-01-15",
				"hours":      2.0,
				"notes":      "Non-billable work",
				"billable":   false,
				"client":     map[string]interface{}{"id": 50, "name": "Test Client"},
				"project":    map[string]interface{}{"id": 100, "name": "Test Project"},
				"task":       map[string]interface{}{"id": 200, "name": "Development"},
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		entry := CreateTimeEntryRequest{
			ProjectID:  100,
			TaskID:     200,
			SpentDate:  "2025-01-15",
			Hours:      2.0,
			Notes:      "Non-billable work",
			IsBillable: &[]bool{false}[0],
		}

		created, err := client.CreateTimeEntry(entry)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if created.IsBillable != false {
			t.Errorf("expected billable false, got %t", created.IsBillable)
		}
	})

	t.Run("given invalid request when CreateTimeEntry called then returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "The project_id field is required.",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		entry := CreateTimeEntryRequest{
			TaskID:    200,
			SpentDate: "2025-01-15",
			Hours:     1.0,
		}

		created, err := client.CreateTimeEntry(entry)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if created != nil {
			t.Errorf("expected nil time entry, got %v", created)
		}

		if !strings.Contains(err.Error(), "400") {
			t.Errorf("expected 400 status in error, got: %s", err.Error())
		}
	})

	t.Run("given unauthorized request when CreateTimeEntry called then returns auth error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "invalid_token",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "invalid-token")
		client.SetBaseURL(server.URL)

		entry := CreateTimeEntryRequest{
			ProjectID: 100,
			TaskID:    200,
			SpentDate: "2025-01-15",
			Hours:     1.0,
		}

		created, err := client.CreateTimeEntry(entry)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if created != nil {
			t.Errorf("expected nil time entry, got %v", created)
		}

		if !strings.Contains(err.Error(), "401") {
			t.Errorf("expected 401 status in error, got: %s", err.Error())
		}
	})
}

func TestUpdateTimeEntry(t *testing.T) {
	t.Run("given valid update data when UpdateTimeEntry called then updates entry and returns it", func(t *testing.T) {
		entryID := 1001
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := fmt.Sprintf("/v2/time_entries/%d", entryID)
			if r.URL.Path != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodPatch {
				t.Errorf("expected method PATCH, got %s", r.Method)
			}

			// Verify Content-Type header
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}

			// Verify request body
			var reqData map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}

			expectedFields := map[string]interface{}{
				"hours": 2.5,
				"notes": "Updated notes",
			}

			for field, expectedValue := range expectedFields {
				if reqData[field] != expectedValue {
					t.Errorf("expected %s=%v, got %v", field, expectedValue, reqData[field])
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         1001,
				"spent_date": "2025-01-15",
				"hours":      2.5,
				"notes":      "Updated notes",
				"is_running": false,
				"is_locked":  false,
				"billable":   true,
				"client": map[string]interface{}{
					"id":   50,
					"name": "Test Client",
				},
				"project": map[string]interface{}{
					"id":   100,
					"name": "Test Project",
				},
				"task": map[string]interface{}{
					"id":   200,
					"name": "Development",
				},
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		update := UpdateTimeEntryRequest{
			Hours: &[]float64{2.5}[0],
			Notes: &[]string{"Updated notes"}[0],
		}

		updated, err := client.UpdateTimeEntry(entryID, update)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if updated.ID != 1001 {
			t.Errorf("expected ID 1001, got %d", updated.ID)
		}
		if updated.Hours != 2.5 {
			t.Errorf("expected hours 2.5, got %f", updated.Hours)
		}
		if updated.Notes != "Updated notes" {
			t.Errorf("expected notes 'Updated notes', got '%s'", updated.Notes)
		}
	})

	t.Run("given billable field when UpdateTimeEntry called then updates billable flag", func(t *testing.T) {
		entryID := 1001
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var reqData map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
				t.Fatalf("failed to decode request body: %v", err)
			}

			if reqData["billable"] != true {
				t.Errorf("expected billable=true, got %v", reqData["billable"])
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":       1001,
				"billable": true,
				"client":   map[string]interface{}{"id": 50, "name": "Test Client"},
				"project":  map[string]interface{}{"id": 100, "name": "Test Project"},
				"task":     map[string]interface{}{"id": 200, "name": "Development"},
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		update := UpdateTimeEntryRequest{
			IsBillable: &[]bool{true}[0],
		}

		updated, err := client.UpdateTimeEntry(entryID, update)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if updated.IsBillable != true {
			t.Errorf("expected billable true, got %t", updated.IsBillable)
		}
	})

	t.Run("given locked time entry when UpdateTimeEntry called then returns error", func(t *testing.T) {
		entryID := 1001
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Time entry is locked and cannot be modified.",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		update := UpdateTimeEntryRequest{
			Hours: &[]float64{2.0}[0],
		}

		updated, err := client.UpdateTimeEntry(entryID, update)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if updated != nil {
			t.Errorf("expected nil time entry, got %v", updated)
		}

		if !strings.Contains(err.Error(), "400") {
			t.Errorf("expected 400 status in error, got: %s", err.Error())
		}
	})

	t.Run("given nonexistent entry when UpdateTimeEntry called then returns not found error", func(t *testing.T) {
		entryID := 999999
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Time entry not found.",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		update := UpdateTimeEntryRequest{
			Hours: &[]float64{2.0}[0],
		}

		updated, err := client.UpdateTimeEntry(entryID, update)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if updated != nil {
			t.Errorf("expected nil time entry, got %v", updated)
		}

		if !strings.Contains(err.Error(), "404") {
			t.Errorf("expected 404 status in error, got: %s", err.Error())
		}
	})

	t.Run("given unauthorized request when UpdateTimeEntry called then returns auth error", func(t *testing.T) {
		entryID := 1001
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "invalid_token",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "invalid-token")
		client.SetBaseURL(server.URL)

		update := UpdateTimeEntryRequest{
			Hours: &[]float64{2.0}[0],
		}

		updated, err := client.UpdateTimeEntry(entryID, update)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if updated != nil {
			t.Errorf("expected nil time entry, got %v", updated)
		}

		if !strings.Contains(err.Error(), "401") {
			t.Errorf("expected 401 status in error, got: %s", err.Error())
		}
	})
}

func TestDeleteTimeEntry(t *testing.T) {
	t.Run("given existing time entry when DeleteTimeEntry called then deletes entry successfully", func(t *testing.T) {
		entryID := 1001
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := fmt.Sprintf("/v2/time_entries/%d", entryID)
			if r.URL.Path != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodDelete {
				t.Errorf("expected method DELETE, got %s", r.Method)
			}

			// Verify auth headers are present
			if r.Header.Get("Harvest-Account-Id") != "12345" {
				t.Errorf("expected Harvest-Account-Id header 12345, got %s", r.Header.Get("Harvest-Account-Id"))
			}
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Errorf("expected Authorization header Bearer test-token, got %s", r.Header.Get("Authorization"))
			}

			// DELETE typically returns 200 OK with empty response body
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		err := client.DeleteTimeEntry(entryID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("given locked time entry when DeleteTimeEntry called then returns error", func(t *testing.T) {
		entryID := 1001
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Time entry is locked and cannot be deleted.",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		err := client.DeleteTimeEntry(entryID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "400") {
			t.Errorf("expected 400 status in error, got: %s", err.Error())
		}
	})

	t.Run("given running time entry when DeleteTimeEntry called then returns error", func(t *testing.T) {
		entryID := 1001
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Cannot delete a running time entry.",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		err := client.DeleteTimeEntry(entryID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "400") {
			t.Errorf("expected 400 status in error, got: %s", err.Error())
		}
	})

	t.Run("given nonexistent entry when DeleteTimeEntry called then returns not found error", func(t *testing.T) {
		entryID := 999999
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Time entry not found.",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		err := client.DeleteTimeEntry(entryID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "404") {
			t.Errorf("expected 404 status in error, got: %s", err.Error())
		}
	})

	t.Run("given unauthorized request when DeleteTimeEntry called then returns auth error", func(t *testing.T) {
		entryID := 1001
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "invalid_token",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "invalid-token")
		client.SetBaseURL(server.URL)

		err := client.DeleteTimeEntry(entryID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "401") {
			t.Errorf("expected 401 status in error, got: %s", err.Error())
		}
	})
}

func TestRestartTimeEntry(t *testing.T) {
	t.Run("given stopped time entry when RestartTimeEntry called then starts timer and returns updated entry", func(t *testing.T) {
		entryID := 1001
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := fmt.Sprintf("/v2/time_entries/%d/restart", entryID)
			if r.URL.Path != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodPatch {
				t.Errorf("expected method PATCH, got %s", r.Method)
			}

			// Verify auth headers are present
			if r.Header.Get("Harvest-Account-Id") != "12345" {
				t.Errorf("expected Harvest-Account-Id header 12345, got %s", r.Header.Get("Harvest-Account-Id"))
			}
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Errorf("expected Authorization header Bearer test-token, got %s", r.Header.Get("Authorization"))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         1001,
				"spent_date": "2025-01-15",
				"hours":      2.5,
				"notes":      "Development work",
				"is_running": true,
				"is_locked":  false,
				"billable":   true,
				"client": map[string]interface{}{
					"id":   50,
					"name": "Test Client",
				},
				"project": map[string]interface{}{
					"id":   100,
					"name": "Test Project",
				},
				"task": map[string]interface{}{
					"id":   200,
					"name": "Development",
				},
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		entry, err := client.RestartTimeEntry(entryID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if entry.ID != 1001 {
			t.Errorf("expected ID 1001, got %d", entry.ID)
		}
		if entry.IsRunning != true {
			t.Errorf("expected IsRunning true, got %t", entry.IsRunning)
		}
	})

	t.Run("given locked time entry when RestartTimeEntry called then returns error", func(t *testing.T) {
		entryID := 1001
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Time entry is locked and cannot be restarted.",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		entry, err := client.RestartTimeEntry(entryID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if entry != nil {
			t.Errorf("expected nil time entry, got %v", entry)
		}

		if !strings.Contains(err.Error(), "400") {
			t.Errorf("expected 400 status in error, got: %s", err.Error())
		}
	})

	t.Run("given nonexistent entry when RestartTimeEntry called then returns not found error", func(t *testing.T) {
		entryID := 999999
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Time entry not found.",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		entry, err := client.RestartTimeEntry(entryID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if entry != nil {
			t.Errorf("expected nil time entry, got %v", entry)
		}

		if !strings.Contains(err.Error(), "404") {
			t.Errorf("expected 404 status in error, got: %s", err.Error())
		}
	})
}

func TestStopTimeEntry(t *testing.T) {
	t.Run("given running time entry when StopTimeEntry called then stops timer and returns updated entry", func(t *testing.T) {
		entryID := 1001
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := fmt.Sprintf("/v2/time_entries/%d/stop", entryID)
			if r.URL.Path != expectedPath {
				t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
			}
			if r.Method != http.MethodPatch {
				t.Errorf("expected method PATCH, got %s", r.Method)
			}

			// Verify auth headers are present
			if r.Header.Get("Harvest-Account-Id") != "12345" {
				t.Errorf("expected Harvest-Account-Id header 12345, got %s", r.Header.Get("Harvest-Account-Id"))
			}
			if r.Header.Get("Authorization") != "Bearer test-token" {
				t.Errorf("expected Authorization header Bearer test-token, got %s", r.Header.Get("Authorization"))
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":         1001,
				"spent_date": "2025-01-15",
				"hours":      3.25,
				"notes":      "Development work",
				"is_running": false,
				"is_locked":  false,
				"billable":   true,
				"client": map[string]interface{}{
					"id":   50,
					"name": "Test Client",
				},
				"project": map[string]interface{}{
					"id":   100,
					"name": "Test Project",
				},
				"task": map[string]interface{}{
					"id":   200,
					"name": "Development",
				},
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		entry, err := client.StopTimeEntry(entryID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if entry.ID != 1001 {
			t.Errorf("expected ID 1001, got %d", entry.ID)
		}
		if entry.IsRunning != false {
			t.Errorf("expected IsRunning false, got %t", entry.IsRunning)
		}
		if entry.Hours != 3.25 {
			t.Errorf("expected hours 3.25, got %f", entry.Hours)
		}
	})

	t.Run("given already stopped time entry when StopTimeEntry called then returns error", func(t *testing.T) {
		entryID := 1001
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Time entry is not running.",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		entry, err := client.StopTimeEntry(entryID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if entry != nil {
			t.Errorf("expected nil time entry, got %v", entry)
		}

		if !strings.Contains(err.Error(), "400") {
			t.Errorf("expected 400 status in error, got: %s", err.Error())
		}
	})

	t.Run("given locked time entry when StopTimeEntry called then returns error", func(t *testing.T) {
		entryID := 1001
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Time entry is locked and cannot be stopped.",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		entry, err := client.StopTimeEntry(entryID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if entry != nil {
			t.Errorf("expected nil time entry, got %v", entry)
		}

		if !strings.Contains(err.Error(), "400") {
			t.Errorf("expected 400 status in error, got: %s", err.Error())
		}
	})

	t.Run("given nonexistent entry when StopTimeEntry called then returns not found error", func(t *testing.T) {
		entryID := 999999
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Time entry not found.",
			})
		}))
		defer server.Close()

		client := NewClient("12345", "test-token")
		client.SetBaseURL(server.URL)

		entry, err := client.StopTimeEntry(entryID)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if entry != nil {
			t.Errorf("expected nil time entry, got %v", entry)
		}

		if !strings.Contains(err.Error(), "404") {
			t.Errorf("expected 404 status in error, got: %s", err.Error())
		}
	})
}

func TestAggregateProjectsWithTasks(t *testing.T) {
	t.Run("given projects and task assignments when aggregated then returns projects with their tasks", func(t *testing.T) {
		projects := []Project{
			{ID: 1, Name: "API Development", Client: ProjectClient{ID: 100, Name: "Acme Corp"}},
			{ID: 2, Name: "Mobile App", Client: ProjectClient{ID: 100, Name: "Acme Corp"}},
			{ID: 3, Name: "Consulting", Client: ProjectClient{ID: 200, Name: "BigCo Industries"}},
		}
		taskAssignments := []TaskAssignment{
			{ID: 1, Project: TaskAssignmentProject{ID: 1, Name: "API Development"}, Task: TaskAssignmentTask{ID: 10, Name: "Code Review"}},
			{ID: 2, Project: TaskAssignmentProject{ID: 1, Name: "API Development"}, Task: TaskAssignmentTask{ID: 11, Name: "Development"}},
			{ID: 3, Project: TaskAssignmentProject{ID: 2, Name: "Mobile App"}, Task: TaskAssignmentTask{ID: 12, Name: "Testing"}},
			{ID: 4, Project: TaskAssignmentProject{ID: 3, Name: "Consulting"}, Task: TaskAssignmentTask{ID: 13, Name: "Meetings"}},
		}

		result := AggregateProjectsWithTasks(projects, taskAssignments)

		if len(result) != 3 {
			t.Fatalf("expected 3 project entries, got %d", len(result))
		}

		// Verify first project has 2 tasks
		found := false
		for _, pe := range result {
			if pe.Project.ID == 1 {
				found = true
				if len(pe.Tasks) != 2 {
					t.Errorf("expected project 1 to have 2 tasks, got %d", len(pe.Tasks))
				}
			}
		}
		if !found {
			t.Error("expected to find project 1 in results")
		}

		// Verify project 2 has 1 task
		found = false
		for _, pe := range result {
			if pe.Project.ID == 2 {
				found = true
				if len(pe.Tasks) != 1 {
					t.Errorf("expected project 2 to have 1 task, got %d", len(pe.Tasks))
				}
			}
		}
		if !found {
			t.Error("expected to find project 2 in results")
		}
	})

	t.Run("given projects and task assignments when aggregated then results are sorted by client then project", func(t *testing.T) {
		projects := []Project{
			{ID: 3, Name: "Zebra Project", Client: ProjectClient{ID: 200, Name: "BigCo Industries"}},
			{ID: 1, Name: "API Development", Client: ProjectClient{ID: 100, Name: "Acme Corp"}},
			{ID: 2, Name: "Mobile App", Client: ProjectClient{ID: 100, Name: "Acme Corp"}},
		}
		taskAssignments := []TaskAssignment{
			{ID: 1, Project: TaskAssignmentProject{ID: 1, Name: "API Development"}, Task: TaskAssignmentTask{ID: 10, Name: "Task"}},
			{ID: 2, Project: TaskAssignmentProject{ID: 2, Name: "Mobile App"}, Task: TaskAssignmentTask{ID: 11, Name: "Task"}},
			{ID: 3, Project: TaskAssignmentProject{ID: 3, Name: "Zebra Project"}, Task: TaskAssignmentTask{ID: 12, Name: "Task"}},
		}

		result := AggregateProjectsWithTasks(projects, taskAssignments)

		if len(result) != 3 {
			t.Fatalf("expected 3 project entries, got %d", len(result))
		}

		// First should be Acme Corp - API Development
		if result[0].Project.Client.Name != "Acme Corp" {
			t.Errorf("expected first entry to be Acme Corp, got %s", result[0].Project.Client.Name)
		}
		if result[0].Project.Name != "API Development" {
			t.Errorf("expected first project name to be API Development, got %s", result[0].Project.Name)
		}

		// Second should be Acme Corp - Mobile App
		if result[1].Project.Client.Name != "Acme Corp" {
			t.Errorf("expected second entry to be Acme Corp, got %s", result[1].Project.Client.Name)
		}
		if result[1].Project.Name != "Mobile App" {
			t.Errorf("expected second project name to be Mobile App, got %s", result[1].Project.Name)
		}

		// Third should be BigCo Industries - Zebra Project
		if result[2].Project.Client.Name != "BigCo Industries" {
			t.Errorf("expected third entry to be BigCo Industries, got %s", result[2].Project.Client.Name)
		}
	})

	t.Run("given project with no task assignments when aggregated then project is excluded", func(t *testing.T) {
		projects := []Project{
			{ID: 1, Name: "API Development", Client: ProjectClient{ID: 100, Name: "Acme Corp"}},
			{ID: 2, Name: "Empty Project", Client: ProjectClient{ID: 100, Name: "Acme Corp"}},
		}
		taskAssignments := []TaskAssignment{
			{ID: 1, Project: TaskAssignmentProject{ID: 1, Name: "API Development"}, Task: TaskAssignmentTask{ID: 10, Name: "Task"}},
		}

		result := AggregateProjectsWithTasks(projects, taskAssignments)

		// Only project with tasks should be included
		if len(result) != 1 {
			t.Fatalf("expected 1 project entry, got %d", len(result))
		}
		if result[0].Project.ID != 1 {
			t.Errorf("expected project ID 1, got %d", result[0].Project.ID)
		}
	})

	t.Run("given empty inputs when aggregated then returns empty slice", func(t *testing.T) {
		result := AggregateProjectsWithTasks([]Project{}, []TaskAssignment{})

		if len(result) != 0 {
			t.Errorf("expected 0 project entries, got %d", len(result))
		}
	})
}

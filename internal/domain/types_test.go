package domain

import "testing"

func TestDurationFormatting(t *testing.T) {
	t.Run("given zero hours when formatted then returns 0:00", func(t *testing.T) {
		result := FormatDuration(0.0)
		expected := "0:00"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("given one hour when formatted then returns 1:00", func(t *testing.T) {
		result := FormatDuration(1.0)
		expected := "1:00"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("given one and a half hours when formatted then returns 1:30", func(t *testing.T) {
		result := FormatDuration(1.5)
		expected := "1:30"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("given quarter hour when formatted then returns 0:15", func(t *testing.T) {
		result := FormatDuration(0.25)
		expected := "0:15"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("given 2.75 hours when formatted then returns 2:45", func(t *testing.T) {
		result := FormatDuration(2.75)
		expected := "2:45"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("given 0.1 hours when formatted then returns 0:06", func(t *testing.T) {
		result := FormatDuration(0.1)
		expected := "0:06"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("given 10.5 hours when formatted then returns 10:30", func(t *testing.T) {
		result := FormatDuration(10.5)
		expected := "10:30"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})
}

func TestDurationParsing(t *testing.T) {
	t.Run("given valid duration 1:30 when parsed then returns 1.5 hours", func(t *testing.T) {
		result, err := ParseDuration("1:30")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := 1.5
		if result != expected {
			t.Errorf("expected %f, got %f", expected, result)
		}
	})

	t.Run("given valid duration 0:15 when parsed then returns 0.25 hours", func(t *testing.T) {
		result, err := ParseDuration("0:15")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := 0.25
		if result != expected {
			t.Errorf("expected %f, got %f", expected, result)
		}
	})

	t.Run("given valid duration 2:45 when parsed then returns 2.75 hours", func(t *testing.T) {
		result, err := ParseDuration("2:45")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := 2.75
		if result != expected {
			t.Errorf("expected %f, got %f", expected, result)
		}
	})

	t.Run("given valid duration 0:00 when parsed then returns 0 hours", func(t *testing.T) {
		result, err := ParseDuration("0:00")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := 0.0
		if result != expected {
			t.Errorf("expected %f, got %f", expected, result)
		}
	})

	t.Run("given empty string when parsed then returns error", func(t *testing.T) {
		_, err := ParseDuration("")
		if err == nil {
			t.Fatal("expected error for empty string")
		}
		if err.Error() != "duration cannot be empty" {
			t.Errorf("expected 'duration cannot be empty', got '%s'", err.Error())
		}
	})

	t.Run("given whitespace only when parsed then returns error", func(t *testing.T) {
		_, err := ParseDuration("   ")
		if err == nil {
			t.Fatal("expected error for whitespace only")
		}
		if err.Error() != "duration cannot be empty" {
			t.Errorf("expected 'duration cannot be empty', got '%s'", err.Error())
		}
	})

	t.Run("given invalid format abc when parsed then returns validation error", func(t *testing.T) {
		_, err := ParseDuration("abc")
		if err == nil {
			t.Fatal("expected error for invalid format")
		}
		expectedMsg := "invalid duration format. Use HH:MM (e.g., 1:30)"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("given invalid format 1:2:3 when parsed then returns validation error", func(t *testing.T) {
		_, err := ParseDuration("1:2:3")
		if err == nil {
			t.Fatal("expected error for too many parts")
		}
		expectedMsg := "invalid duration format. Use HH:MM (e.g., 1:30)"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("given invalid hours abc:30 when parsed then returns validation error", func(t *testing.T) {
		_, err := ParseDuration("abc:30")
		if err == nil {
			t.Fatal("expected error for non-numeric hours")
		}
		expectedMsg := "invalid duration format. Use HH:MM (e.g., 1:30)"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("given invalid minutes 1:abc when parsed then returns validation error", func(t *testing.T) {
		_, err := ParseDuration("1:abc")
		if err == nil {
			t.Fatal("expected error for non-numeric minutes")
		}
		expectedMsg := "invalid duration format. Use HH:MM (e.g., 1:30)"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("given negative hours -1:30 when parsed then returns validation error", func(t *testing.T) {
		_, err := ParseDuration("-1:30")
		if err == nil {
			t.Fatal("expected error for negative hours")
		}
		expectedMsg := "invalid duration format. Use HH:MM (e.g., 1:30)"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("given invalid minutes 1:60 when parsed then returns validation error", func(t *testing.T) {
		_, err := ParseDuration("1:60")
		if err == nil {
			t.Fatal("expected error for minutes >= 60")
		}
		expectedMsg := "invalid duration format. Use HH:MM (e.g., 1:30)"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}
	})

	t.Run("given negative minutes 1:-15 when parsed then returns validation error", func(t *testing.T) {
		_, err := ParseDuration("1:-15")
		if err == nil {
			t.Fatal("expected error for negative minutes")
		}
		expectedMsg := "invalid duration format. Use HH:MM (e.g., 1:30)"
		if err.Error() != expectedMsg {
			t.Errorf("expected '%s', got '%s'", expectedMsg, err.Error())
		}
	})
}

func TestClientSorting(t *testing.T) {
	t.Run("given unsorted clients when sorted then returns alphabetical order", func(t *testing.T) {
		clients := []Client{
			{ID: 1, Name: "Zebra Corp"},
			{ID: 2, Name: "Apple Inc"},
			{ID: 3, Name: "Microsoft"},
			{ID: 4, Name: "Amazon"},
		}

		sorted := SortClients(clients)

		expected := []string{"Amazon", "Apple Inc", "Microsoft", "Zebra Corp"}
		for i, client := range sorted {
			if client.Name != expected[i] {
				t.Errorf("expected client %d to be %s, got %s", i, expected[i], client.Name)
			}
		}
	})

	t.Run("given already sorted clients when sorted then maintains order", func(t *testing.T) {
		clients := []Client{
			{ID: 1, Name: "Amazon"},
			{ID: 2, Name: "Apple Inc"},
			{ID: 3, Name: "Microsoft"},
		}

		sorted := SortClients(clients)

		for i, client := range sorted {
			if client.Name != clients[i].Name {
				t.Errorf("expected client %d to be %s, got %s", i, clients[i].Name, client.Name)
			}
		}
	})

	t.Run("given empty client list when sorted then returns empty list", func(t *testing.T) {
		clients := []Client{}
		sorted := SortClients(clients)

		if len(sorted) != 0 {
			t.Errorf("expected empty list, got %d items", len(sorted))
		}
	})
}

func TestProjectSorting(t *testing.T) {
	t.Run("given projects with same client when sorted then returns alphabetical by project name", func(t *testing.T) {
		client := Client{ID: 1, Name: "Acme Corp"}
		projects := []Project{
			{ID: 1, Name: "Website Redesign", Client: client},
			{ID: 2, Name: "API Development", Client: client},
			{ID: 3, Name: "Mobile App", Client: client},
		}

		sorted := SortProjects(projects)

		expected := []string{"API Development", "Mobile App", "Website Redesign"}
		for i, project := range sorted {
			if project.Name != expected[i] {
				t.Errorf("expected project %d to be %s, got %s", i, expected[i], project.Name)
			}
		}
	})

	t.Run("given projects with different clients when sorted then sorts by client then project", func(t *testing.T) {
		clientB := Client{ID: 2, Name: "BigCo Industries"}
		clientA := Client{ID: 1, Name: "Acme Corp"}

		projects := []Project{
			{ID: 1, Name: "Website", Client: clientB},
			{ID: 2, Name: "Mobile App", Client: clientA},
			{ID: 3, Name: "API", Client: clientB},
			{ID: 4, Name: "Backend", Client: clientA},
		}

		sorted := SortProjects(projects)

		expectedOrder := []struct {
			clientName  string
			projectName string
		}{
			{"Acme Corp", "Backend"},
			{"Acme Corp", "Mobile App"},
			{"BigCo Industries", "API"},
			{"BigCo Industries", "Website"},
		}

		for i, project := range sorted {
			if project.Client.Name != expectedOrder[i].clientName {
				t.Errorf("expected project %d client to be %s, got %s", i, expectedOrder[i].clientName, project.Client.Name)
			}
			if project.Name != expectedOrder[i].projectName {
				t.Errorf("expected project %d name to be %s, got %s", i, expectedOrder[i].projectName, project.Name)
			}
		}
	})
}

func TestDailyTotal(t *testing.T) {
	t.Run("given empty time entries when calculated then returns 0", func(t *testing.T) {
		entries := []TimeEntry{}
		total := CalculateDailyTotal(entries)

		if total != 0.0 {
			t.Errorf("expected 0, got %f", total)
		}
	})

	t.Run("given single time entry when calculated then returns entry hours", func(t *testing.T) {
		entries := []TimeEntry{
			{Hours: 2.5},
		}
		total := CalculateDailyTotal(entries)

		if total != 2.5 {
			t.Errorf("expected 2.5, got %f", total)
		}
	})

	t.Run("given multiple time entries when calculated then returns sum", func(t *testing.T) {
		entries := []TimeEntry{
			{Hours: 2.5},
			{Hours: 1.25},
			{Hours: 3.0},
			{Hours: 0.5},
		}
		total := CalculateDailyTotal(entries)

		expected := 7.25
		if total != expected {
			t.Errorf("expected %f, got %f", expected, total)
		}
	})

	t.Run("given time entries with zero hours when calculated then includes zeros in sum", func(t *testing.T) {
		entries := []TimeEntry{
			{Hours: 2.0},
			{Hours: 0.0},
			{Hours: 1.5},
		}
		total := CalculateDailyTotal(entries)

		expected := 3.5
		if total != expected {
			t.Errorf("expected %f, got %f", expected, total)
		}
	})
}

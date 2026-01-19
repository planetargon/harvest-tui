package domain

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Client struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Client Client `json:"client"`
	Tasks  []Task `json:"tasks"`
}

type Task struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TimeEntry struct {
	ID         int     `json:"id"`
	Client     Client  `json:"client"`
	Project    Project `json:"project"`
	Task       Task    `json:"task"`
	Notes      string  `json:"notes"`
	Hours      float64 `json:"hours"`
	SpentDate  string  `json:"spent_date"`
	IsRunning  bool    `json:"is_running"`
	IsLocked   bool    `json:"is_locked"`
	IsBillable bool    `json:"is_billable"`
}

func FormatDuration(hours float64) string {
	totalMinutes := int(hours * 60)
	h := totalMinutes / 60
	m := totalMinutes % 60
	return fmt.Sprintf("%d:%02d", h, m)
}

func ParseDuration(durationStr string) (float64, error) {
	durationStr = strings.TrimSpace(durationStr)
	if durationStr == "" {
		return 0, fmt.Errorf("duration cannot be empty")
	}

	parts := strings.Split(durationStr, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid duration format. Use HH:MM (e.g., 1:30)")
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid duration format. Use HH:MM (e.g., 1:30)")
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid duration format. Use HH:MM (e.g., 1:30)")
	}

	if hours < 0 || minutes < 0 || minutes >= 60 {
		return 0, fmt.Errorf("invalid duration format. Use HH:MM (e.g., 1:30)")
	}

	return float64(hours) + float64(minutes)/60.0, nil
}

func SortClients(clients []Client) []Client {
	sorted := make([]Client, len(clients))
	copy(sorted, clients)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}

func SortProjects(projects []Project) []Project {
	sorted := make([]Project, len(projects))
	copy(sorted, projects)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Client.Name != sorted[j].Client.Name {
			return sorted[i].Client.Name < sorted[j].Client.Name
		}
		return sorted[i].Name < sorted[j].Name
	})

	return sorted
}

func CalculateDailyTotal(entries []TimeEntry) float64 {
	total := 0.0
	for _, entry := range entries {
		total += entry.Hours
	}
	return total
}

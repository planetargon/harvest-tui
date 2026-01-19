package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type State struct {
	Recents []RecentEntry `json:"recents"`
}

type RecentEntry struct {
	ClientID  int `json:"client_id"`
	ProjectID int `json:"project_id"`
	TaskID    int `json:"task_id"`
}

func Load() (*State, error) {
	statePath, err := getStatePath()
	if err != nil {
		return nil, fmt.Errorf("could not determine state path: %w", err)
	}

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return &State{Recents: []RecentEntry{}}, nil
	}

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("could not read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("could not parse state file: %w", err)
	}

	if state.Recents == nil {
		state.Recents = []RecentEntry{}
	}

	return &state, nil
}

func (s *State) Save() error {
	statePath, err := getStatePath()
	if err != nil {
		return fmt.Errorf("could not determine state path: %w", err)
	}

	stateDir := filepath.Dir(statePath)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("could not create state directory: %w", err)
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal state: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("could not write state file: %w", err)
	}

	return nil
}

func (s *State) AddRecent(clientID, projectID, taskID int) {
	newEntry := RecentEntry{
		ClientID:  clientID,
		ProjectID: projectID,
		TaskID:    taskID,
	}

	for i, entry := range s.Recents {
		if entry.ClientID == clientID && entry.ProjectID == projectID && entry.TaskID == taskID {
			s.Recents = append(s.Recents[:i], s.Recents[i+1:]...)
			break
		}
	}

	s.Recents = append([]RecentEntry{newEntry}, s.Recents...)

	if len(s.Recents) > 3 {
		s.Recents = s.Recents[:3]
	}
}

func getStatePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "harvest-tui", "state.json"), nil
}

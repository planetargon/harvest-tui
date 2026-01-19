# Harvest TUI Time Tracker - Requirements

## Overview

A terminal-based time tracking application that connects to the Harvest API v2, allowing users to manage their daily time entries from the command line.

## Technical Stack

- **Language**: Go (1.21+)
- **TUI Framework**: Bubble Tea
- **Styling**: Lip Gloss
- **Components**: Bubbles (list, text input, spinner)
- **Target Platforms**: macOS, Linux

## Dependencies

Pin these versions in `go.mod`:
````
github.com/charmbracelet/bubbletea v1.2.4
github.com/charmbracelet/bubbles v0.20.0
github.com/charmbracelet/lipgloss v1.0.0
github.com/BurntSushi/toml v1.4.0
````

## API Reference

- Harvest API v2 Documentation: https://help.getharvest.com/api-v2/
- When implementing API interactions, reference specific documentation pages in commit messages

## Project Structure
````
harvest-tui/
├── cmd/
│   └── harvest-tui/
│       └── main.go
├── internal/
│   ├── config/          # Configuration loading/saving
│   ├── harvest/         # Harvest API client
│   ├── domain/          # Core types and business logic
│   ├── state/           # Application state persistence
│   └── tui/
│       ├── app.go       # Main Bubble Tea application
│       ├── keys.go      # Keybinding definitions
│       ├── styles.go    # Lip Gloss styles
│       └── views/       # Individual view components
├── bin/                  # Build output (gitignored)
├── config.example.toml
├── Makefile
├── CLAUDE.md
├── README.md
├── REQUIREMENTS.md
├── PROGRESS.md
├── .gitignore
├── go.mod
└── go.sum
````

---

## Configuration

### Location
- Config file: `~/.config/harvest-tui/config.toml`
- State file: `~/.config/harvest-tui/state.json`

### Config File Format (`config.example.toml`)
````toml
[harvest]
account_id = ""
access_token = ""
````

### State File Format
````json
{
  "recents": [
    {"client_id": 123, "project_id": 456, "task_id": 789},
    {"client_id": 124, "project_id": 457, "task_id": 790},
    {"client_id": 125, "project_id": 458, "task_id": 791}
  ]
}
````

---

## Data Model

### Client
- ID (int)
- Name (string)

### Project
- ID (int)
- Name (string)
- Client (Client)
- Tasks ([]Task)

### Task
- ID (int)
- Name (string)

### TimeEntry
- ID (int)
- Client (Client)
- Project (Project)
- Task (Task)
- Notes (string)
- Hours (float64)
- SpentDate (string, YYYY-MM-DD)
- IsRunning (bool)
- IsLocked (bool)
- IsBillable (bool)

---

## Features

### Authentication
- Load credentials from config file on startup
- Validate credentials by making a test API call
- Display clear error message if credentials are invalid or missing
- Exit gracefully if authentication fails

### Data Fetching
- On launch, fetch:
  - User's assigned projects (includes client info)
  - Tasks for each project
  - Time entries for the focused day (defaults to today)
- Show loading spinner during fetch
- Handle API rate limits with retry and backoff
- Handle network errors gracefully with user feedback

### Main List View
- Display focused date prominently with day of week
- Show navigation hints for changing days
- List all time entries for the focused day
- Each entry displays:
  - Client → Project → Task
  - Duration (HH:MM format)
  - Notes (truncated if long)
  - Running indicator (⏱️) if `is_running: true`
  - Locked indicator (🔒) if `is_locked: true`
- Show daily total (sum of all entry durations)
- Show "No entries" message if day is empty
- Highlight currently selected entry

### Day Navigation
- Left arrow (←): Move to previous day
- Right arrow (→): Move to next day
- `t`: Jump to today
- Fetch entries when day changes
- Show loading state during fetch

### Create New Entry
- Press `n` to initiate
- Show project selection list:
  - Top section: 3 most recent `Client → Project → Task` combos
  - Divider
  - Bottom section: All projects as `Client → Project`, sorted alphabetically by client then project
- Filter list by typing (substring match via Bubbles list)
- If recent combo selected: skip task selection
- If project selected: show task selection list for that project
- After project/task selected:
  - Notes input (optional, can be empty)
  - Billable toggle (default based on project setting)
  - Start timer immediately on confirm
- Update recents list (most recent first, cap at 3, no duplicates)
- Escape at any point cancels and returns to main list

### Start/Stop Timer
- Press `s` on selected entry
- If entry is running: stop the timer
- If entry is stopped: start the timer
- Note: Harvest only allows one running timer; starting one stops any other
- Blocked if entry is locked (show message)

### Edit Entry
- Press `e` on selected entry
- Blocked if entry is locked (show message)
- Edit view allows changing:
  - Notes (text input)
  - Duration (text input, validate format)
  - Billable toggle
- Does NOT allow changing project/task (create new entry instead)
- Save on confirm, cancel with Escape

### Delete Entry
- Press `d` on selected entry
- Blocked if entry is locked (show message)
- Show confirmation dialog: "Delete this entry? (y/n)"
- `y` confirms deletion
- `n` or Escape cancels
- Return to main list after deletion

### Refresh
- Press `r` to manually refresh data from API
- Re-fetches entries for current focused day
- Shows loading state

### Help
- Press `?` to show help overlay
- Lists all keybindings
- Press any key to dismiss

### Quit
- Press `q` to quit
- No confirmation needed (data is saved to API in real-time)

---

## Keybindings
````
Navigation
  j / ↓       Move selection down
  k / ↑       Move selection up
  ←           Previous day
  →           Next day
  t           Jump to today

Actions
  n           New time entry
  e           Edit selected entry
  s           Start/stop timer on selected entry
  d           Delete selected entry
  r           Refresh data from API

General
  ?           Show help
  q           Quit
  Esc         Cancel / back
````

---

## UI Layout

### Main View
````
┌─────────────────────────────────────────────────────────────┐
│  Harvest Time Tracker            ◀ Mon, Jan 13, 2025 ▶     │
├─────────────────────────────────────────────────────────────┤
│  Entries                                    Total: 6:45    │
│  ───────────────────────────────────────────────────────────│
│    Acme Corp → API Development → Code Review      1:30     │
│    "Reviewing authentication module PRs"                   │
│                                                            │
│  ▶ Acme Corp → Mobile App → Development           2:00 ⏱️  │
│    "Feature: push notifications"                           │
│                                                            │
│    BigCo Industries → Consulting → Meetings       0:45 🔒  │
│    "Weekly sync"                                           │
│                                                            │
│    Internal → Admin → Email                       0:30     │
│                                                            │
├─────────────────────────────────────────────────────────────┤
│  n new  e edit  s start/stop  d delete  ? help  q quit    │
└─────────────────────────────────────────────────────────────┘
````

### Project Selection View
````
┌─────────────────────────────────────────────────────────────┐
│  New Entry: Select Project                                 │
├─────────────────────────────────────────────────────────────┤
│  Filter: _                                                 │
│  ───────────────────────────────────────────────────────────│
│  Recent                                                    │
│  ▶ Acme Corp → API Development → Code Review               │
│    Acme Corp → Mobile App → Development                    │
│    Internal → Admin → Email                                │
│  ───────────────────────────────────────────────────────────│
│  All Projects                                              │
│    Acme Corp → API Development                             │
│    Acme Corp → Mobile App                                  │
│    Acme Corp → Website Redesign                            │
│    BigCo Industries → Consulting                           │
│    BigCo Industries → Maintenance                          │
│    Internal → Admin                                        │
│    Internal → Meetings                                     │
│                                                            │
├─────────────────────────────────────────────────────────────┤
│  ↑↓ navigate  enter select  esc cancel                    │
└─────────────────────────────────────────────────────────────┘
````

### Delete Confirmation
````
┌─────────────────────────────────────────────────────────────┐
│  Delete Entry?                                             │
│  ───────────────────────────────────────────────────────────│
│  Acme Corp → API Development → Code Review                 │
│  1:30 - "Reviewing authentication module PRs"              │
│                                                            │
│  This action cannot be undone.                             │
│                                                            │
│  y confirm  n cancel                                       │
└─────────────────────────────────────────────────────────────┘
````

---

## Error Handling

### Error Message Guidelines

Format: `Error: <what failed>. <what to do>`

Examples:
- `Error: Could not load config file. Create ~/.config/harvest-tui/config.toml with your Harvest credentials.`
- `Error: Authentication failed. Check your account_id and access_token in config.toml.`
- `Error: Network request failed. Press 'r' to retry.`
- `Error: Entry is locked and cannot be modified.`
- `Error: Invalid duration format. Use HH:MM (e.g., 1:30).`
- `Error: Rate limited by Harvest API. Retrying in 5 seconds...`

### Error Scenarios
- Network errors: Show message, allow retry with `r`
- Auth errors: Show message, instruct user to check config, exit
- API rate limits: Auto-retry with backoff, show "Rate limited, retrying..."
- Invalid input: Show inline validation messages
- Locked entry actions: Show "Entry is locked and cannot be modified"

---

## Testing Requirements

All tests should be written in BDD style using descriptive test names that read as specifications.

### Test File Naming
- `*_test.go` files alongside implementation
- Use table-driven tests where appropriate

### Test Style (BDD)
````go
func TestTimeEntry(t *testing.T) {
    t.Run("given a running entry when stopped then sets is_running to false", func(t *testing.T) {
        // ...
    })
    
    t.Run("given a locked entry when edit requested then returns error", func(t *testing.T) {
        // ...
    })
}
````

### Required Test Coverage

#### Domain Tests (`internal/domain/*_test.go`)
- Time calculations (sum durations, format hours)
- Entry sorting
- Client/project sorting (alphabetical)
- Recents management (add, bump to top, cap at 3, dedup)
- Duration parsing and formatting

#### Config Tests (`internal/config/*_test.go`)
- Load valid config file
- Handle missing config file
- Handle malformed config file
- Handle missing required fields

#### State Tests (`internal/state/*_test.go`)
- Load existing state file
- Handle missing state file (create defaults)
- Save state file
- Recents persistence

#### Harvest Client Tests (`internal/harvest/*_test.go`)
- Fetch projects (mock HTTP)
- Fetch tasks for project (mock HTTP)
- Fetch time entries for date (mock HTTP)
- Create time entry (mock HTTP)
- Update time entry (mock HTTP)
- Delete time entry (mock HTTP)
- Start timer (mock HTTP)
- Stop timer (mock HTTP)
- Handle 401 unauthorized
- Handle 429 rate limit
- Handle network timeout
- Handle malformed JSON response

#### TUI Model Tests (`internal/tui/*_test.go`)
- Navigation: move selection up/down
- Navigation: change focused day
- Navigation: jump to today
- View transitions: main → new entry → project select
- View transitions: project select → task select
- View transitions: edit view
- View transitions: delete confirmation
- Action: cannot edit locked entry
- Action: cannot delete locked entry
- Action: start timer updates state
- Action: stop timer updates state
- Filter: project list filtering
- Recents: appear at top of project list
- Recents: selecting recent skips task selection

---

## Development Workflow

### Makefile
````makefile
.PHONY: build test lint fmt run clean check

build:
	go build -o bin/harvest-tui ./cmd/harvest-tui

test:
	go test -v ./...

lint:
	go vet ./...

fmt:
	go fmt ./...

check: fmt lint test

run: build
	./bin/harvest-tui

clean:
	rm -rf bin/
````

### .gitignore
````
bin/
*.exe
.DS_Store
*.log
````

### Commit Guidelines
- Small, focused commits (one logical change per commit)
- Run `make check` before each commit
- All tests must pass before committing
- Update documentation if behavior changes
- Reference Harvest API docs in commits that implement API calls

### Commit Message Format

Use conventional commits:
- `feat(scope): description` for new features
- `test(scope): description` for tests
- `fix(scope): description` for bug fixes
- `docs: description` for documentation
- `chore: description` for maintenance tasks

Example with API reference:
````
feat(harvest): implement fetch projects endpoint

Implements GET /v2/projects endpoint to retrieve user's assigned projects.

API Reference: https://help.getharvest.com/api-v2/projects-api/projects/projects/
````

### Pre-commit Checklist
- [ ] `make check` passes (fmt, lint, test)
- [ ] README updated (if user-facing changes)
- [ ] PROGRESS.md updated with completed step

---

## Progress Tracking

Track progress in `PROGRESS.md`. Update after completing each step.

---

## Implementation Steps

### Phase 1: Project Setup (Steps 1-6)
1. Initialize Go module with `go mod init github.com/USER/harvest-tui`
2. Create project directory structure (all folders under `cmd/` and `internal/`)
3. Create `.gitignore` file
4. Create `Makefile`
5. Create `config.example.toml`
6. Create `CLAUDE.md` with agent instructions

### Phase 2: Documentation Setup (Steps 7-9)
7. Create `README.md` with installation and build instructions (placeholder for usage)
8. Create `PROGRESS.md` with all steps listed
9. Initial commit with project scaffolding

### Phase 3: Minimal TUI (Steps 10-12)
10. Add Bubble Tea, Bubbles, and Lip Gloss dependencies with pinned versions
11. Create `main.go` with minimal Bubble Tea app that displays "Harvest TUI" and quits on `q`
12. Verify `make build` and `make run` work

### Phase 4: Configuration (Steps 13-18)
13. Define config types in `internal/config/config.go`
14. Write tests for config loading (valid file)
15. Write tests for config loading (missing file, malformed file, missing fields)
16. Implement config loading from `~/.config/harvest-tui/config.toml`
17. Integrate config loading into main.go with appropriate error messages
18. Verify app shows helpful error when config is missing

### Phase 5: State Management (Steps 19-23)
19. Define state types in `internal/state/state.go`
20. Write tests for state loading (existing file, missing file)
21. Write tests for state saving
22. Implement state loading/saving for recents
23. Write tests for recents management (add, bump to top, cap at 3, dedup) and implement

### Phase 6: Domain Types (Steps 24-28)
24. Define Client, Project, Task, TimeEntry types in `internal/domain/types.go`
25. Write tests for duration formatting (hours float to HH:MM string)
26. Implement duration formatting
27. Write tests for duration parsing (HH:MM string to hours float)
28. Implement duration parsing with validation

### Phase 7: Domain Logic (Steps 29-33)
29. Write tests for sorting clients alphabetically
30. Write tests for sorting projects alphabetically within client
31. Implement sorting functions
32. Write tests for daily total calculation
33. Implement daily total calculation

### Phase 8: Harvest API Client - Setup (Steps 34-38)
34. Create HTTP client wrapper in `internal/harvest/client.go` with auth headers
35. Write tests for authentication validation endpoint (mock HTTP)
36. Implement authentication validation (GET /v2/users/me)
37. Write tests for error handling (401 unauthorized)
38. Write tests for error handling (429 rate limit, timeout, malformed JSON)

### Phase 9: Harvest API Client - Read Operations (Steps 39-46)
39. Write tests for fetch projects endpoint (mock HTTP)
40. Implement fetch projects (GET /v2/projects) - include client data
41. Write tests for fetch project task assignments endpoint (mock HTTP)
42. Implement fetch project task assignments (GET /v2/task_assignments)
43. Write tests for fetch time entries endpoint (mock HTTP)
44. Implement fetch time entries (GET /v2/time_entries?from=DATE&to=DATE)
45. Write tests for aggregating projects with their tasks
46. Implement project/task aggregation sorted alphabetically by client then project

### Phase 10: Harvest API Client - Write Operations (Steps 47-54)
47. Write tests for create time entry endpoint (mock HTTP)
48. Implement create time entry (POST /v2/time_entries)
49. Write tests for update time entry endpoint (mock HTTP)
50. Implement update time entry (PATCH /v2/time_entries/{id})
51. Write tests for delete time entry endpoint (mock HTTP)
52. Implement delete time entry (DELETE /v2/time_entries/{id})
53. Write tests for start/stop timer endpoints (mock HTTP)
54. Implement restart timer (PATCH /v2/time_entries/{id}/restart) and stop timer (PATCH /v2/time_entries/{id}/stop)

### Phase 11: TUI Foundation (Steps 55-60)
55. Define view states enum (list, selectProject, selectTask, editEntry, confirmDelete, help)
56. Define app model with all required state in `internal/tui/app.go`
57. Define keybindings in `internal/tui/keys.go`
58. Define base styles in `internal/tui/styles.go`
59. Write tests for initial model state
60. Implement model initialization

### Phase 12: TUI Main List View (Steps 61-68)
61. Write tests for main list view rendering with entries
62. Implement main list view rendering with mock data
63. Write tests for main list view rendering with no entries
64. Implement empty state display
65. Write tests for selection navigation (j/k, up/down)
66. Implement selection navigation
67. Write tests for daily total display
68. Implement daily total in header

### Phase 13: TUI Day Navigation (Steps 69-74)
69. Write tests for day navigation (left arrow = previous day)
70. Write tests for day navigation (right arrow = next day)
71. Write tests for jump to today (t key)
72. Implement day navigation with date state changes
73. Implement loading spinner during data fetch
74. Implement error display for failed fetches

### Phase 14: TUI New Entry - Project Selection (Steps 75-80)
75. Write tests for view transition to project selection on `n`
76. Implement project selection view with Bubbles list
77. Write tests for project list displaying Client → Project format
78. Implement alphabetically sorted project list
79. Write tests for recents appearing at top of project list
80. Implement recents section with divider in project selection

### Phase 15: TUI New Entry - Task Selection (Steps 81-85)
81. Write tests for transition to task selection after project chosen
82. Write tests for skipping task selection when recent combo selected
83. Implement task selection view
84. Write tests for escape returning to main list from any point
85. Implement cancel behavior throughout new entry flow

### Phase 16: TUI New Entry - Details (Steps 86-91)
86. Write tests for notes input view
87. Implement notes input with text input component
88. Write tests for billable toggle
89. Implement billable toggle
90. Write tests for entry creation API call on confirm
91. Implement entry creation with API call and recents update

### Phase 17: TUI Start/Stop Timer (Steps 92-96)
92. Write tests for start timer action on stopped entry
93. Write tests for stop timer action on running entry
94. Write tests for blocked start/stop on locked entry
95. Implement start/stop timer with API call
96. Implement locked entry check with user feedback

### Phase 18: TUI Edit Entry (Steps 97-103)
97. Write tests for view transition to edit on `e`
98. Write tests for blocked edit on locked entry
99. Implement edit view (notes field)
100. Implement edit view (duration field with validation)
101. Implement edit view (billable toggle)
102. Write tests for edit save with API call
103. Implement edit save functionality

### Phase 19: TUI Delete Entry (Steps 104-109)
104. Write tests for view transition to delete confirmation on `d`
105. Write tests for blocked delete on locked entry
106. Implement delete confirmation view
107. Write tests for delete action on `y` confirm
108. Write tests for cancel delete on `n` or escape
109. Implement delete functionality with API call

### Phase 20: TUI Polish (Steps 110-115)
110. Implement help overlay view listing all keybindings
111. Implement status messages for successful actions (created, updated, deleted)
112. Implement status messages for errors
113. Add running timer indicator (⏱️) to entry display
114. Add locked indicator (🔒) to entry display
115. Review and refine all view layouts for consistency

### Phase 21: Integration (Steps 116-120)
116. Wire up real API client to TUI (replace mock data)
117. Integrate config loading at startup with error handling
118. Integrate state loading/saving for recents
119. Test full flow: launch → load data → create entry
120. Test full flow: start/stop timer → edit → delete

### Phase 22: Edge Cases (Steps 121-125)
121. Handle edge case: no projects assigned to user
122. Handle edge case: project with no tasks
123. Handle edge case: API unavailable at startup
124. Handle edge case: API fails mid-operation
125. Handle edge case: very long project/task/client names (truncation)

### Phase 23: Final Polish (Steps 126-130)
126. Final README update with complete usage instructions and keybindings
127. Review all error messages for consistency with guidelines
128. Remove any dead code or unused imports
129. Run `make check` and fix any issues
130. Final manual testing of all acceptance criteria

---

## CLAUDE.md

Create this file in the repo root:
````markdown
# Claude Code Instructions

## Before Every Commit
Run `make check` and ensure it passes.

## Commit Message Format
Use conventional commits:
- `feat(scope): description` for new features
- `test(scope): description` for tests
- `fix(scope): description` for bug fixes
- `docs: description` for documentation
- `chore: description` for maintenance tasks

When implementing Harvest API calls, include the API reference URL in the commit body.

## Progress Tracking
Update PROGRESS.md after completing each step. Mark the step as complete and update "Current Step" to the next step.

## Test Style
Write tests in BDD style with descriptive names:
```go
t.Run("given X when Y then Z", func(t *testing.T) { ... })
```

## If Stuck
1. Document the blocker in PROGRESS.md under "Blockers"
2. Note what was attempted
3. Stop and wait for human input
````

---

## PROGRESS.md

Create this file in the repo root:
````markdown
# Progress

## Current Step
Step 1: Initialize Go module with `go mod init github.com/USER/harvest-tui`

## Completed Steps
(none yet)

## Blockers
(none)

## Notes
Starting fresh.

---

## All Steps

### Phase 1: Project Setup (Steps 1-6)
- [ ] Step 1: Initialize Go module with `go mod init github.com/USER/harvest-tui`
- [ ] Step 2: Create project directory structure (all folders under `cmd/` and `internal/`)
- [ ] Step 3: Create `.gitignore` file
- [ ] Step 4: Create `Makefile`
- [ ] Step 5: Create `config.example.toml`
- [ ] Step 6: Create `CLAUDE.md` with agent instructions

### Phase 2: Documentation Setup (Steps 7-9)
- [ ] Step 7: Create `README.md` with installation and build instructions (placeholder for usage)
- [ ] Step 8: Create `PROGRESS.md` with all steps listed
- [ ] Step 9: Initial commit with project scaffolding

### Phase 3: Minimal TUI (Steps 10-12)
- [ ] Step 10: Add Bubble Tea, Bubbles, and Lip Gloss dependencies with pinned versions
- [ ] Step 11: Create `main.go` with minimal Bubble Tea app that displays "Harvest TUI" and quits on `q`
- [ ] Step 12: Verify `make build` and `make run` work

### Phase 4: Configuration (Steps 13-18)
- [ ] Step 13: Define config types in `internal/config/config.go`
- [ ] Step 14: Write tests for config loading (valid file)
- [ ] Step 15: Write tests for config loading (missing file, malformed file, missing fields)
- [ ] Step 16: Implement config loading from `~/.config/harvest-tui/config.toml`
- [ ] Step 17: Integrate config loading into main.go with appropriate error messages
- [ ] Step 18: Verify app shows helpful error when config is missing

### Phase 5: State Management (Steps 19-23)
- [ ] Step 19: Define state types in `internal/state/state.go`
- [ ] Step 20: Write tests for state loading (existing file, missing file)
- [ ] Step 21: Write tests for state saving
- [ ] Step 22: Implement state loading/saving for recents
- [ ] Step 23: Write tests for recents management (add, bump to top, cap at 3, dedup) and implement

### Phase 6: Domain Types (Steps 24-28)
- [ ] Step 24: Define Client, Project, Task, TimeEntry types in `internal/domain/types.go`
- [ ] Step 25: Write tests for duration formatting (hours float to HH:MM string)
- [ ] Step 26: Implement duration formatting
- [ ] Step 27: Write tests for duration parsing (HH:MM string to hours float)
- [ ] Step 28: Implement duration parsing with validation

### Phase 7: Domain Logic (Steps 29-33)
- [ ] Step 29: Write tests for sorting clients alphabetically
- [ ] Step 30: Write tests for sorting projects alphabetically within client
- [ ] Step 31: Implement sorting functions
- [ ] Step 32: Write tests for daily total calculation
- [ ] Step 33: Implement daily total calculation

### Phase 8: Harvest API Client - Setup (Steps 34-38)
- [ ] Step 34: Create HTTP client wrapper in `internal/harvest/client.go` with auth headers
- [ ] Step 35: Write tests for authentication validation endpoint (mock HTTP)
- [ ] Step 36: Implement authentication validation (GET /v2/users/me)
- [ ] Step 37: Write tests for error handling (401 unauthorized)
- [ ] Step 38: Write tests for error handling (429 rate limit, timeout, malformed JSON)

### Phase 9: Harvest API Client - Read Operations (Steps 39-46)
- [ ] Step 39: Write tests for fetch projects endpoint (mock HTTP)
- [ ] Step 40: Implement fetch projects (GET /v2/projects) - include client data
- [ ] Step 41: Write tests for fetch project task assignments endpoint (mock HTTP)
- [ ] Step 42: Implement fetch project task assignments (GET /v2/task_assignments)
- [ ] Step 43: Write tests for fetch time entries endpoint (mock HTTP)
- [ ] Step 44: Implement fetch time entries (GET /v2/time_entries?from=DATE&to=DATE)
- [ ] Step 45: Write tests for aggregating projects with their tasks
- [ ] Step 46: Implement project/task aggregation sorted alphabetically by client then project

### Phase 10: Harvest API Client - Write Operations (Steps 47-54)
- [ ] Step 47: Write tests for create time entry endpoint (mock HTTP)
- [ ] Step 48: Implement create time entry (POST /v2/time_entries)
- [ ] Step 49: Write tests for update time entry endpoint (mock HTTP)
- [ ] Step 50: Implement update time entry (PATCH /v2/time_entries/{id})
- [ ] Step 51: Write tests for delete time entry endpoint (mock HTTP)
- [ ] Step 52: Implement delete time entry (DELETE /v2/time_entries/{id})
- [ ] Step 53: Write tests for start/stop timer endpoints (mock HTTP)
- [ ] Step 54: Implement restart timer (PATCH /v2/time_entries/{id}/restart) and stop timer (PATCH /v2/time_entries/{id}/stop)

### Phase 11: TUI Foundation (Steps 55-60)
- [ ] Step 55: Define view states enum (list, selectProject, selectTask, editEntry, confirmDelete, help)
- [ ] Step 56: Define app model with all required state in `internal/tui/app.go`
- [ ] Step 57: Define keybindings in `internal/tui/keys.go`
- [ ] Step 58: Define base styles in `internal/tui/styles.go`
- [ ] Step 59: Write tests for initial model state
- [ ] Step 60: Implement model initialization

### Phase 12: TUI Main List View (Steps 61-68)
- [ ] Step 61: Write tests for main list view rendering with entries
- [ ] Step 62: Implement main list view rendering with mock data
- [ ] Step 63: Write tests for main list view rendering with no entries
- [ ] Step 64: Implement empty state display
- [ ] Step 65: Write tests for selection navigation (j/k, up/down)
- [ ] Step 66: Implement selection navigation
- [ ] Step 67: Write tests for daily total display
- [ ] Step 68: Implement daily total in header

### Phase 13: TUI Day Navigation (Steps 69-74)
- [ ] Step 69: Write tests for day navigation (left arrow = previous day)
- [ ] Step 70: Write tests for day navigation (right arrow = next day)
- [ ] Step 71: Write tests for jump to today (t key)
- [ ] Step 72: Implement day navigation with date state changes
- [ ] Step 73: Implement loading spinner during data fetch
- [ ] Step 74: Implement error display for failed fetches

### Phase 14: TUI New Entry - Project Selection (Steps 75-80)
- [ ] Step 75: Write tests for view transition to project selection on `n`
- [ ] Step 76: Implement project selection view with Bubbles list
- [ ] Step 77: Write tests for project list displaying Client → Project format
- [ ] Step 78: Implement alphabetically sorted project list
- [ ] Step 79: Write tests for recents appearing at top of project list
- [ ] Step 80: Implement recents section with divider in project selection

### Phase 15: TUI New Entry - Task Selection (Steps 81-85)
- [ ] Step 81: Write tests for transition to task selection after project chosen
- [ ] Step 82: Write tests for skipping task selection when recent combo selected
- [ ] Step 83: Implement task selection view
- [ ] Step 84: Write tests for escape returning to main list from any point
- [ ] Step 85: Implement cancel behavior throughout new entry flow

### Phase 16: TUI New Entry - Details (Steps 86-91)
- [ ] Step 86: Write tests for notes input view
- [ ] Step 87: Implement notes input with text input component
- [ ] Step 88: Write tests for billable toggle
- [ ] Step 89: Implement billable toggle
- [ ] Step 90: Write tests for entry creation API call on confirm
- [ ] Step 91: Implement entry creation with API call and recents update

### Phase 17: TUI Start/Stop Timer (Steps 92-96)
- [ ] Step 92: Write tests for start timer action on stopped entry
- [ ] Step 93: Write tests for stop timer action on running entry
- [ ] Step 94: Write tests for blocked start/stop on locked entry
- [ ] Step 95: Implement start/stop timer with API call
- [ ] Step 96: Implement locked entry check with user feedback

### Phase 18: TUI Edit Entry (Steps 97-103)
- [ ] Step 97: Write tests for view transition to edit on `e`
- [ ] Step 98: Write tests for blocked edit on locked entry
- [ ] Step 99: Implement edit view (notes field)
- [ ] Step 100: Implement edit view (duration field with validation)
- [ ] Step 101: Implement edit view (billable toggle)
- [ ] Step 102: Write tests for edit save with API call
- [ ] Step 103: Implement edit save functionality

### Phase 19: TUI Delete Entry (Steps 104-109)
- [ ] Step 104: Write tests for view transition to delete confirmation on `d`
- [ ] Step 105: Write tests for blocked delete on locked entry
- [ ] Step 106: Implement delete confirmation view
- [ ] Step 107: Write tests for delete action on `y` confirm
- [ ] Step 108: Write tests for cancel delete on `n` or escape
- [ ] Step 109: Implement delete functionality with API call

### Phase 20: TUI Polish (Steps 110-115)
- [ ] Step 110: Implement help overlay view listing all keybindings
- [ ] Step 111: Implement status messages for successful actions (created, updated, deleted)
- [ ] Step 112: Implement status messages for errors
- [ ] Step 113: Add running timer indicator (⏱️) to entry display
- [ ] Step 114: Add locked indicator (🔒) to entry display
- [ ] Step 115: Review and refine all view layouts for consistency

### Phase 21: Integration (Steps 116-120)
- [ ] Step 116: Wire up real API client to TUI (replace mock data)
- [ ] Step 117: Integrate config loading at startup with error handling
- [ ] Step 118: Integrate state loading/saving for recents
- [ ] Step 119: Test full flow: launch → load data → create entry
- [ ] Step 120: Test full flow: start/stop timer → edit → delete

### Phase 22: Edge Cases (Steps 121-125)
- [ ] Step 121: Handle edge case: no projects assigned to user
- [ ] Step 122: Handle edge case: project with no tasks
- [ ] Step 123: Handle edge case: API unavailable at startup
- [ ] Step 124: Handle edge case: API fails mid-operation
- [ ] Step 125: Handle edge case: very long project/task/client names (truncation)

### Phase 23: Final Polish (Steps 126-130)
- [ ] Step 126: Final README update with complete usage instructions and keybindings
- [ ] Step 127: Review all error messages for consistency with guidelines
- [ ] Step 128: Remove any dead code or unused imports
- [ ] Step 129: Run `make check` and fix any issues
- [ ] Step 130: Final manual testing of all acceptance criteria
````

---

## README Structure

The README.md should contain:

1. **Title and Description** - One-line description of what the app does
2. **Installation**
   - Prerequisites (Go 1.21+)
   - Build from source instructions
   - Binary installation (if applicable)
3. **Configuration**
   - How to get Harvest API credentials
   - Config file location and format
   - Link to `config.example.toml`
4. **Usage**
   - Launching the app
   - Keybindings reference
   - Common workflows (start tracking, stop, edit, delete)
5. **Development**
   - Running tests (`make test`)
   - Building locally (`make build`)
   - Full check (`make check`)
6. **License**

---

## Resumption Protocol

If work is interrupted and needs to resume:

1. Read `PROGRESS.md` to identify current step
2. Check git log for last successful commit
3. Run `make check` to verify current state
4. Continue from the current step noted in PROGRESS.md
5. If blocked, document the blocker in PROGRESS.md and request human input

---

## Acceptance Tests (Manual)

After completion, verify these scenarios work:

1. **Fresh install**: Delete config, launch app, see helpful error about missing config
2. **Invalid credentials**: Use bad token, see auth error, app exits cleanly
3. **Empty day**: Navigate to a day with no entries, see "No entries" message
4. **Full create flow**: Press `n`, select project, select task, add notes, confirm → entry appears with timer running
5. **Recent combo**: Create entry, press `n` again, recent appears at top, select it → skips task selection
6. **Start/stop timer**: Select entry, press `s` to stop, press `s` again to start
7. **Edit entry**: Select entry, press `e`, change notes and duration, save → changes reflected
8. **Delete entry**: Select entry, press `d`, see confirmation, press `y` → entry removed
9. **Cancel delete**: Select entry, press `d`, press `n` → returns to list, entry intact
10. **Locked entry**: Navigate to a day with locked entries, verify `e`, `s`, and `d` are blocked with message
11. **Day navigation**: Press left/right arrows through several days, verify data loads correctly
12. **Jump to today**: Navigate away, press `t`, verify returns to today
13. **Help overlay**: Press `?`, see keybindings, press any key to dismiss
14. **Recents cap**: Create 4 entries with different project/task combos, verify only 3 most recent appear
15. **Refresh**: Press `r`, see loading state, data refreshes

---

## Definition of Done

The project is complete when:

- [ ] All 130 steps are marked complete in PROGRESS.md
- [ ] All tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] No vet warnings (`make lint`)
- [ ] `make check` passes with no issues
- [ ] README contains complete installation and usage instructions
- [ ] App successfully connects to Harvest API with valid credentials
- [ ] All 15 acceptance tests pass manually
- [ ] All features listed in this document are functional


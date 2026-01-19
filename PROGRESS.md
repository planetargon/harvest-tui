# Progress

## Current Step
Step 13: Define config types in `internal/config/config.go`

## Completed Steps
- [x] Step 1: Initialize Go module with `go mod init github.com/planetargon/argon-harvest-tui`
- [x] Step 2: Create project directory structure (all folders under `cmd/` and `internal/`)
- [x] Step 3: Create `.gitignore` file
- [x] Step 4: Create `Makefile`
- [x] Step 5: Create `config.example.toml`
- [x] Step 6: Create `CLAUDE.md` with agent instructions
- [x] Step 7: Create `README.md` with installation and build instructions (placeholder for usage)
- [x] Step 8: Create `PROGRESS.md` with all steps listed
- [x] Step 9: Initial commit with project scaffolding
- [x] Step 10: Add Bubble Tea, Bubbles, and Lip Gloss dependencies with pinned versions
- [x] Step 11: Create `main.go` with minimal Bubble Tea app that displays "Harvest TUI" and quits on `q`
- [x] Step 12: Verify `make build` and `make run` work

## Blockers
(none)

## Notes
Phase 1 (Project Setup) and Phase 2 (Documentation Setup) completed successfully.
Phase 3 (Minimal TUI) completed successfully - basic Bubble Tea app working.
Starting Phase 4: Configuration.

---

## All Steps

### Phase 1: Project Setup (Steps 1-6)
- [x] Step 1: Initialize Go module with `go mod init github.com/USER/harvest-tui`
- [x] Step 2: Create project directory structure (all folders under `cmd/` and `internal/`)
- [x] Step 3: Create `.gitignore` file
- [x] Step 4: Create `Makefile`
- [x] Step 5: Create `config.example.toml`
- [x] Step 6: Create `CLAUDE.md` with agent instructions

### Phase 2: Documentation Setup (Steps 7-9)
- [x] Step 7: Create `README.md` with installation and build instructions (placeholder for usage)
- [x] Step 8: Create `PROGRESS.md` with all steps listed
- [x] Step 9: Initial commit with project scaffolding

### Phase 3: Minimal TUI (Steps 10-12)
- [x] Step 10: Add Bubble Tea, Bubbles, and Lip Gloss dependencies with pinned versions
- [x] Step 11: Create `main.go` with minimal Bubble Tea app that displays "Harvest TUI" and quits on `q`
- [x] Step 12: Verify `make build` and `make run` work

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
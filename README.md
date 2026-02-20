# Harvest TUI Time Tracker

A terminal-based time tracking application that connects to the Harvest API v2, allowing users to manage their daily time entries.

### Time Sheet view

<img width="649" height="349" alt="image" src="https://github.com/user-attachments/assets/364c6784-8c2e-4779-be8c-8794c0038917" />

### Add/Edit Time Entry
<img width="644" height="252" alt="image" src="https://github.com/user-attachments/assets/380d8eeb-e67e-4864-b676-7dbdb0f7f292" />

### Help Menu
<img width="652" height="428" alt="image" src="https://github.com/user-attachments/assets/c3c7720f-e66c-45a3-83a2-d81b4bde664d" />



## Installation

### Prerequisites

- Go 1.21 or higher

### Build from Source

1. Clone this repository:
   ```bash
   git clone https://github.com/planetargon/argon-harvest-tui.git
   cd argon-harvest-tui
   ```

2. Build the application:
   ```bash
   make build
   ```

The binary will be created at `bin/harvest-tui`.

3. Run the application:
   ```bash
   ./bin/harvest-tui
   ```

## Configuration

### Getting Harvest API Credentials

1. Log into your Harvest account
2. Go to Settings → Integrations → Developers
3. Create a new Personal Access Token
4. Note your Account ID and Access Token

### Setup Config File

1. Copy the example config:
   ```bash
   mkdir -p ~/.config/harvest-tui
   cp config.example.toml ~/.config/harvest-tui/config.toml
   ```

2. Edit `~/.config/harvest-tui/config.toml` with your credentials:
   ```toml
   [harvest]
   account_id = "YOUR_ACCOUNT_ID"
   access_token = "YOUR_ACCESS_TOKEN"
   ```

## Usage

Launch the application:
```bash
./bin/harvest-tui
```

### Keybindings

#### Navigation
| Key | Action |
|-----|--------|
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `←` / `h` | Previous day |
| `→` / `l` | Next day |
| `t` | Jump to today |

#### Time Entry Actions
| Key | Action |
|-----|--------|
| `n` | Create new time entry |
| `e` | Edit selected entry |
| `d` | Delete selected entry |
| `s` | Start/stop timer on selected entry |

#### General
| Key | Action |
|-----|--------|
| `?` | Toggle help overlay |
| `q` / `Esc` | Quit / go back |
| `Ctrl+C` | Force quit |

## Development

### Running Tests
```bash
make test
```

### Building Locally
```bash
make build
```

### Full Check (Format, Lint, Test)
```bash
make check
```

## License

This project is licensed under the [MIT License](LICENSE).

# Harvest TUI Time Tracker

A terminal-based time tracking application that connects to the Harvest API v2, allowing users to manage their daily time entries from the command line.

<img width="651" height="383" alt="image" src="https://github.com/user-attachments/assets/0b434def-7c80-4f90-8ce6-6145208e9d6d" />

<img width="654" height="330" alt="image" src="https://github.com/user-attachments/assets/ba1ba147-05a0-4c6a-a028-dd5a1638dad0" />

<img width="296" height="386" alt="image" src="https://github.com/user-attachments/assets/99de0478-bc13-4fb8-9355-0da7334a88c1" />



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

*Note: Usage instructions will be added as features are implemented.*

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

MIT

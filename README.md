# Harvest TUI Time Tracker

A terminal-based time tracking application that connects to the [Harvest API v2](https://help.getharvest.com/api-v2/), allowing users to manage their daily time entries from the CLI.

[Planet Argon](https://www.planetargon.com) has been a long-time customer of Harvest and our software engineers wanted to build a tool for tracking client billables from the command line. This is an open source project that integrates with the Harvest API — there is no collaboration with or endorsement by either party.

### Time Sheet view

<img width="649" height="349" alt="image" src="https://github.com/user-attachments/assets/364c6784-8c2e-4779-be8c-8794c0038917" />

### Add/Edit Time Entry
<img width="644" height="252" alt="image" src="https://github.com/user-attachments/assets/380d8eeb-e67e-4864-b676-7dbdb0f7f292" />

### Help Menu
<img width="652" height="428" alt="image" src="https://github.com/user-attachments/assets/c3c7720f-e66c-45a3-83a2-d81b4bde664d" />



## Installation

### Download a Release Binary (Recommended)

Download the latest binary for your platform from the [Releases page](https://github.com/planetargon/harvest-tui/releases).

Or use curl to download directly (example for macOS Apple Silicon):

```bash
curl -sL https://github.com/planetargon/harvest-tui/releases/latest/download/harvest-tui_darwin_arm64.tar.gz | tar xz
sudo mv harvest-tui /usr/local/bin/
```

### Install with Go

```bash
go install github.com/planetargon/harvest-tui/cmd/harvest-tui@latest
```

The binary is installed to `~/go/bin`. If that directory isn't already on your `$PATH`, add it:

```bash
export PATH="$HOME/go/bin:$PATH"
```

Otherwise, you can run it directly with `~/go/bin/harvest-tui`.

### Build from Source

1. Clone this repository:
   ```bash
   git clone https://github.com/planetargon/harvest-tui.git
   cd harvest-tui
   ```

2. Build the application:
   ```bash
   make build
   ```

The binary will be created at `bin/harvest-tui`.

## Updating

- **Release binary:** Download the latest version from the [Releases page](https://github.com/planetargon/harvest-tui/releases) and replace the existing binary.
- **Go install:** Run `go install github.com/planetargon/harvest-tui/cmd/harvest-tui@latest` again.
- **Build from source:** Pull the latest changes and rebuild:
  ```bash
  git pull
  make build
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
harvest-tui
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

## Disclaimer

[Harvest](https://www.getharvest.com/) is a registered trademark of [Bending Spoons US Inc](https://bendingspoons.com/). This project has no direct affiliation with Harvest or Bending Spoons. It is an independent open source project that integrates with the [Harvest API v2](https://help.getharvest.com/api-v2/).

## License

This project is licensed under the [MIT License](LICENSE).

## About Planet Argon

![Planet Argon](https://pa-github-assets.s3.amazonaws.com/PARGON_logo_digital_COL-small.jpg)

Oh My Zsh was started by the team at [Planet Argon](https://www.planetargon.com/?utm_source=github), a
[Ruby on Rails development consultancy](https://www.planetargon.com/services/ruby-on-rails-development?utm_source=github).
Check out our [other open source projects](https://www.planetargon.com/open-source?utm_source=github).

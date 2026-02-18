# Claude Switch

[![Go Version](https://img.shields.io/github/go-mod/go-version/caeser1996/claude-switch)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/caeser1996/claude-switch/ci.yml?branch=main)](https://github.com/caeser1996/claude-switch/actions)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey.svg)](https://github.com/caeser1996/claude-switch/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/caeser1996/claude-switch)](https://goreportcard.com/report/github.com/caeser1996/claude-switch)

**A single-binary CLI tool for managing multiple Claude Code accounts.**
No dependencies. No Node.js. No Python. Just download and run.

---

## Why?

If you use Claude Code with multiple accounts (work, personal, client projects), switching between them means manually copying credential files every time. **Claude Switch** makes it a one-command operation.

## Features

- **One-command switch** — `cs use work` and you're done
- **Profile management** — Import, list, remove profiles
- **Auto-backup** — Credentials are backed up before every switch
- **Health check** — `cs doctor` diagnoses your setup
- **Cross-platform** — Linux, macOS, Windows (amd64 & arm64)
- **Zero dependencies** — Single binary, no runtime required
- **Secure** — Credentials stored with `0600` permissions, no telemetry

## Installation

### One-liner (Linux & macOS — no Go required)

```bash
curl -fsSL https://raw.githubusercontent.com/caeser1996/claude-switch/main/scripts/install.sh | bash
```

Installs `claude-switch` and a `cs` shortcut to `/usr/local/bin`. Use `INSTALL_DIR` to override:

```bash
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/caeser1996/claude-switch/main/scripts/install.sh | bash
```

### Homebrew (macOS & Linux)

```bash
brew install caeser1996/tap/claude-switch
```

### Manual download

Download the pre-built binary for your platform from [GitHub Releases](https://github.com/caeser1996/claude-switch/releases):

| Platform | File |
|----------|------|
| Linux x86-64 | `claude-switch_X.Y.Z_linux_amd64.tar.gz` |
| Linux ARM64 | `claude-switch_X.Y.Z_linux_arm64.tar.gz` |
| macOS Intel | `claude-switch_X.Y.Z_darwin_amd64.tar.gz` |
| macOS Apple Silicon | `claude-switch_X.Y.Z_darwin_arm64.tar.gz` |
| Windows x86-64 | `claude-switch_X.Y.Z_windows_amd64.zip` |

```bash
tar -xzf claude-switch_*_linux_amd64.tar.gz
sudo mv claude-switch /usr/local/bin/
sudo ln -s /usr/local/bin/claude-switch /usr/local/bin/cs
```

### From source (Go 1.24+)

```bash
go install github.com/caeser1996/claude-switch@latest
```

### Build from source

```bash
git clone https://github.com/caeser1996/claude-switch.git
cd claude-switch
make build
# Binary is in ./bin/claude-switch
```

## Quick Start

```bash
# 1. Log into your work Claude account, then save it
cs import work -d "Work account"

# 2. Log into your personal account, then save it
cs import personal -d "Personal account"

# 3. Switch between them anytime
cs use work        # → Switched to "work"
cs use personal    # → Switched to "personal"

# 4. See all profiles
cs list
```

## Commands

### Account Management

| Command | Description |
|---------|-------------|
| `cs import <name>` | Save current Claude session as a named profile |
| `cs login <name>` | Run `claude login` then auto-import as a profile |
| `cs use <name>` | Switch to a profile (auto-detects `.claude-profile`) |
| `cs switch` | Interactive profile selector (TUI) |
| `cs list` | List all profiles (`*` = active) |
| `cs status` | Show auth status and token expiry for all profiles |
| `cs remove <name>` | Delete a profile |
| `cs current` | Show active profile name |
| `cs info <name>` | Detailed info about a profile |

### Execution

| Command | Description |
|---------|-------------|
| `cs exec <profile> -- <cmd>` | Run command with profile's credentials (no switch) |
| `cs limits` | Show usage limits for active profile |

### Sharing & Encryption

| Command | Description |
|---------|-------------|
| `cs export <name>` | Export profile as encrypted `.csprofile` file |
| `cs import-file <file>` | Import profile from encrypted file |

### Shell Integration

| Command | Description |
|---------|-------------|
| `cs alias` | Generate shell aliases (`claude-work`, `claude-personal`, etc.) |
| `cs completion <shell>` | Shell completions (bash, zsh, fish, powershell) |

### Maintenance

| Command | Description |
|---------|-------------|
| `cs doctor` | Run health checks on your installation |
| `cs backup list` | Show available backups |
| `cs backup restore <ts>` | Restore from a backup |
| `cs config show/edit/path` | View or edit configuration |
| `cs update` | Self-update to the latest release |
| `cs version` | Show version info |

### Flags

| Flag | Description |
|------|-------------|
| `--verbose, -v` | Verbose output |
| `--no-color` | Disable colored output |

## How It Works

Claude Switch stores copies of your Claude Code credential files in isolated profile directories:

```
~/.claude-switch/
├── config.json              # Which profile is active
├── profiles/
│   ├── work/
│   │   └── .credentials.json
│   └── personal/
│       └── .credentials.json
└── backups/
    └── 20260219-143022/     # Auto-backup before each switch
        └── .credentials.json
```

When you run `cs use work`:
1. Current credentials are backed up
2. Work profile credentials are copied to `~/.claude/`
3. Config is updated to mark "work" as active

That's it. No symlinks, no env vars, no magic.

## Per-Project Profiles

Create a `.claude-profile` file in any project directory:

```bash
echo "work" > ~/projects/company-app/.claude-profile
```

Now when you run `cs use` (with no argument) inside that directory, it automatically switches to the `work` profile. The file is searched up the directory tree, so it works in subdirectories too.

## Encrypted Profile Sharing

Share profiles securely between machines or team members:

```bash
# Export (encrypts with AES-256-GCM)
cs export work -o work.csprofile
# Enter passphrase: ********

# Import on another machine
cs import-file work.csprofile
# Enter passphrase: ********
```

## Shell Aliases

Generate convenience aliases for your shell:

```bash
cs alias >> ~/.bashrc    # or ~/.zshrc
source ~/.bashrc

# Now you can use:
claude-work              # switch to work + launch claude
claude-personal          # switch to personal + launch claude
cs-work                  # just switch to work
```

## Configuration

The config file lives at `~/.claude-switch/config.json`:

```json
{
  "active_profile": "work",
  "profiles": {
    "work": {
      "name": "work",
      "email": "you@company.com",
      "description": "Work account",
      "created_at": "2026-02-19T14:30:00Z",
      "is_active": true
    }
  },
  "settings": {
    "auto_backup": true,
    "max_backups": 10,
    "color_output": true
  }
}
```

## Security

- All credential files are stored with **`0600`** permissions (owner read/write only)
- Profile directories use **`0700`** permissions
- No credentials are printed or logged (even in verbose mode)
- Backups are auto-pruned (default: keep 10 most recent)
- **Encrypted exports** — AES-256-GCM encryption for shared profiles
- **Token expiry detection** — Warns when tokens are expired or expiring soon
- **No telemetry, no phone-home, fully open source**

## Development

```bash
# Run tests
make test

# Run linter
make lint

# Build for current platform
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

## Roadmap

- [x] `cs exec <profile> -- claude "prompt"` — Isolated parallel execution
- [x] `cs limits` — Usage/rate limit display
- [x] `cs completion bash/zsh/fish` — Shell completions
- [x] `cs backup list/restore` — Backup management
- [x] `cs update` — Self-update from GitHub releases
- [x] `cs login <name>` — Login and auto-import
- [x] `cs export / import-file` — Encrypted profile sharing (AES-256-GCM)
- [x] `cs status` — Token expiry and auth health
- [x] `cs switch` — Interactive TUI profile selector
- [x] `cs alias` — Shell alias generation
- [x] `.claude-profile` — Per-project auto-switch
- [x] Token refresh detection and warnings
- [ ] Git hook integration (auto-switch based on repo)
- [ ] Profile encryption at rest

## License

MIT License. See [LICENSE](LICENSE) for details.

## Author

**Sumanta Mukhopadhyay** — [GitHub](https://github.com/caeser1996)

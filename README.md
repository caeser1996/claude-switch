# Claude Switch

[![Go Version](https://img.shields.io/github/go-mod/go-version/caeser1996/claude-switch)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/caeser1996/claude-switch/ci.yml?branch=main)](https://github.com/caeser1996/claude-switch/actions)
[![Release](https://img.shields.io/github/v/release/caeser1996/claude-switch)](https://github.com/caeser1996/claude-switch/releases)
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

### From Source (Go 1.21+)

```bash
go install github.com/caeser1996/claude-switch@latest
```

### From Releases

Download the binary for your platform from [GitHub Releases](https://github.com/caeser1996/claude-switch/releases).

### Build from Source

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
| `cs use <name>` | Switch to a profile |
| `cs list` | List all profiles (`*` = active) |
| `cs remove <name>` | Delete a profile |
| `cs current` | Show active profile name |

### Diagnostics

| Command | Description |
|---------|-------------|
| `cs doctor` | Run health checks on your installation |
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

- [x] `cs exec <profile> -- claude "prompt"` — Run with a specific profile without switching
- [x] `cs limits` — Check usage/rate limits
- [x] `cs completion bash/zsh/fish` — Shell completions
- [x] `cs backup list/restore` — Manage backups
- [ ] `cs update` — Self-update from GitHub releases
- [ ] Profile encryption at rest

## License

MIT License. See [LICENSE](LICENSE) for details.

## Author

**Sumanta Mukhopadhyay** — [GitHub](https://github.com/caeser1996)

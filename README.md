# Terminal History Navigator

TUI application for browsing shell command history on macOS and Linux.

## Features

- Browse commands from zsh/bash history files
- Chronological ordering based on actual timestamps (zsh extended history)
- Accurate frequency tracking across sessions
- Search with whole-word and prefix matching
- Command templates with categories — add any command with one key, delete with one key
- Copy commands to clipboard
- Frequency and chronological sorting
- Exit code indicators (✓ / ✗) for zsh extended history
- Centered cursor scrolling

## Requirements

- macOS 10.12+ or Linux
- Go 1.21+ (for building only)
- Linux clipboard: `xclip` or `xsel`

## Installation

```bash
git clone https://github.com/4ndew/terminal-history-navigator
cd terminal-history-navigator
make install
```

Add alias to shell config:

```bash
echo 'alias h="terminal-history-navigator"' >> ~/.zshrc
source ~/.zshrc
```

## Usage

```bash
h
# or
terminal-history-navigator
```

## Key Bindings

### Navigation

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Enter` | Copy selected command to clipboard |
| `q` / `Ctrl+C` | Quit |

### Modes

| Key | Action |
|-----|--------|
| `h` | Switch to history mode |
| `t` | Toggle templates mode |
| `/` | Enter search mode |
| `f` | Toggle frequency / chronological sort (history mode) |
| `?` | Show help |

### Templates

| Key | Action |
|-----|--------|
| `s` | Save selected command as template (history / search mode) |
| `d` | Delete selected template (templates mode) |
| `Enter` | Copy template command to clipboard |

Category is detected automatically from the command prefix (git, docker, ssh, go, npm, etc.).

### Search

| Key | Action |
|-----|--------|
| Type | Filter as you type |
| `↑` / `↓` | Navigate results |
| `Enter` | Copy selected result |
| `Esc` | Exit search, return to history |
| `Backspace` | Delete last character |

Search matches commands containing all query words as whole words or prefixes. For example, `git c` matches `git clone` and `git commit` but not `git branch`.

## Configuration

Config files are created automatically on first run.

### Main config — `~/.config/history-nav/config.yaml`

```yaml
sources:
  - ~/.zsh_history
  - ~/.bash_history
exclude_patterns:
  - "^sudo su"
  - "password"
  - "token"
  - "secret"
  - "^exit$"
  - "^clear$"
  - "^\d+$"
ui:
  max_items: 1000
  show_timestamps: true
  show_frequency: true
performance:
  max_history_lines: 10000
```

### Templates — `~/.config/history-nav/templates.yaml`

```yaml
templates:
  - name: "Git log oneline"
    command: "git log --oneline -10"
    description: "Show last 10 commits"
    category: "git"
```

Templates can be managed directly from the TUI — no manual editing required.

## Recommended Shell Settings

For accurate timestamps and exit code tracking, add to `~/.zshrc`:

```bash
# Write timestamp and duration to each history entry
setopt extended_history

# Append to history immediately, share across sessions
setopt inc_append_history
setopt share_history
```

Without `extended_history`, commands are ordered by line number in the history file rather than by actual time. With it, ordering is based on unix timestamps and is accurate across multiple files and sessions.

## Troubleshooting

**No history showing**

Check that the files exist:
```bash
ls ~/.zsh_history ~/.bash_history
```

Force-save the current session before running:
```bash
fc -W        # zsh
history -a   # bash
```

**History not updating between sessions**

Add to `~/.zshrc`:
```bash
setopt inc_append_history
setopt share_history
```

**Exit code indicators not showing**

Requires `setopt extended_history` in `~/.zshrc`. Indicators show ✓ for exit code 0 and ✗ for anything else.

**Clipboard not working on Linux**

Install either `xclip` or `xsel`:
```bash
sudo apt install xclip   # Debian/Ubuntu
sudo dnf install xclip   # Fedora
brew install xclip       # Homebrew on Linux
```

## Development

```bash
make build    # Build binary to bin/
make run      # Build and run
make clean    # Remove build artifacts
```

## Manual Installation

```bash
make build
sudo cp bin/terminal-history-navigator /usr/local/bin/
make setup-config
```

## License

MIT
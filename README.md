# Terminal History Navigator

TUI application for browsing shell command history on macOS and Linux.

## Features

- Browse commands from zsh/bash history files
- Chronological ordering based on actual timestamps (zsh `extended_history`, bash `HISTTIMEFORMAT`)
- Files without timestamps are merged using approximate times derived from file modification time
- Multiline commands (backslash continuations) are kept as single entries and copied with line breaks preserved
- Accurate frequency tracking across sessions (real usage counts, no heuristics)
- Search with whole-word and prefix matching, Unicode-aware
- Command templates with categories — add any command with one key, delete with one key
- Copy commands to clipboard
- Frequency / chronological sorting toggle
- Mouse wheel scrolling
- Centered cursor scrolling

## Requirements

- macOS 10.12+ or Linux
- Go 1.21+ (for building only)
- Linux clipboard: `xclip` or `xsel`

## Installation

### From source

```bash
git clone https://github.com/4ndew/terminal-history-navigator
cd terminal-history-navigator
make install
```

### From GitHub Releases

Download the archive for your OS/arch from the Releases page, unpack and put the binary into your `PATH`:

```bash
tar -xzf terminal-history-navigator_<os>_<arch>.tar.gz
sudo mv terminal-history-navigator /usr/local/bin/
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
| Mouse wheel | Scroll list |
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
  - "^\\d+$"
ui:
  max_items: 1000
  show_timestamps: true
  show_frequency: true
performance:
  max_history_lines: 10000
```

Note: exclude patterns are regular expressions. Anchor them (`^...$`) when you
mean an exact command, otherwise the pattern matches as a substring.

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

### zsh — add to `~/.zshrc`:

```bash
# Write timestamp and duration to each history entry
setopt extended_history

# Append to history immediately, share across sessions
setopt inc_append_history
setopt share_history
```

### bash — add to `~/.bashrc`:

```bash
# Store timestamps in history
export HISTTIMEFORMAT="%F %T "

# Append to history file on each command
shopt -s histappend
export PROMPT_COMMAND="history -a; $PROMPT_COMMAND"
```

Without timestamps, command order is approximated from the history file's
modification time and line order. With timestamps, ordering is exact across
multiple files and sessions.

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

**Ordering looks wrong when mixing zsh and bash history**

Enable timestamps in both shells (see Recommended Shell Settings). Entries
without timestamps can only be ordered approximately.

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

## Releases

Binaries are built by GoReleaser in CI:

- every push to `main` produces snapshot binaries (Actions → Artifacts)
- pushing a tag `v*` publishes a GitHub Release with archives for
  linux/darwin/windows × amd64/arm64

```bash
git tag v0.1.0
git push origin v0.1.0
```

## Manual Installation

```bash
make build
sudo cp bin/terminal-history-navigator /usr/local/bin/
make setup-config
```

## License

MIT
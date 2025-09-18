# Terminal History Navigator

A fast TUI (Terminal User Interface) application for browsing and using your shell command history on macOS/Linux.

## Features

- **Browse History**: View commands from zsh/bash history files
- **Smart Search**: Find commands with fuzzy search
- **Command Templates**: Predefined commands with descriptions
- **One-Click Copy**: Copy commands to clipboard instantly
- **Frequency Sorting**: See most-used commands first
- **Clean Interface**: Keyboard-driven TUI with vim-like navigation

## Quick Start

### 1. Install
```bash
git clone https://github.com/4ndew/terminal-history-navigator
cd terminal-history-navigator
make install
```

### 2. Setup Shell Alias
Add to your `~/.zshrc` or `~/.bashrc`:
```bash
alias h='terminal-history-navigator'
```
Reload: `source ~/.zshrc`

### 3. Run
```bash
h
# or full command name
terminal-history-navigator
```

## Usage

### Navigation
| Key | Action |
|-----|--------|
| `↑/k` | Move up |
| `↓/j` | Move down |
| `Enter` | Copy command to clipboard |
| `q` | Quit |

### Features  
| Key | Action |
|-----|--------|
| `/` | Search commands |
| `t` | Toggle templates mode |
| `e` | Edit templates (in templates mode) |
| `f` | Sort by frequency |
| `r` | Refresh data |
| `?` | Show help |

### Search Mode
| Key | Action |
|-----|--------|
| Type | Search as you type |
| `↑/↓` | Navigate results |
| `Enter` | Select result |
| `Esc` | Exit search |

## Configuration

On first run, config files are created:

**Main config**: `~/.config/history-nav/config.yaml`
```yaml
sources:
  - ~/.zsh_history
  - ~/.bash_history
exclude_patterns:
  - "^sudo "
  - "password"
ui:
  max_items: 1000
```

**Templates**: `~/.config/history-nav/templates.yaml`
```yaml
templates:
  - name: "Git status"
    command: "git status" 
    description: "Show working tree status"
    category: "git"
```

## Manual Installation Steps

If `make install` doesn't work:

```bash
# 1. Build
make build

# 2. Copy binary
sudo cp bin/terminal-history-navigator /usr/local/bin/

# 3. Create config
make setup-config

# 4. Add alias to shell config
echo 'alias h="terminal-history-navigator"' >> ~/.zshrc
source ~/.zshrc
```

## Troubleshooting

**No history showing?**
- Check history files exist: `ls ~/.zsh_history ~/.bash_history`
- Verify config sources in `~/.config/history-nav/config.yaml`
- Force save current session: `fc -W` (zsh) or `history -a` (bash)

**History not updating in real-time?**
- By default, zsh only saves history on session exit
- Force save current session: `fc -W` (zsh) or `history -a` (bash)
- For real-time history (optional): add to `~/.zshrc`:
  ```bash
  setopt inc_append_history
  setopt share_history
  ```

**Want to see command success/failure status?**
- For enhanced history with exit codes: add to `~/.zshrc`:
  ```bash
  setopt extended_history
  ```
- Commands will show ✓ (success) or ✗ (failed) indicators

**Commands not being saved?**
- Check if commands are being filtered by settings in `~/.zshrc`
- Some commands may be excluded by `HIST_IGNORE_SPACE` or similar options

**Clipboard not working?**
- macOS: Should work out of box
- Linux: Install `xclip` or `xsel`

**Permission denied?**
- Make sure binary is executable: `chmod +x /usr/local/bin/terminal-history-navigator`

## Development

```bash
# Build
make build

# Test run  
make run

# Install locally
make install

# Clean
make clean
```

## Requirements

- macOS 10.12+ or Linux
- Go 1.21+ (for building from source)

## License

MIT
# Terminal History Navigator

TUI application for browsing shell command history on macOS/Linux.

## Features

- Browse commands from zsh/bash history files
- Search commands with whole word/prefix matching
- Command templates with descriptions
- Copy commands to clipboard
- Frequency and chronological sorting
- Command success/failure indicators
- Automatic filtering of problematic commands
- Multi-line command support

## Installation

```bash
git clone https://github.com/4ndew/terminal-history-navigator
cd terminal-history-navigator
make install
```

Add to shell config:
```bash
alias h='terminal-history-navigator'
```
Reload: `source ~/.zshrc`

Run
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

### Modes
| Key | Action |
|-----|--------|
| `t` | Toggle templates mode |
| `/` | Search mode |
| `f` | Sort by frequency |
| `?` | Show help |

### Search
| Key | Action |
|-----|--------|
| Type | Search as you type |
| `↑/↓` | Navigate results |
| `Enter` | Select result |
| `Esc` | Exit search |
| `Backspace` | Delete character |

Search finds commands containing all query words as whole words or prefixes. Query "git c" matches "git clone", "git commit" but not "git branch".

## Configuration

Config files created on first run:

**Main config**: `~/.config/history-nav/config.yaml`
```yaml
sources:
  - ~/.zsh_history
  - ~/.bash_history
exclude_patterns:
  - "^sudo su"
  - "password"
  - "token"
  - "^exit$"
  - "^\\d+$"
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

## Manual Installation

```bash
make build
sudo cp bin/terminal-history-navigator /usr/local/bin/
make setup-config
echo 'alias h="terminal-history-navigator"' >> ~/.zshrc
```

## Troubleshooting

**No history showing:**
- Check files exist: `ls ~/.zsh_history ~/.bash_history`
- Verify config sources
- Force save: `fc -W` (zsh) or `history -a` (bash)

**History not updating:**
- Force save current session: `fc -W` (zsh) or `history -a` (bash)
- For real-time updates add to `~/.zshrc`:
  ```bash
  setopt inc_append_history
  setopt share_history
  ```

**Command status indicators:**
- For exit code tracking add to `~/.zshrc`:
  ```bash
  setopt extended_history
  ```
- Shows ✓ (success) or ✗ (failed) for commands

**Clipboard issues:**
- macOS: Works by default
- Linux: Install `xclip` or `xsel`

## Development

```bash
make build    # Build binary
make run      # Test run
make clean    # Clean artifacts
```

## Requirements

- macOS 10.12+ or Linux
- Go 1.21+ (for building)

## License

MIT
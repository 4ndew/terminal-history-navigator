# History Navigator

Terminal history viewer with TUI interface for macOS.

## Features

- View zsh/bash command history
- Search and filter commands
- Command templates with descriptions
- Copy to clipboard
- Frequency-based sorting

## Installation

```bash
git clone https://github.com/4ndew/terminal-history-navigator
cd terminal-history-navigator
make install
```

## Usage

```bash
history-navigator
# or use alias
h
```

### Controls

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate |
| `Enter` | Copy to clipboard |
| `/` | Search |
| `Esc` | Exit search |
| `q` | Quit |
| `t` | Toggle templates |

## Configuration

Config file: `~/.config/history-nav/config.yaml`

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

## Shell Integration

Add to `~/.zshrc` or `~/.bashrc`:
```bash
alias h='history-navigator'
```

## Requirements

- macOS 10.12+
- Go 1.21+ (for building)

## License

MIT
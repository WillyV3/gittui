# gittui

![gittui-demo](https://github.com/user-attachments/assets/30db2955-d60a-4d55-bec8-b158e15b988b)


A terminal user interface for viewing GitHub profiles, contributions, and activity.

## Features

- View GitHub profile information and statistics
- Colorized braille avatar display
- Interactive contribution graph (52-week GitHub-style heatmap)
- Top programming languages with visual breakdown
- Contribution streak tracking
- **Activity Metrics** - Push rate tracking and Peak Coding Hour analysis
- Top repositories sorted by stars
- Recent activity timeline with colorized table view
- Toggle between public-only and all repositories (for authenticated users)
- **361 built-in themes** from the Gogh collection
- Fully responsive terminal layout
- ASCII art username display

## Installation

### Homebrew

```bash
brew install willyv3/tap/gittui
```

### Pre-built Binaries

Download the latest release for your platform from [GitHub Releases](https://github.com/willyv3/gittui/releases/latest).

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/willyv3/gittui/releases/latest/download/gittui-darwin-arm64
chmod +x gittui-darwin-arm64
mv gittui-darwin-arm64 /usr/local/bin/gittui

# macOS (Intel)
curl -LO https://github.com/willyv3/gittui/releases/latest/download/gittui-darwin-amd64
chmod +x gittui-darwin-amd64
mv gittui-darwin-amd64 /usr/local/bin/gittui

# Linux (x86_64)
curl -LO https://github.com/willyv3/gittui/releases/latest/download/gittui-linux-amd64
chmod +x gittui-linux-amd64
sudo mv gittui-linux-amd64 /usr/local/bin/gittui

# Linux (ARM64)
curl -LO https://github.com/willyv3/gittui/releases/latest/download/gittui-linux-arm64
chmod +x gittui-linux-arm64
sudo mv gittui-linux-arm64 /usr/local/bin/gittui

# Windows (x86_64) - PowerShell
Invoke-WebRequest -Uri "https://github.com/willyv3/gittui/releases/latest/download/gittui-windows-x86_64.zip" -OutFile "gittui.zip"
Expand-Archive gittui.zip -DestinationPath .
Move-Item gittui.exe C:\Windows\System32\gittui.exe

# Windows (ARM64) - PowerShell
Invoke-WebRequest -Uri "https://github.com/willyv3/gittui/releases/latest/download/gittui-windows-arm64.zip" -OutFile "gittui.zip"
Expand-Archive gittui.zip -DestinationPath .
Move-Item gittui.exe C:\Windows\System32\gittui.exe
```

### From Source

```bash
go install github.com/willyv3/gittui@latest
```

Or clone and build:

```bash
git clone https://github.com/willyv3/gittui.git
cd gittui
go build -o gittui .
```

## Usage

View your own profile (requires GitHub CLI authentication):

```bash
gittui
```

View another user's profile:

```bash
gittui username
```

### Authentication

gittui uses the GitHub CLI for authentication:

```bash
gh auth login
```

This allows you to:
- View your private repositories and contributions
- Access organization repositories you're a member of
- See private activity

### Keyboard Controls

- `q` or `Ctrl+C` - Quit
- `r` - Refresh all data
- `t` - Cycle through themes
- `p` - Toggle between public-only and all repositories (own profile only)
- `↑↓` or `j/k` - Scroll activity timeline

## Requirements

- Go 1.23 or later (for building from source)
- GitHub CLI (`gh`) installed and authenticated
- Terminal with color support
- Minimum terminal width: 108 columns for full graph display

## Architecture

Built with:
- [Charmbracelet](https://github.com/charmbracelet) ecosystem - Terminal UI framework
- GH CLI

Color schemes sourced from [Gogh](https://github.com/Gogh-Co/Gogh) terminal theme collection.

I used their theme files to make a super fast gogh-themes package for the app https://github.com/willyv3/gogh-themes

## Development

Run tests:

```bash
go test ./...
```

Build binary:

```bash
go build -o gittui .
```

Run locally:

```bash
./gittui
```

## License

MIT

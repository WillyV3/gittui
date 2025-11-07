# gittui

A terminal UI for GitHub profiles. Shows contribution graphs, stats, and activity. That's it.

## Install

### Homebrew

```sh
brew install willyv3/tap/gittui
```

### From source

```sh
git clone https://github.com/willyv3/gittui.git
cd gittui
make install
```

Installs to `~/.local/bin`. Add it to your PATH if you haven't already.

### Pre-built binaries

Get them from [releases](https://github.com/willyv3/gittui/releases/latest).

## Usage

```sh
gittui username
```

Or just `gittui` if you've authenticated with `gh auth login`.

## Configuration

Run `make config` or edit `~/.local/gittui/config.json` directly.

```json
{
  "username": "torvalds",
  "theme": "Dracula",
  "avatar_width": 60,
  "avatar_source": "personal",
  "avatar_renderer": "mosaic",
  "allowed_orgs": [],
  "include_personal_private": false
}
```

Config is optional. Defaults work fine.

## Features

- 52-week contribution graph
- Top languages, repos by stars
- Activity metrics (push rate, peak coding hour)
- Contribution streaks
- Recent activity with scrollable modal
- 361 themes from [Gogh](https://github.com/Gogh-Co/Gogh)
- ASCII avatar (two renderers available)
- Private repo support via `gh` CLI
- Organization repo filtering
- Responsive layout for small terminals
- Help modal with all keybindings

## Keybindings

Press `?` or `h` for help modal with all controls.

Common:
- `q` - quit
- `r` - refresh
- `t` - cycle themes
- `a` - activity modal
- `?` - help

See help modal for the rest.

## Requirements

- Terminal with 24-bit color
- `gh` CLI (for private repos)
- Go 1.23+ (building from source)
- `gum`, `jq` (for config script)

## Authentication

```sh
gh auth login
```

Without auth, you'll only see public data.

## Notes

The contribution graph needs ~108 columns minimum. Smaller terminals get a compact view.

Avatar rendering uses either `mosaic` (half-block, supports backgrounds) or `ascii-image-converter` (traditional ASCII art). Both look terrible but that's terminal graphics for you.

## License

MIT

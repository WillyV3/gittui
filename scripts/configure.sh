#!/bin/bash
# gittui setup script - Interactive configuration using gh CLI and gum
set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

CONFIG_DIR="$HOME/.local/gittui"
CONFIG_FILE="$CONFIG_DIR/config.json"

echo "🚀 gittui Setup"
echo ""

# Check for required tools
echo "Checking dependencies..."

MISSING_DEPS=false

if ! command -v gh &> /dev/null; then
    echo -e "${RED}✗${NC} GitHub CLI (gh) is not installed"
    MISSING_DEPS=true
fi

if ! command -v gum &> /dev/null; then
    echo -e "${RED}✗${NC} gum is not installed"
    MISSING_DEPS=true
fi

if ! command -v jq &> /dev/null; then
    echo -e "${RED}✗${NC} jq is not installed"
    MISSING_DEPS=true
fi

if [ "$MISSING_DEPS" = true ]; then
    echo ""
    echo "Run 'make deps' to install dependencies, or install manually:"
    echo "  gh:  https://cli.github.com"
    echo "  gum: https://github.com/charmbracelet/gum"
    echo "  jq:  https://jqlang.github.io/jq"
    exit 1
fi

echo -e "${GREEN}✓${NC} All dependencies found"
echo ""

# Check gh auth status
echo "🔐 Checking GitHub authentication..."
if ! gh auth status &> /dev/null; then
    echo -e "${RED}✗${NC} Not authenticated with GitHub CLI"
    echo ""
    gum confirm "Would you like to authenticate now?" && gh auth login || exit 1
    echo ""
fi

# Get authenticated user
GH_USER=$(gh api user -q .login)
echo -e "${GREEN}✓${NC} Authenticated as: ${GREEN}$GH_USER${NC}"
echo ""

# Welcome message
gum style \
    --foreground 212 --border-foreground 212 --border double \
    --align center --width 60 --margin "1 2" --padding "1 2" \
    "gittui Configuration" "" "Let's set up your profile preferences"

echo ""

# Explain private repo access
gum style \
    --foreground 220 --margin "0 2" \
    "ℹ️  About Private Repository Access" \
    "" \
    "By default, gittui shows PUBLIC data only." \
    "" \
    "Since you're authenticated with gh CLI, gittui will automatically" \
    "access private repos based on your configuration below." \
    "" \
    "Configure which orgs/repos to include:" \
    "  • Personal private repos" \
    "  • Organization private repos" \
    "" \
    "Note: Some organizations hide activity even for members." \
    "This is an org setting outside your control."

echo ""
echo ""

# Check if authenticated with gh CLI (go-gh handles this automatically)
# No need to set GITHUB_TOKEN manually - go-gh library does it for us
if gh auth status &> /dev/null; then
    echo -e "${GREEN}✓${NC} GITHUB_TOKEN is set"
    echo ""

    # Ask about personal private repos
    gum style \
        --foreground 212 --bold \
        "📦 Personal Private Repositories"
    echo ""

    if gum confirm "Include your personal private repositories?"; then
        INCLUDE_PERSONAL_PRIVATE="true"
    else
        INCLUDE_PERSONAL_PRIVATE="false"
    fi

    echo ""
    echo ""

    # Fetch user's organizations
    gum style \
        --foreground 212 --bold \
        "🏢 Organization Access"
    echo ""
    echo "Fetching your organizations..."

    # Get orgs with proper error handling
    ORGS_JSON=$(gh api user/orgs --paginate 2>/dev/null || echo "[]")
    ORG_COUNT=$(echo "$ORGS_JSON" | jq '. | length')

    if [ "$ORG_COUNT" -eq 0 ]; then
        echo "No organizations found."
        ALLOWED_ORGS="[]"
    else
        echo "Found $ORG_COUNT organization(s)"
        echo ""

        # Create list of org names
        ORG_NAMES=$(echo "$ORGS_JSON" | jq -r '.[].login' | sort)

        # Show explanation
        gum style \
            --foreground 220 --margin "0 0" \
            "Select organizations to include PRIVATE repos from:" \
            "" \
            "Use SPACE to select, ENTER when done" \
            "(Leave all unchecked to skip org repos)"

        echo ""

        # Let user select orgs using gum choose with multiselect
        SELECTED_ORGS=$(echo "$ORG_NAMES" | gum choose --no-limit --height 15 --header "Select organizations (space to select, enter to confirm):" || echo "")

        if [ -z "$SELECTED_ORGS" ]; then
            echo "No organizations selected."
            ALLOWED_ORGS="[]"
        else
            # Convert to JSON array
            ALLOWED_ORGS=$(echo "$SELECTED_ORGS" | jq -R -s -c 'split("\n") | map(select(length > 0))')

            echo ""
            echo -e "${GREEN}✓${NC} Selected organizations:"
            echo "$SELECTED_ORGS" | while read -r org; do
                echo "  • $org"
            done
        fi
    fi

    INCLUDE_PRIVATE="true"
fi

echo ""
echo ""

# Theme selection
gum style \
    --foreground 212 --bold \
    "🎨 Theme Selection"
echo ""
echo "You can change themes anytime in the app:"
echo "  • Press 't' to cycle through themes"
echo "  • Press 'T' (Shift+t) to save current as default"
echo ""

THEME=$(gum input --placeholder "Enter theme name (default: Dracula)" --value "Dracula")
if [ -z "$THEME" ]; then
    THEME="Dracula"
fi

echo ""
echo ""

# Summary
gum style \
    --foreground 212 --border-foreground 212 --border rounded \
    --margin "1 2" --padding "1 2" \
    "📋 Configuration Summary" \
    "" \
    "Username: $GH_USER" \
    "Theme: $THEME" \
    "Personal Private Repos: $INCLUDE_PERSONAL_PRIVATE" \
    "Organization Private Repos: $(echo $ALLOWED_ORGS | jq -r 'if . == [] then "None" else . | join(", ") end')"

echo ""

if ! gum confirm "Save this configuration?"; then
    echo "Setup cancelled."
    exit 0
fi

# Create config directory
mkdir -p "$CONFIG_DIR"

# Build JSON config
cat > "$CONFIG_FILE" <<EOF
{
  "username": "$GH_USER",
  "theme": "$THEME",
  "avatar_width": 60,
  "avatar_source": "personal",
  "avatar_renderer": "mosaic",
  "allowed_orgs": $ALLOWED_ORGS,
  "include_personal_private": $INCLUDE_PERSONAL_PRIVATE
}
EOF

echo ""
echo -e "${GREEN}✓${NC} Configuration saved to: $CONFIG_FILE"
echo ""

# Final instructions
gum style \
    --foreground 212 --border-foreground 212 --border double \
    --margin "1 2" --padding "1 2" \
    "🎉 Setup Complete!" \
    "" \
    "Run gittui with:" \
    "  gittui" \
    "" \
    "Your username is saved in config." \
    "You can override with:" \
    "  gittui <other-username>"

echo ""

gum confirm "Launch gittui now?" && exec gittui || echo "Goodbye! 👋"

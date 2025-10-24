#!/bin/bash

# Full Release Pipeline for gittui
# Just code and run - handles everything automatically

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Configuration
REPO_ORG="willyv3"
REPO_NAME="gittui"
HOMEBREW_TAP_PATH="$HOME/homebrew-tap"
FORMULA_PATH="$HOMEBREW_TAP_PATH/Formula/gittui.rb"

# Function to print colored output
print_step() {
    echo -e "${BLUE}→${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Get current version from tags
get_current_version() {
    git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"
}

# Calculate next version
get_next_version() {
    local current=$1
    local bump_type=$2

    # Remove 'v' prefix
    version=${current#v}

    # Split into parts
    IFS='.' read -r major minor patch <<< "$version"

    case $bump_type in
        major)
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        minor)
            minor=$((minor + 1))
            patch=0
            ;;
        patch)
            patch=$((patch + 1))
            ;;
        *)
            echo "$current"
            return
            ;;
    esac

    echo "v${major}.${minor}.${patch}"
}

# Generate changelog from commits
generate_changelog() {
    local from_tag=$1
    local to_tag=$2

    echo "## What's Changed"
    echo ""

    # Group commits by type
    local features=""
    local fixes=""
    local other=""

    while IFS= read -r commit; do
        if [[ $commit == *"feat:"* ]] || [[ $commit == *"add:"* ]] || [[ $commit == *"Add"* ]]; then
            features="${features}- ${commit}\n"
        elif [[ $commit == *"fix:"* ]] || [[ $commit == *"Fix"* ]]; then
            fixes="${fixes}- ${commit}\n"
        else
            other="${other}- ${commit}\n"
        fi
    done < <(git log ${from_tag}..HEAD --pretty=format:"%s" --no-merges)

    if [[ -n $features ]]; then
        echo "### Features"
        echo -e "$features"
    fi

    if [[ -n $fixes ]]; then
        echo "### Bug Fixes"
        echo -e "$fixes"
    fi

    if [[ -n $other ]]; then
        echo "### Other Changes"
        echo -e "$other"
    fi

    echo ""
    echo "**Full Changelog**: https://github.com/${REPO_ORG}/${REPO_NAME}/compare/${from_tag}...${to_tag}"
}

# Main release process
main() {
    echo ""
    echo "========================================="
    echo "   gittui Release Pipeline"
    echo "========================================="
    echo ""

    # Parse arguments
    BUMP_TYPE=${1:-patch}
    CUSTOM_MESSAGE=${2:-""}

    if [[ $BUMP_TYPE != "major" && $BUMP_TYPE != "minor" && $BUMP_TYPE != "patch" ]]; then
        print_error "Invalid bump type. Use: major, minor, or patch"
    fi

    # Step 1: Commit any pending changes
    print_step "Checking for changes..."
    if [[ -n $(git status -s) ]]; then
        print_warning "Uncommitted changes found. Adding and committing..."

        git add -A

        if [[ -n $CUSTOM_MESSAGE ]]; then
            git commit -m "$CUSTOM_MESSAGE"
        else
            # Generate commit message from changes
            COMMIT_MSG="Release prep: $(git diff --cached --name-only | head -3 | xargs basename | paste -sd ', ' -)"
            git commit -m "$COMMIT_MSG"
        fi
        print_success "Changes committed"
    else
        print_success "No uncommitted changes"
    fi

    # Step 2: Get version info
    CURRENT_VERSION=$(get_current_version)
    NEW_VERSION=$(get_next_version "$CURRENT_VERSION" "$BUMP_TYPE")

    echo ""
    print_step "Current version: ${CURRENT_VERSION}"
    print_step "New version: ${NEW_VERSION}"
    echo ""

    # Step 3: Build Go binary to ensure it compiles
    print_step "Building gittui..."
    go build -o gittui .
    print_success "Build successful"

    # Step 4: Create and push tag
    print_step "Creating git tag ${NEW_VERSION}..."
    git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"

    print_step "Pushing to GitHub..."
    git push origin main
    git push origin "$NEW_VERSION"
    print_success "Tag and code pushed to GitHub"

    # Step 5: Calculate SHA256 for new tarball
    print_step "Waiting for GitHub to process tag..."
    sleep 5

    TARBALL_URL="https://github.com/${REPO_ORG}/${REPO_NAME}/archive/${NEW_VERSION}.tar.gz"
    print_step "Calculating SHA256 for ${TARBALL_URL}..."

    SHA256=$(curl -sL "$TARBALL_URL" | shasum -a 256 | cut -d' ' -f1)

    if [[ -z $SHA256 ]]; then
        print_error "Failed to download tarball or calculate SHA256"
    fi

    print_success "SHA256: ${SHA256}"

    # Step 6: Update formula with new URL and SHA256
    print_step "Updating Homebrew formula..."

    if [[ ! -f $FORMULA_PATH ]]; then
        print_error "Formula not found: $FORMULA_PATH"
    fi

    # Update URL and SHA256 in the tap formula
    sed -i '' "s|url \".*\"|url \"${TARBALL_URL}\"|" "$FORMULA_PATH"
    sed -i '' "s|sha256 \".*\"|sha256 \"${SHA256}\"|" "$FORMULA_PATH"

    # Commit and push to homebrew tap
    cd "$HOMEBREW_TAP_PATH"
    git add Formula/gittui.rb
    git commit -m "Release gittui ${NEW_VERSION}"
    git push
    cd - > /dev/null

    print_success "Homebrew formula updated in tap"

    # Step 7: Generate changelog
    print_step "Generating changelog..."
    CHANGELOG=$(generate_changelog "$CURRENT_VERSION" "$NEW_VERSION")

    # Step 8: Create GitHub release
    print_step "Creating GitHub release..."

    gh release create "$NEW_VERSION" \
        --repo "${REPO_ORG}/${REPO_NAME}" \
        --title "Release ${NEW_VERSION}" \
        --notes "$CHANGELOG" \
        --latest

    print_success "GitHub release created"

    echo ""
    echo "========================================="
    echo -e "${GREEN}✓ Release ${NEW_VERSION} complete!${NC}"
    echo "========================================="
    echo ""
    echo "Users can now install/upgrade with:"
    echo "  brew install ${REPO_ORG}/tap/gittui"
    echo "  brew upgrade gittui"
    echo ""
    echo "Release URL:"
    echo "  https://github.com/${REPO_ORG}/${REPO_NAME}/releases/tag/${NEW_VERSION}"
    echo ""
}

# Run main function
main "$@"

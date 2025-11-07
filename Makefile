.PHONY: build install uninstall clean test config deps help

# Variables
BINARY_NAME=gittui
CONFIG_SCRIPT=scripts/configure.sh
INSTALL_PATH=$(HOME)/.local/bin
CONFIG_DIR=$(HOME)/.local/gittui

# Default target
help:
	@echo "GitTUI - GitHub TUI Dashboard"
	@echo ""
	@echo "Usage:"
	@echo "  make build      - Build the binary"
	@echo "  make install    - Install binary and config script to ~/.local/bin"
	@echo "  make uninstall  - Remove installed files"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make deps       - Install required dependencies (gh, gum, jq)"
	@echo "  make config     - Run the interactive configuration tool"
	@echo ""
	@echo "After installation:"
	@echo "  gittui <username>       - Run the TUI"
	@echo "  gittui-config           - Configure settings interactively"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd/gittui
	@echo "✓ Built $(BINARY_NAME)"

# Install binary and scripts
install: build
	@echo "Installing $(BINARY_NAME)..."
	@mkdir -p $(INSTALL_PATH)
	@mkdir -p $(CONFIG_DIR)
	@cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	@cp $(CONFIG_SCRIPT) $(INSTALL_PATH)/gittui-config
	@chmod +x $(INSTALL_PATH)/gittui-config
	@echo "✓ Installed $(BINARY_NAME) to $(INSTALL_PATH)"
	@echo "✓ Installed gittui-config to $(INSTALL_PATH)"
	@echo ""
	@echo "Make sure $(INSTALL_PATH) is in your PATH:"
	@echo "  export PATH=\"\$$HOME/.local/bin:\$$PATH\""
	@echo ""
	@echo "Run 'gittui-config' to configure your settings"

# Uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	@rm -f $(INSTALL_PATH)/$(BINARY_NAME)
	@rm -f $(INSTALL_PATH)/gittui-config
	@echo "✓ Uninstalled from $(INSTALL_PATH)"
	@echo ""
	@echo "Config files remain in $(CONFIG_DIR)"
	@echo "To remove config: rm -rf $(CONFIG_DIR)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@go clean
	@echo "✓ Cleaned"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Install dependencies (gh, gum, jq)
deps:
	@echo "Installing dependencies..."
	@if command -v brew >/dev/null 2>&1; then \
		echo "Using Homebrew..."; \
		brew install gh gum jq; \
		echo "✓ Dependencies installed"; \
	elif command -v apt-get >/dev/null 2>&1; then \
		echo "Using apt..."; \
		sudo apt-get update && sudo apt-get install -y jq; \
		echo "Installing gh CLI..."; \
		curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg; \
		echo "deb [arch=$$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null; \
		sudo apt-get update && sudo apt-get install -y gh; \
		echo "Installing gum..."; \
		sudo mkdir -p /etc/apt/keyrings; \
		curl -fsSL https://repo.charm.sh/apt/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/charm.gpg; \
		echo "deb [signed-by=/etc/apt/keyrings/charm.gpg] https://repo.charm.sh/apt/ * *" | sudo tee /etc/apt/sources.list.d/charm.list; \
		sudo apt-get update && sudo apt-get install -y gum; \
		echo "✓ Dependencies installed"; \
	elif command -v yum >/dev/null 2>&1; then \
		echo "Using yum..."; \
		sudo yum install -y jq; \
		echo "Installing gh CLI..."; \
		sudo yum-config-manager --add-repo https://cli.github.com/packages/rpm/gh-cli.repo; \
		sudo yum install -y gh; \
		echo "Installing gum..."; \
		echo '[charm]' | sudo tee /etc/yum.repos.d/charm.repo; \
		echo 'name=Charm' | sudo tee -a /etc/yum.repos.d/charm.repo; \
		echo 'baseurl=https://repo.charm.sh/yum/' | sudo tee -a /etc/yum.repos.d/charm.repo; \
		echo 'enabled=1' | sudo tee -a /etc/yum.repos.d/charm.repo; \
		echo 'gpgcheck=1' | sudo tee -a /etc/yum.repos.d/charm.repo; \
		echo 'gpgkey=https://repo.charm.sh/yum/gpg.key' | sudo tee -a /etc/yum.repos.d/charm.repo; \
		sudo yum install -y gum; \
		echo "✓ Dependencies installed"; \
	elif command -v pacman >/dev/null 2>&1; then \
		echo "Using pacman..."; \
		sudo pacman -Sy --noconfirm github-cli gum jq; \
		echo "✓ Dependencies installed"; \
	else \
		echo "❌ Could not detect package manager"; \
		echo ""; \
		echo "Please install manually:"; \
		echo "  gh:  https://cli.github.com"; \
		echo "  gum: https://github.com/charmbracelet/gum#installation"; \
		echo "  jq:  https://jqlang.github.io/jq/download/"; \
		exit 1; \
	fi

# Run the configuration tool
config:
	@./$(CONFIG_SCRIPT)

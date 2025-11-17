#!/bin/bash
# Automated GitHub Actions Runner Setup
# Uses gh CLI to automatically generate tokens and configure runners
# Works for both repository-level and organization-level runners

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
cat << 'EOF'
╔═══════════════════════════════════════════════════════════════════╗
║                                                                   ║
║   Automated GitHub Actions Runner Setup                          ║
║   Uses gh CLI to generate tokens automatically                   ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}"

# Check for gh CLI
if ! command -v gh &> /dev/null; then
    echo -e "${RED}❌ gh CLI not found${NC}"
    echo ""
    echo "Install with:"
    echo "  brew install gh"
    echo "  # or download from: https://cli.github.com/"
    exit 1
fi
echo -e "${GREEN}✅ gh CLI found${NC}"

# Check authentication
echo ""
echo "Checking GitHub authentication..."
if ! gh auth status &> /dev/null; then
    echo -e "${RED}❌ Not authenticated with GitHub${NC}"
    echo ""
    echo "Run: gh auth login"
    exit 1
fi
echo -e "${GREEN}✅ Authenticated with GitHub${NC}"

# Get user info
GITHUB_USER=$(gh api user -q .login)
echo "Logged in as: ${GITHUB_USER}"
echo ""

# Setup type selection
echo "Setup Type:"
echo "  1) Organization-level (ONE token, works for ALL repos)"
echo "  2) Repository-level (one token per repo)"
echo ""
read -p "Choose [1 or 2]: " SETUP_TYPE

if [ "$SETUP_TYPE" = "1" ]; then
    # Organization setup
    echo ""
    echo "Available organizations:"
    gh api user/orgs --jq '.[].login' | nl -w2 -s'. '
    echo ""
    read -p "Enter organization name: " ORG_NAME
    
    if [ -z "$ORG_NAME" ]; then
        echo -e "${RED}❌ Organization name required${NC}"
        exit 1
    fi
    
    URL="https://github.com/${ORG_NAME}"
    SCOPE="org"
    
elif [ "$SETUP_TYPE" = "2" ]; then
    # Repository setup
    echo ""
    read -p "Repository owner: " OWNER
    read -p "Repository name: " REPO
    
    if [ -z "$OWNER" ] || [ -z "$REPO" ]; then
        echo -e "${RED}❌ Owner and repo name required${NC}"
        exit 1
    fi
    
    URL="https://github.com/${OWNER}/${REPO}"
    SCOPE="repo"
else
    echo -e "${RED}❌ Invalid choice${NC}"
    exit 1
fi

# Number of runners
echo ""
read -p "Number of runners to create [3]: " NUM_RUNNERS
NUM_RUNNERS=${NUM_RUNNERS:-3}

if ! [[ "$NUM_RUNNERS" =~ ^[0-9]+$ ]] || [ "$NUM_RUNNERS" -lt 1 ]; then
    echo -e "${RED}❌ Invalid number of runners${NC}"
    exit 1
fi

# Installation directory
INSTALL_DIR="${HOME}/actions-runners"
echo ""
read -p "Installation directory [${INSTALL_DIR}]: " CUSTOM_DIR
INSTALL_DIR=${CUSTOM_DIR:-$INSTALL_DIR}

# Confirm
echo ""
echo -e "${YELLOW}═══════════════════════════════════════════════${NC}"
echo "Setup Summary:"
echo "  Type: $([ "$SETUP_TYPE" = "1" ] && echo "Organization-level" || echo "Repository-level")"
echo "  Target: ${URL}"
echo "  Runners: ${NUM_RUNNERS}"
echo "  Directory: ${INSTALL_DIR}"
echo -e "${YELLOW}═══════════════════════════════════════════════${NC}"
echo ""
read -p "Proceed? [y/N]: " CONFIRM

if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo "Aborted"
    exit 0
fi

# Create installation directory
echo ""
echo -e "${BLUE}Creating installation directory...${NC}"
mkdir -p "${INSTALL_DIR}"
cd "${INSTALL_DIR}"

# Download runner package
RUNNER_VERSION="2.321.0"
RUNNER_PACKAGE="actions-runner-osx-arm64-${RUNNER_VERSION}.tar.gz"
RUNNER_URL="https://github.com/actions/runner/releases/download/v${RUNNER_VERSION}/${RUNNER_PACKAGE}"

if [ ! -f "${RUNNER_PACKAGE}" ]; then
    echo -e "${BLUE}Downloading runner package...${NC}"
    curl -o "${RUNNER_PACKAGE}" -L "${RUNNER_URL}"
    echo -e "${GREEN}✅ Downloaded${NC}"
else
    echo -e "${GREEN}✅ Runner package already downloaded${NC}"
fi

# Generate registration token
echo ""
echo -e "${BLUE}Generating registration token...${NC}"

if [ "$SETUP_TYPE" = "1" ]; then
    # Organization token
    TOKEN=$(gh api -X POST "orgs/${ORG_NAME}/actions/runners/registration-token" --jq .token 2>&1)
else
    # Repository token
    TOKEN=$(gh api -X POST "repos/${OWNER}/${REPO}/actions/runners/registration-token" --jq .token 2>&1)
fi

if [ -z "$TOKEN" ] || [[ "$TOKEN" == *"HTTP"* ]]; then
    echo -e "${RED}❌ Failed to generate token${NC}"
    echo "Error: $TOKEN"
    echo ""
    echo "Common issues:"
    echo "  • You need admin access to the repository/organization"
    echo "  • Your gh CLI token needs 'admin:org' or 'repo' scope"
    echo ""
    echo "Try running: gh auth refresh -h github.com -s admin:org"
    exit 1
fi

echo -e "${GREEN}✅ Token generated${NC}"

# Set up runners
SUCCESS_COUNT=0
for i in $(seq 1 ${NUM_RUNNERS}); do
    RUNNER_NAME="runner-${i}"
    RUNNER_DIR="${INSTALL_DIR}/${RUNNER_NAME}"
    
    echo ""
    echo -e "${YELLOW}═══════════════════════════════════════════════${NC}"
    echo -e "${BLUE}Setting up ${RUNNER_NAME}${NC}"
    echo -e "${YELLOW}═══════════════════════════════════════════════${NC}"
    
    # Create directory
    mkdir -p "${RUNNER_DIR}"
    cd "${RUNNER_DIR}"
    
    # Extract
    echo "  Extracting..."
    tar xzf "../${RUNNER_PACKAGE}"
    
    # Configure
    echo "  Configuring..."
    ./config.sh \
        --url "${URL}" \
        --token "${TOKEN}" \
        --name "${RUNNER_NAME}" \
        --labels self-hosted,macOS,ARM64 \
        --unattended \
        2>&1 | grep -E "✓|✗|Connected|successfully" || true
    
    if [ $? -eq 0 ]; then
        # Install as service
        echo "  Installing as service..."
        ./svc.sh install 2>&1 | grep -v "warning" || true
        
        # Start service
        echo "  Starting..."
        ./svc.sh start
        
        echo -e "${GREEN}  ✅ ${RUNNER_NAME} is running!${NC}"
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    else
        echo -e "${RED}  ❌ Failed to configure ${RUNNER_NAME}${NC}"
    fi
    
    cd "${INSTALL_DIR}"
done

# Create monitoring script
echo ""
echo -e "${BLUE}Creating monitoring script...${NC}"
cat > "${INSTALL_DIR}/monitor.sh" << 'MONITOR'
#!/bin/bash
echo "═══════════════════════════════════════════════"
echo "  GitHub Actions Runners Status"
echo "═══════════════════════════════════════════════"
echo ""

RUNNER_DIRS=$(find . -maxdepth 1 -name "runner-*" -type d | sort)

for dir in $RUNNER_DIRS; do
    RUNNER_NAME=$(basename "$dir")
    echo "${RUNNER_NAME}:"
    
    if pgrep -f "Runner.Listener.*${RUNNER_NAME}" > /dev/null; then
        PID=$(pgrep -f "Runner.Listener.*${RUNNER_NAME}" | head -1)
        echo "  ✅ Running (PID: $PID)"
        
        # Resource usage
        ps -p $PID -o %cpu,%mem,etime 2>/dev/null | tail -1 | \
            awk '{print "     CPU: " $1 "% | Memory: " $2 "% | Uptime: " $3}'
        
        # Check for active job
        if pgrep -P $PID > /dev/null 2>&1; then
            echo "  🔨 Job active"
        else
            echo "  💤 Idle"
        fi
    else
        echo "  ❌ Not running"
        echo "     Start with: cd $dir && ./svc.sh start"
    fi
    echo ""
done

echo "═══════════════════════════════════════════════"
echo "System Resources:"
echo "  Load: $(uptime | awk -F'load average:' '{print $2}')"
echo "  Location: $(pwd)"
echo "═══════════════════════════════════════════════"
MONITOR

chmod +x "${INSTALL_DIR}/monitor.sh"
echo -e "${GREEN}✅ Created monitor.sh${NC}"

# Create management scripts
echo ""
echo -e "${BLUE}Creating management scripts...${NC}"

# Start all
cat > "${INSTALL_DIR}/start-all.sh" << 'STARTALL'
#!/bin/bash
for dir in runner-*/; do
    [ -d "$dir" ] || continue
    echo "Starting $(basename $dir)..."
    cd "$dir" && ./svc.sh start && cd ..
done
echo "Done!"
STARTALL
chmod +x "${INSTALL_DIR}/start-all.sh"

# Stop all
cat > "${INSTALL_DIR}/stop-all.sh" << 'STOPALL'
#!/bin/bash
for dir in runner-*/; do
    [ -d "$dir" ] || continue
    echo "Stopping $(basename $dir)..."
    cd "$dir" && ./svc.sh stop && cd ..
done
echo "Done!"
STOPALL
chmod +x "${INSTALL_DIR}/stop-all.sh"

# Status all
cat > "${INSTALL_DIR}/status-all.sh" << 'STATUSALL'
#!/bin/bash
for dir in runner-*/; do
    [ -d "$dir" ] || continue
    echo "$(basename $dir):"
    cd "$dir" && ./svc.sh status && cd ..
    echo ""
done
STATUSALL
chmod +x "${INSTALL_DIR}/status-all.sh"

echo -e "${GREEN}✅ Created management scripts${NC}"

# Summary
echo ""
echo -e "${GREEN}"
cat << 'EOF'
╔═══════════════════════════════════════════════════════════════════╗
║                                                                   ║
║   ✅  SETUP COMPLETE!                                             ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
EOF
echo -e "${NC}"

echo "Successfully configured: ${SUCCESS_COUNT}/${NUM_RUNNERS} runners"
echo ""
echo "Location: ${INSTALL_DIR}"
echo ""

if [ "$SETUP_TYPE" = "1" ]; then
    echo "Verify on GitHub:"
    echo "  https://github.com/organizations/${ORG_NAME}/settings/actions/runners"
else
    echo "Verify on GitHub:"
    echo "  https://github.com/${OWNER}/${REPO}/settings/actions/runners"
fi

echo ""
echo "Useful commands:"
echo "  Monitor runners:  ${INSTALL_DIR}/monitor.sh"
echo "  Start all:        ${INSTALL_DIR}/start-all.sh"
echo "  Stop all:         ${INSTALL_DIR}/stop-all.sh"
echo "  Check status:     ${INSTALL_DIR}/status-all.sh"
echo ""
echo "Individual runner control:"
echo "  cd ${INSTALL_DIR}/runner-1"
echo "  ./svc.sh status|start|stop"
echo ""

# Run monitor
echo "Current status:"
echo ""
"${INSTALL_DIR}/monitor.sh"

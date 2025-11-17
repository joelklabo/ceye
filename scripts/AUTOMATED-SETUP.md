# Automated Runner Setup with gh CLI

**Zero manual token copying!** This script uses the GitHub CLI to automatically generate tokens and set up runners.

## Prerequisites

```bash
# Install gh CLI (if not already installed)
brew install gh

# Authenticate
gh auth login

# Refresh with proper scopes
gh auth refresh -h github.com -s admin:org
```

## Usage

### Interactive Setup

```bash
cd /path/to/ceye
./Scripts/setup-runners-automated.sh
```

The script will:
1. ✅ Check for gh CLI and authentication
2. ✅ Ask if you want org-level or repo-level runners
3. ✅ **Automatically generate registration token**
4. ✅ Download runner package (if needed)
5. ✅ Configure N runners (you choose how many)
6. ✅ Install as services (auto-start on boot)
7. ✅ Create monitoring and management scripts

### What You'll Be Asked

```
1. Setup type?
   - Organization-level (works for ALL repos)
   - Repository-level (one repo only)

2. Organization name or Repo owner/name?

3. How many runners? [default: 3]

4. Installation directory? [default: ~/actions-runners]

5. Confirm and proceed
```

## Example Run

```bash
$ ./Scripts/setup-runners-automated.sh

╔═══════════════════════════════════════════════════════════════════╗
║   Automated GitHub Actions Runner Setup                          ║
║   Uses gh CLI to generate tokens automatically                   ║
╚═══════════════════════════════════════════════════════════════════╝

✅ gh CLI found
✅ Authenticated with GitHub
Logged in as: joelklabo

Setup Type:
  1) Organization-level (ONE token, works for ALL repos)
  2) Repository-level (one token per repo)

Choose [1 or 2]: 1

Available organizations:
 1. gim-home
 2. yammer-microsoft
 3. ms-copilot

Enter organization name: my-personal-org

Number of runners to create [3]: 3

Installation directory [/Users/honk/actions-runners]: 

═══════════════════════════════════════════════
Setup Summary:
  Type: Organization-level
  Target: https://github.com/my-personal-org
  Runners: 3
  Directory: /Users/honk/actions-runners
═══════════════════════════════════════════════

Proceed? [y/N]: y

Generating registration token...
✅ Token generated

═══════════════════════════════════════════════
Setting up runner-1
═══════════════════════════════════════════════
  Extracting...
  Configuring...
  ✓ Connected to GitHub
  ✓ Runner successfully added
  Installing as service...
  Starting...
  ✅ runner-1 is running!

[... runner-2 and runner-3 ...]

╔═══════════════════════════════════════════════════════════════════╗
║   ✅  SETUP COMPLETE!                                             ║
╚═══════════════════════════════════════════════════════════════════╝

Successfully configured: 3/3 runners

Verify on GitHub:
  https://github.com/organizations/my-personal-org/settings/actions/runners

Useful commands:
  Monitor runners:  /Users/honk/actions-runners/monitor.sh
  Start all:        /Users/honk/actions-runners/start-all.sh
  Stop all:         /Users/honk/actions-runners/stop-all.sh
```

## Generated Files

After running, you'll have:

```
~/actions-runners/
├── actions-runner-osx-arm64-2.321.0.tar.gz
├── runner-1/          # Configured and running
├── runner-2/          # Configured and running
├── runner-3/          # Configured and running
├── monitor.sh         # Check runner status
├── start-all.sh       # Start all runners
├── stop-all.sh        # Stop all runners
└── status-all.sh      # Check service status
```

## Management

### Monitor Runners
```bash
~/actions-runners/monitor.sh
```

Output:
```
═══════════════════════════════════════════════
  GitHub Actions Runners Status
═══════════════════════════════════════════════

runner-1:
  ✅ Running (PID: 12345)
     CPU: 0.1% | Memory: 0.2% | Uptime: 1:23:45
  💤 Idle

runner-2:
  ✅ Running (PID: 12346)
     CPU: 45.2% | Memory: 1.8% | Uptime: 1:23:40
  🔨 Job active

runner-3:
  ✅ Running (PID: 12347)
     CPU: 0.1% | Memory: 0.2% | Uptime: 1:23:35
  💤 Idle
```

### Start/Stop All Runners
```bash
~/actions-runners/start-all.sh    # Start all
~/actions-runners/stop-all.sh     # Stop all
~/actions-runners/status-all.sh   # Check status
```

### Individual Runner Control
```bash
cd ~/actions-runners/runner-1
./svc.sh status    # Check status
./svc.sh start     # Start
./svc.sh stop      # Stop
./svc.sh restart   # Restart
```

## Advantages Over Manual Setup

✅ **No manual token copying**
   - Script generates token automatically via API

✅ **No token expiration worries**
   - Token generated fresh each time

✅ **Repeatable**
   - Run script again to add more runners
   - Same process for org or repo level

✅ **Error handling**
   - Checks authentication
   - Validates permissions
   - Shows clear error messages

✅ **Management scripts included**
   - monitor.sh, start-all.sh, stop-all.sh
   - No need to remember commands

## Troubleshooting

### Authentication Error

```
❌ Not authenticated with GitHub
Run: gh auth login
```

**Solution:**
```bash
gh auth login
# Follow prompts to authenticate
```

### Permission Error

```
❌ Failed to generate token
Error: HTTP 403
You need admin access to the repository/organization
```

**Solution:**
```bash
# Refresh token with proper scopes
gh auth refresh -h github.com -s admin:org
```

For repository-level:
```bash
gh auth refresh -h github.com -s repo
```

### Runner Already Exists

If a runner with the same name already exists:
1. Remove it from GitHub (Settings → Actions → Runners)
2. Or choose a different runner name
3. Re-run the script

### Service Won't Start

```bash
cd ~/actions-runners/runner-1
./svc.sh status    # Check status
./svc.sh uninstall # Uninstall
./svc.sh install   # Reinstall
./svc.sh start     # Start
```

## Advanced Usage

### Different Number of Runners

```bash
# When prompted "Number of runners to create [3]:"
# Enter any number: 1, 2, 5, 10, etc.
```

### Custom Installation Directory

```bash
# When prompted "Installation directory:"
# Enter custom path: /opt/runners or ~/custom-location
```

### Add More Runners Later

Just run the script again! It will:
- Detect existing runners
- Add new runners alongside them
- Reuse the downloaded runner package

## Comparison: Manual vs Automated

| Task | Manual Setup | Automated Script |
|------|--------------|------------------|
| Get token | Copy from GitHub web UI | ✅ Auto-generated |
| Token expires? | After 1 hour | ✅ Fresh each time |
| Multiple runners | Repeat 3x | ✅ One command |
| Management scripts | Create manually | ✅ Auto-created |
| Error handling | DIY | ✅ Built-in |
| Time | ~15-20 min | ✅ ~3-5 min |

## Cleanup

To remove all runners:

```bash
cd ~/actions-runners

# Stop and uninstall services
for dir in runner-*/; do
  cd "$dir"
  ./svc.sh stop
  ./svc.sh uninstall
  cd ..
done

# Remove from GitHub (get removal token first)
# Then: ./config.sh remove --token REMOVAL_TOKEN

# Delete directory
cd ~
rm -rf ~/actions-runners
```

## Script Source

Location: `Scripts/setup-runners-automated.sh`

The script is self-contained and portable. Copy it to any Mac and run!

## Requirements

- macOS (tested on macOS 12+)
- gh CLI installed (`brew install gh`)
- GitHub account with appropriate permissions
- Bash 3.2+ (built-in on macOS)

## Security

The script:
- ✅ Uses secure GitHub API
- ✅ Tokens generated fresh each time
- ✅ No tokens stored in files
- ✅ Requires authentication via gh CLI
- ✅ Uses official GitHub Actions runner

## Support

For issues:
1. Check `gh auth status` - are you authenticated?
2. Check permissions - do you have admin access?
3. Check logs: `~/actions-runners/runner-1/_diag/Runner_*.log`
4. Run monitor.sh to check runner status

---

**Bottom line:** Never copy tokens manually again! 🎉

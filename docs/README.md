# ceye Documentation

## Quick Start

**Current Plan**: See [plan.md](plan.md) for the active development roadmap

## Directory Structure

```
docs/
├── README.md          # This file
├── plan.md            # Active development plan and roadmap
├── plans/             # Detailed implementation plans (historical)
└── references/        # Reference documentation and guides
```

## Main Documents

### [plan.md](plan.md) - Active Development Plan
The master plan tracking all current and future work:
- ✅ Completed features (Phases 1-5)
- 🚧 Current sprint (Options 2-5)
- 📋 Future roadmap
- 📊 Success metrics
- 🔄 Development process

**This is the single source of truth for what's being worked on.**

## Plans Directory

Historical implementation plans that led to completed features:

- **[ci-status-dashboard.md](plans/ci-status-dashboard.md)** - Original TUI dashboard plan (Phase 1)
- **[ui-enhancements.md](plans/ui-enhancements.md)** - 5 panel enhancement plan
- **[web-ui.md](plans/web-ui.md)** - Web UI implementation plan (Phase 2)

These documents are kept for historical reference and architecture decisions.

## References Directory

Guides, reference documentation, and completion reports:

### Guides
- **[readme.md](references/readme.md)** - Main project README
- **[testing-guide.md](references/testing-guide.md)** - Testing strategy and howto
- **[agents.md](references/agents.md)** - Provider (agent) architecture

### Completion Reports
- **[provider-interface-complete.md](references/provider-interface-complete.md)** - SafeProvider implementation summary
- **[provider-interface-hardening.md](references/provider-interface-hardening.md)** - Detailed hardening plan
- **[integration-tests-complete.md](references/integration-tests-complete.md)** - Integration test suite summary
- **[provider-contract-tests-complete.md](references/provider-contract-tests-complete.md)** - Contract test suite summary

### Demos
- **[ui-demo.txt](references/ui-demo.txt)** - TUI demo output

## How to Use This Documentation

### For Contributors

1. **Starting new work?** Check [plan.md](plan.md) for what's prioritized
2. **Need implementation details?** See the relevant section in [plan.md](plan.md)
3. **Working on a feature?** Update [plan.md](plan.md) with progress
4. **Feature complete?** Mark it done in [plan.md](plan.md)

### For Reviewers

1. Check [plan.md](plan.md) to see current sprint goals
2. Verify PR aligns with documented plan
3. Ensure tests match testing standards in plan
4. Verify documentation is updated

### For Users

1. See [references/readme.md](references/readme.md) for getting started
2. See [references/testing-guide.md](references/testing-guide.md) for running tests
3. See [plan.md](plan.md) for what's coming next

## File Naming Convention

All documentation files use **lowercase with hyphens**:
- ✅ `provider-interface.md`
- ✅ `testing-guide.md`
- ❌ `ProviderInterface.md`
- ❌ `Testing_Guide.md`

## Documentation Standards

### Plan Updates
- Keep [plan.md](plan.md) current
- Mark items complete when done
- Move detailed plans to `plans/` when archived
- Add completion reports to `references/`

### New Documents
- Reference docs go in `references/`
- Implementation plans go in `plans/` (if not part of main plan)
- Use lowercase-with-hyphens naming
- Add entry to this README

### Content Guidelines
- Use clear, concise language
- Include code examples where helpful
- Add test coverage info
- Link to related docs
- Keep diagrams simple

## Quick Links

### Getting Started
- [Main README](references/readme.md)
- [Testing Guide](references/testing-guide.md)

### Development
- [Current Plan](plan.md)
- [Provider Architecture](references/agents.md)

### Testing
- [Testing Guide](references/testing-guide.md)
- [Integration Tests](references/integration-tests-complete.md)
- [Contract Tests](references/provider-contract-tests-complete.md)

### Architecture
- [Provider Interface](references/provider-interface-complete.md)
- [Provider Hardening](references/provider-interface-hardening.md)

## Version History

### v1.0 (2025-11-16)
- Initial consolidated documentation structure
- Created master plan.md
- Organized into plans/ and references/
- Standardized naming conventions

## Contributing to Documentation

1. **Update plan.md** when starting/completing work
2. **Add new docs** to appropriate directory
3. **Update this README** when adding new docs
4. **Follow naming conventions**
5. **Keep it current** - docs should match reality

## Questions?

Check the [main README](references/readme.md) or open an issue.

# Documentation Inventory

**Last Updated**: 2025-11-16  
**Status**: Consolidated and current

## Structure

```
docs/
├── plan.md                          # Master development plan (current priorities)
├── README.md                        # Documentation index
└── references/                      # Reference documents (6 files)
    ├── agents.md                    # Agent context (symlinked from /AGENTS.md)
    ├── doc-inventory.md             # This file
    ├── readme.md                    # Project README
    ├── testing-guide.md             # Testing standards, contract tests, safety tests
    ├── webhook-guide.md             # Webhook implementation (consolidated from 6 files)
    └── web-ui-architecture.md       # Web UI design decisions
```

## Root Files

- `AGENTS.md` - Symlink to `docs/references/agents.md` (for GitHub Copilot/Claude)
- No other markdown files in root

## Recent Consolidation (2025-11-16)

**Deleted** (completed work, redundant, or superseded):
- docs-summary.md (reorganization complete)
- documentation-reorganization.md (reorganization complete)
- provider-interface-hardening.md (old plan, completed)
- provider-interface-complete.md (merged into testing-guide.md)
- provider-contract-tests-complete.md (merged into testing-guide.md)
- integration-tests-complete.md (merged into testing-guide.md)
- azure-provider-phase1-complete.md (key info in plan.md)
- next-session-start.md (outdated session notes)
- plans/ci-status-dashboard.md (superseded by plan.md)
- plans/ui-enhancements.md (completed features)

**Consolidated**:
- 6 webhook files → webhook-guide.md (543 lines → practical guide)
  - webhook-research.md (843 lines)
  - webhook-research-summary.md (146 lines)
  - webhook-localhost-analysis.md (372 lines)
  - webhook-setup-status.md (308 lines)
  - webhook-test-guide.md (334 lines)
  - webhook-test-results.md (231 lines)

**Moved**:
- plans/web-ui.md → references/web-ui-architecture.md

**Result**: 24 files → 6 files (75% reduction)

## Naming Convention

All files follow **lowercase-with-hyphens** (kebab-case) naming:
- ✅ `webhook-test-guide.md`
- ✅ `next-session-start.md`
- ❌ `WEBHOOK_TEST_GUIDE.md` (old style)

## Key Documents

| Document | Purpose |
|----------|---------|
| `plan.md` | Current sprint, priorities, roadmap |
| `agents.md` | Full project context for AI agents |
| `testing-guide.md` | Testing standards, contract tests, completion reports |
| `webhook-guide.md` | Webhook implementation (setup, testing, integration) |
| `web-ui-architecture.md` | Web UI design and architecture decisions |
| `readme.md` | Project overview and getting started |

## Organization Principles

1. **Naming**: All files use lowercase-with-hyphens (kebab-case)
2. **Location**: All markdown in `docs/` (except AGENTS.md symlink)
3. **Lifecycle**: Delete implementation plans after completion
4. **Consolidation**: Merge related docs (6 webhook files → 1 guide)
5. **Focus**: Keep only active/useful documentation
6. **Referencing**: Use relative paths in links

## Adding New Documentation

When adding new docs:
1. Use lowercase-with-hyphens naming
2. Place in `docs/references/` (reference material)
3. Update `docs/README.md` if it's a key document
4. Delete old implementation plans when feature complete
5. Consolidate if you create multiple related docs

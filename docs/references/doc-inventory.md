# Documentation Inventory

**Last Updated**: 2025-11-16  
**Status**: Complete

## Structure

```
docs/
├── plan.md                          # Master development plan (current priorities)
├── README.md                        # Documentation index
├── plans/                           # Implementation plans
│   ├── ci-status-dashboard.md
│   ├── ui-enhancements.md
│   └── web-ui.md
└── references/                      # Reference documents
    ├── agents.md                    # Agent context (symlinked from root)
    ├── azure-provider-phase1-complete.md
    ├── docs-summary.md
    ├── documentation-reorganization.md
    ├── integration-tests-complete.md
    ├── next-session-start.md
    ├── provider-contract-tests-complete.md
    ├── provider-interface-complete.md
    ├── provider-interface-hardening.md
    ├── readme.md
    ├── testing-guide.md
    ├── webhook-localhost-analysis.md
    ├── webhook-research-summary.md
    ├── webhook-research.md
    ├── webhook-setup-status.md
    ├── webhook-test-guide.md
    └── webhook-test-results.md
```

## Root Files

- `AGENTS.md` - Symlink to `docs/references/agents.md` (for GitHub Copilot/Claude)
- No other markdown files in root

## Naming Convention

All files follow **lowercase-with-hyphens** (kebab-case) naming:
- ✅ `webhook-test-guide.md`
- ✅ `next-session-start.md`
- ❌ `WEBHOOK_TEST_GUIDE.md` (old style)

## Key Documents

| Document | Location | Purpose |
|----------|----------|---------|
| Master Plan | `docs/plan.md` | Current sprint, priorities, roadmap |
| Agent Context | `docs/references/agents.md` | Full project context for AI agents |
| Testing Guide | `docs/references/testing-guide.md` | Testing standards and practices |
| Next Session | `docs/references/next-session-start.md` | Quick start for next work session |

## Document Types

### Plans (`docs/plans/`)
Historical implementation plans for completed features:
- Web UI implementation
- CI status dashboard design
- UI enhancements

### References (`docs/references/`)
Reference documentation, guides, and completion summaries:
- Webhook implementation docs
- Provider interface specifications
- Testing completion reports
- Setup and configuration guides

## Updates

When adding new documentation:
1. Use lowercase-with-hyphens naming
2. Place in `docs/plans/` if it's an implementation plan
3. Place in `docs/references/` if it's reference/guide material
4. Update `docs/README.md` if it's a key document
5. No markdown files in project root (except AGENTS.md symlink)

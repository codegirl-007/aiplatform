---
description: Work a Linear ticket (plan + worktrees + phased subtasks)
---

Work on Linear ticket $1.

CRITICAL:
- Plan file path: `.opencode/plans/$1.md` (source of truth for resuming)
- Before doing any implementation, run the preflight and report findings.

## Preflight: Resume Context + Worktrees

1) Check for existing plan file and load it if present:
!`test -f .opencode/plans/$1.md && echo "FOUND PLAN: .opencode/plans/$1.md" && sed -n '1,200p' .opencode/plans/$1.md || echo "NO PLAN FILE: .opencode/plans/$1.md"`

2) Check existing worktrees:
!`git worktree list`

3) If a plan exists, audit each worktree referenced by the plan:
- Does it exist?
- Any uncommitted/staged changes?
- On expected branch?
Run (for each worktree path found in the plan table):
- `git -C {worktree} status -sb`
- `git -C {worktree} rev-parse --abbrev-ref HEAD`
- (optional) `git -C {worktree} diff --stat` and `git -C {worktree} diff --staged --stat`

4) Report to the user BEFORE proceeding:
- Which expected worktrees exist/missing
- Which have dirty state (staged/unstaged)
- Branch mismatches
- What the plan says “Next Steps” are (if present)

5) Validate plan completeness/currentness (if plan exists):
Required sections:
- Summary
- Subtasks (with phases + statuses + worktree paths + dependencies)
- Dependency Graph
- Key Files
- Next Steps
Freshness checks:
- Linear parent/subtask states match plan
- Worktree reality matches plan (exists, branch, dirty/clean)
If anything is missing/outdated: UPDATE THE PLAN FIRST, then resume from “Next Steps”.

If no plan exists, proceed with Step 1.

## Step 1: Fetch Task & Existing Subtasks (Linear)

- Fetch parent issue: use Linear MCP `get_issue` for $1
- List existing subtasks/children: `list_issues` with `parentId`
- Present a concise summary:
  - Goal
  - Acceptance criteria
  - Existing subtasks (if any)

## Step 2: Research & Propose (Phased)

Explore the repo to determine:
- Key files/modules involved
- Existing patterns to follow
- Testing approach + constraints
- Dependencies between components

Ask clarifying questions ONLY if materially blocking.

Propose additional subtasks if needed. Rules:
- Atomic, independently mergeable units
- Each should be a sensibly small PR (aim <400 LOC changed)
- Each should include:
  - Goal
  - Acceptance Criteria
  - Technical Notes (patterns, files, constraints)
- CRITICAL: Define PHASES before proceeding (Phase 1 independent, Phase 2 depends on Phase 1, etc.)
Present phase breakdown clearly.

Also propose an updated parent ticket description (summary + phase plan + research notes + key files).
After user approval: update parent via `update_issue`.

Ask the user to approve:
- New subtasks
- Subtask descriptions
- Phase breakdown
BEFORE creating anything in Linear.

## Step 3: Create Subtasks in Linear (If Approved)

For each approved subtask:
- `create_issue` with `parentId` = parent issue id
- Match parent team/project
- Assign to `me`
- Set state to “Todo” (or team-equivalent)
- Prefix title with phase: `[Phase 1] ...`
- Include Dependencies section when needed

Create in phase order.

## Step 4: Create/Update Plan File (Source of Truth)

Create or update `.opencode/plans/$1.md` after:
- research + proposed phases
- Linear updates (parent + subtasks)
- any worktree creation
- any agent start/finish
- PR merged status changes
- discussions/decisions/blockers

Plan template:

# $1: {title}

**Status:** In Progress
**Created:** {date}
**Project:** {project}
**Team:** {team}

## Summary
{brief overall description}

## Subtasks

| Ticket | Title | Phase | Status | Worktree | Depends On |
|--------|-------|-------|--------|----------|------------|
| ...    | ...   | 1     | ⏳/🔄/✅/❌ | `../platform-...` | ... |

Status emojis: ⏳ Queued | 🔄 Agent Running | ✅ Done | ❌ Failed

## Dependency Graph
{ASCII dependencies}

## Key Files
| Component | Location |
|----------|----------|
| ...      | ...      |

## Research Notes
{findings}

## Next Steps
{explicit next actions}

## Step 5: Worktrees (Phase-Gated)

Worktree naming (default; adjust if repo uses a different convention):
- worktree: `../platform-$TICKET`
- branch: `feature/$TICKET`

Phase gate:
- Do NOT create worktrees / start work for later phases until earlier phase PRs are MERGED to main.

Create examples:
- Independent: `git worktree add ../platform-$TICKET -b feature/$TICKET main`
- Dependent: `git worktree add ../platform-$TICKET -b feature/$TICKET feature/$DEPENDENCY_TICKET`

If worktree already exists: inspect and continue; do not delete silently.

## Step 6: Execution Mode (Pick Based on Scope)

Option A: Single-agent (default)
- Implement tasks sequentially in the appropriate worktree(s)

Option B: Background agents (only when explicitly chosen)
- Spawn per-subtask agents, track in plan, phase-gated
- Update Linear status to “In Progress” when an agent starts
- Update plan statuses as agents complete/fail

## Step 7: Report

Tell the user:
- Current phase
- Worktrees + branches
- What’s running (if using agents)
- How to check status
- What remains

## Project Standards Reminders

- Follow invariant-driven development (`docs/ALGO.md`, `docs/TRADING.md`)
- Minimum 2 assertions per function (avg)
- Hard limits: 70 lines/function, 100 columns
- Use snake_case for internal identifiers; `go fmt`
- Test valid AND invalid cases
- Never commit secrets or `.env`

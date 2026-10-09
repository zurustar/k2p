---
description: Workflow to follow before starting any implementation work
---

# Development Guidelines

## **MANDATORY: Documentation First, Tests First**

When implementing any new feature, bug fix, or significant change:

### Order of Operations

**YOU MUST FOLLOW THIS ORDER:**

0. **Check docs/tasks.md and Define Git Branch Strategy**
   - Read `docs/tasks.md` first. If work is in progress, resume it from "Next Step".
   - Read `.agent/workflows/git-flow.md`.
   - **CRITICAL**: In your **Implementation Plan**, you MUST include a specific step to create or switch to the correct branch (e.g., `git checkout -b feature/name`).
   - Run `git status` to verify your current environment.

1. **Update docs/requirements.md FIRST** (if new requirements/features are requested)
   - Add new User Stories or Acceptance Criteria
   - Ensure the "What" is defined before the "How"
   - Verify with user if requirements are ambiguous

2. **Update docs/design.md SECOND** (if design changes are needed)
   - Add or modify relevant sections (Components, Data Models, Workflow, etc.)
   - Update interfaces if they change
   - Document new correctness properties if applicable
   - Keep the design document in sync with implementation

3. **Update docs/tasks.md BEFORE implementing**
   - Add the tasks for this work under an "In Progress" section
   - Mark tasks as `[ ]` (incomplete) initially
   - Update the "Current Status" section (branch, state, next step)

4. **Write tests FIRST**
   - Write tests that express the acceptance criteria and correctness properties
   - Run them and **confirm they fail** before writing any implementation
   - Do not weaken or skip a test to make it pass

5. **Implement the code**
   - Write the minimum code needed to make the tests pass
   - Follow the plan from the documentation

6. **Mark tasks complete in docs/tasks.md**
   - Update tasks to `[x]` when completed
   - Keep task list synchronized with actual progress

7. **Verify**
   - Run `make build` and all tests
   - For anything that needs the real Kindle app, follow `.agent/workflows/manual-verification.md`
   - Ensure docs/requirements.md and docs/design.md match what was actually built

8. **Clean up docs/tasks.md when merging**
   - Remove the completed tasks for this work (history lives in git and PRs)

## Why This Matters

- **Design.md** is the source of truth for architecture and interfaces
- **Tests written first** turn requirements into executable checks, so an AI agent cannot claim "done" without evidence
- **Tasks.md** lets anyone (human or AI) resume interrupted work
- Keeping docs in sync prevents drift between design and implementation

## Consequences of Not Following

- ❌ Design document becomes outdated and useless
- ❌ Untested or unverifiable changes slip in
- ❌ Interrupted work cannot be resumed
- ❌ Future developers (including AI agents) will be confused

## Task Management Rules (docs/tasks.md)

`docs/tasks.md` is a **handoff file**, not a history log.

- Keep only in-progress and upcoming tasks
- Update it immediately after completing each task or sub-task
- Keep "Current Status" accurate: branch, state, and the concrete next step

### Interrupting Work
Before stopping in the middle of work (end of session, context limit, waiting for the user):
1. Update "Current Status" with the exact next step and anything the next person must know
2. Commit `docs/tasks.md` together with the work in progress on the feature branch

### Resuming Work
1. Read `docs/tasks.md`
2. Switch to the branch in "Current Status"
3. Continue from "Next Step"

## Implementation Notes

### Testing
- Write tests before implementation (see Order of Operations)
- Ensure all tests pass before moving to the next phase
- Write property-based tests with minimum 100 iterations

### Code Quality
- Follow Go best practices and idioms
- Add comprehensive error handling
- Include detailed comments for complex logic
- Keep functions focused and testable

## Quick Reference

**For ANY code change (including bug fixes):**
1. ✓ Read docs/tasks.md, create/switch branch
2. ✓ Update docs/requirements.md (if new requirements)
3. ✓ Update docs/design.md (if design changes)
4. ✓ Add tasks to docs/tasks.md
5. ✓ Write tests and confirm they fail
6. ✓ Implement until tests pass
7. ✓ Mark tasks as [x] in docs/tasks.md
8. ✓ Run `make build` and tests; manual verification if needed
9. ✓ Verify docs (requirements/design) still match implementation
10. ✓ Remove completed tasks from docs/tasks.md when merging

**NEVER skip documentation updates or tests, even for "small" changes.**

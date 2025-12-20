---
description: Workflow to follow before starting any implementation work
---

# Development Guidelines

## **MANDATORY: Update Documentation Before Implementation**

When implementing any new feature or making significant changes:

### Order of Operations

**YOU MUST FOLLOW THIS ORDER:**

0. **Check Git Workflow**
   - Read `.agent/workflows/git-flow.md` to ensure you are on the correct branch.
   - Run `git status` to verify your environment.

1. **Update docs/design.md FIRST** (if design changes are needed)
   - Add or modify relevant sections (Components, Data Models, Workflow, etc.)
   - Update interfaces if they change
   - Document new correctness properties if applicable
   - Keep the design document in sync with implementation

2. **Update docs/tasks.md BEFORE implementing**
   - Add new tasks/phases for the feature or bug fix
   - Place them in the appropriate section (Bug Fixes, Additional Features, etc.)
   - Mark tasks as `[ ]` (incomplete) initially
   - Update "Current Status" section if needed

3. **Implement the code**
   - Now you can make the actual code changes
   - Follow the plan from the documentation

4. **Mark tasks complete in docs/tasks.md**
   - Update tasks to `[x]` when completed
   - Keep task list synchronized with actual progress

5. **Verify design.md still matches implementation**
   - Ensure documentation reflects what was actually built

## Why This Matters

- **Design.md** is the source of truth for architecture and interfaces
- **Tasks.md** tracks implementation progress and serves as a checklist
- Updating documentation first ensures thoughtful design before coding
- Keeping docs in sync prevents drift between design and implementation

## Consequences of Not Following

- ❌ Design document becomes outdated and useless
- ❌ Task tracking becomes meaningless
- ❌ Future developers (including AI agents) will be confused
- ❌ Code reviews become harder without up-to-date design docs

## Task Management Rules

### Task Completion Updates
When completing tasks during implementation:
1. **Update `docs/tasks.md`** immediately after completing each task or sub-task
2. Mark completed items with `[x]` instead of `[ ]`
3. Keep the task list synchronized with actual progress
4. Update the "Current Status" section at the bottom of `docs/tasks.md`

### Workflow
1. Complete a task or sub-task
2. Update `docs/tasks.md` to mark it as complete
3. Continue to next task
4. Repeat

This ensures the task tracking document always reflects the current state of the project.

## Implementation Notes

### Testing
- Run tests after implementing each component
- Ensure all tests pass before moving to the next phase
- Write property-based tests with minimum 100 iterations

### Code Quality
- Follow Go best practices and idioms
- Add comprehensive error handling
- Include detailed comments for complex logic
- Keep functions focused and testable

## Quick Reference

**For ANY code change (including bug fixes):**
1. ✓ Update docs/tasks.md FIRST (add task items)
2. ✓ Update docs/design.md if needed (design changes)
3. ✓ Implement the code
4. ✓ Mark tasks as [x] in docs/tasks.md
5. ✓ Run `make build` and tests
6. ✓ Verify docs still match implementation

**NEVER skip documentation updates, even for "small" changes.**

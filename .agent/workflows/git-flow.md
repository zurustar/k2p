---
description: Workflow for managing git branches for tasks
---

# Git Branch Workflow

This workflow defines how to manage git branches for development tasks.

## Rules

1.  **One Task, One Branch**:
    - Always create a new branch for a unit of work defined in `docs/tasks.md`.
    - Do not commit directly to `main` (or `master`) unless it's a trivial documentation fix.

2.  **Branch Naming Convention**:
    - `feature/name` - New features
    - `fix/name` - Bug fixes
    - `refactor/name` - Refactoring
    - `docs/name` - Documentation only
    - Example: `feature/remove-config`, `refactor/pkg-to-internal`

3.  **Workflow Steps**:
    1.  **Start**: Before writing code, create and switch to a new branch.
        ```bash
        git checkout -b <branch-name>
        ```
    2.  **Work**: Implement changes, running tests as you go.
    3.  **Commit**: Commit changes with clear messages.
    4.  **Finish**:
        - Merge to main (simulating a PR merge).
        - Delete the feature branch.
        ```bash
        git checkout main
        git merge <branch-name>
        git branch -d <branch-name>
        ```

## When to apply
- Apply this workflow automatically when starting a new top-level item from `docs/tasks.md`.

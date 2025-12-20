---
description: Workflow for handling manual verification tasks requiring real device interaction
---

# Manual Verification Workflow

This workflow defines the protocol for tasks that require real device execution (e.g., controlling the Kindle app), which the agent cannot perform directly.

## Rules

1.  **Identify Manual Tasks**:
    - Tasks involving "real device", "Kindle app", "screenshots on screen", or "AppleScript automation" typically require user intervention.
    - Check if the task in `docs/tasks.md` implies physical device interaction.

2.  **Request Verification**:
    - **Do NOT** try to automate this or skip it.
    - Ask the user explicitly to perform the verification.
    - Provide clear instructions (e.g., "Please run `k2p` with a book open").
    - Reference specific steps from `docs/MANUAL_TESTING.md` if applicable.

3.  **Wait for Report**:
    - Wait for the user to report the results (e.g., "It worked," "The page didn't turn").
    - **Do NOT** assume success until the user confirms.

4.  **Completion**:
    - If the user reports success, you may mark the task as complete `[x]` in `docs/tasks.md`.
    - If the user reports failure, treat it as a bug and switch to fixing it.

## Example Interaction

**Agent**:
> "The implementation for page turning is complete. This requires manual verification on your device. Please run `make build && ./build/k2p` with a book open and confirm if pages turn automatically."

**User**:
> "I tried it. It turned pages correctly."

**Agent**:
> "Great. I will mark the automation task as complete."

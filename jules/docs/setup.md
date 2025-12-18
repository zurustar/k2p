# Setup and Permissions

## Overview
This tool automates capturing screenshots and turning pages on the Kindle for Mac application. To perform these actions, the tool interacts with the macOS operating system in ways that require explicit user permission.

## Required Permissions

To use this tool, you must grant the following permissions to the terminal application (e.g., Terminal, iTerm2, VS Code) where you are running the command:

### 1. Screen Recording
**Reason:** The tool uses the `screencapture` command to take screenshots of the book pages.
**How to Enable:**
1. Open **System Settings** (or System Preferences).
2. Go to **Privacy & Security**.
3. Select **Screen & System Audio Recording**.
4. Find your terminal application (e.g., iTerm, Terminal) in the list.
5. Toggle the switch to **ON**.

### 2. Accessibility (System Events)
**Reason:** The tool uses AppleScript (`osascript`) to simulate keyboard presses (Left/Right arrow keys) for page turning. This requires controlling "System Events".
**How to Enable:**
1. Open **System Settings** (or System Preferences).
2. Go to **Privacy & Security**.
3. Select **Accessibility**.
4. Find your terminal application (e.g., iTerm, Terminal) in the list.
5. Toggle the switch to **ON**.

> **Note:** If you run the tool without these permissions, it may fail silently (screenshots might be empty/black) or crash with an error related to `osascript`.

## Troubleshooting

- **"System Events got an error: Application isn't running"**: This usually means Accessibility permission is missing.
- **Black/Empty Screenshots**: This usually means Screen Recording permission is missing.
- **Permission changes not taking effect**: After changing permissions, you usually need to **restart your terminal application** for them to take effect.

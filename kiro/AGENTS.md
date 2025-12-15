# Development Guidelines

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

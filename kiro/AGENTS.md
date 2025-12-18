# Development Guidelines

## Documentation Update Rules

### **MANDATORY: Update Documentation Before Implementation**

When implementing any new feature or making significant changes:

1. **ALWAYS update `docs/design.md` FIRST**
   - Add or modify relevant sections (Components, Data Models, Workflow, etc.)
   - Update interfaces if they change
   - Document new correctness properties if applicable
   - Keep the design document in sync with implementation

2. **ALWAYS update `docs/tasks.md` DURING implementation**
   - Add new tasks/phases for the feature
   - Mark tasks as `[x]` when completed
   - Update "Current Status" section
   - Keep task list synchronized with actual progress

3. **Order of Operations**
   ```
   1. Update docs/design.md (design the change)
   2. Update docs/tasks.md (plan the tasks)
   3. Implement the code
   4. Mark tasks complete in docs/tasks.md
   5. Verify design.md still matches implementation
   ```

### Why This Matters

- **Design.md** is the source of truth for architecture and interfaces
- **Tasks.md** tracks implementation progress and serves as a checklist
- Updating documentation first ensures thoughtful design before coding
- Keeping docs in sync prevents drift between design and implementation

### Consequences of Not Following

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

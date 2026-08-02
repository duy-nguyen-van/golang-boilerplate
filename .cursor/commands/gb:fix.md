Fix the described bug or issue systematically.

**Rule: Find root cause FIRST. No random fixes.**

## Workflow

### Step 1: Understand
- Read error, stack trace, or reproduction steps
- Check RBAC/Keycloak/OTel docs if auth or observability-related

### Step 2: Root Cause Analysis
- Trace handler → service → repo/cache/integration
- `git log --oneline -10` for recent changes
- 2–3 hypotheses; verify with code/logs/tests

### Step 3: Implement Fix
- Minimal change at correct layer
- Use `AppError` patterns from surrounding code

### Step 4: Verify
- Add regression test
- `make lint` and `make tests` (or scoped package test)

### Step 5: Self-Review
- Right layer? Concurrent access safe? Errors wrapped?

## Red Flags
- "Just try changing this" — stop, find root cause
- "Tests pass" — confirm the actual bug is fixed

Suggest conventional commit message when done. Do not commit unless asked.

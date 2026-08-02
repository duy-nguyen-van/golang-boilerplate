Review recent code changes systematically.

## Steps
1. Identify changed files (`git diff` or `git diff --cached`)
2. Per file, check priority list in `.cursor/rules/code-review.mdc`:
   - Correctness, security (JWT/RBAC/CSRF), errors, concurrency, performance, tests, OpenAPI
3. Scout edge cases: JWT expiry, missing roles, DTO leaking sensitive fields, cache races
4. Verify routes registered under `/api/v1` in `router.go`
5. Report Critical / High / Medium / Low with fix suggestions
6. Flag missing `make swagger-load` when HTTP surface changed

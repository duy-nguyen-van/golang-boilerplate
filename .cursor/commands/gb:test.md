Run the test suite and analyze results.

## Steps
1. `make lint` (or `go vet ./...` if lint unavailable)
2. Scoped tests if changes are localized:
   - `make test-handlers` / `make test-services` / `make test-repositories` / `make test-utils`
3. Full suite: `make tests` (coverage + race + count=1)
4. Optional coverage report: `make test-coverage` or `make test-coverage-html`
5. For each failure: root cause + suggested fix
6. Report: pass/fail/skip counts, coverage %, race findings, coverage gaps

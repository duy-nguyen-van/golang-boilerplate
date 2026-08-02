# Contributing

Thank you for your interest in contributing! Please follow these guidelines to help us keep things organized and productive.

## Getting Started

- Fork the repo and create your branch from `main`.
- Ensure you have Go 1.25+, Docker, and Make installed.
- Use `make container-up` to start dependencies, `make up` to run the app.
- Using Cursor? Read [AGENTS.md](AGENTS.md) for layer boundaries, make targets, and `.cursor/` rules/commands (`gb:*`).

## Development

- Run `make dep` before committing to ensure modules are tidy.
- Run `make tests` and `make lint` locally.
- Run `make swagger-load` when HTTP routes or swag annotations change.
- Update or add tests for any changes.
- Prefer conventional commits (`feat(scope):`, `fix(scope):`, …) — see `.cursor/rules/git-conventions.mdc` if present.

## Commit Messages

- Use clear and descriptive messages.
- Reference issues where relevant.

## Pull Requests

- Fill out the PR template.
- Include screenshots or logs for API changes when helpful.
- Ensure CI passes (lint, build, tests).

## Code Style

- Follow idiomatic Go practices.
- Prefer clarity and readable names.
- Avoid catching errors without handling.

## Security

- Do not include secrets in commits.
- Report vulnerabilities privately (see SECURITY.md).

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

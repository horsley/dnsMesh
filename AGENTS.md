# Repository Guidelines

## Project Structure & Module Organization
- **Backend**: `backend/cmd` hosts the Go entry point; core logic lives under `backend/internal` (`handlers`, `services`, `models`, `middleware`, `database`). Shared utilities belong in `backend/pkg`, and bundled frontend assets ship via `backend/public`.
- **Frontend**: Maintain Mithril views, components, services, and styles inside `frontend/src/{views,components,services,styles}`. Keep static assets in `frontend/public` to preserve Vite paths.
- **Scripts**: Reuse `docker-compose.yml`, `start.sh`, and `test.sh` for orchestration instead of adding new bootstrap scripts.

## Build, Test, and Development Commands
- **Provision DB**: `docker-compose up -d postgres` starts the required PostgreSQL instance.
- **Run backend**: `cd backend && go mod tidy && go run cmd/main.go` refreshes modules and serves the API on `localhost:8080`. Because the backend is a server process listening on port 8080 and it's a long-running process which will not exit, be careful not to run the backend server directly without nohup or any other process manager to avoid blocking the terminal or agent conversation context.
- **Run frontend**: `cd frontend && npm install && npm run dev` launches the Vite dev server on `localhost:3000`.
- **Backend tests**: `cd backend && go test ./...` executes Go unit tests; narrow packages when iterating.
- **Smoke tests**: `./test.sh` validates auth, provider, record, and audit endpoints against a running backend.

## Codex CLI Usage Notes
- **Background jobs**: Use `bash -lc "nohup <command> > /tmp/<name>.log 2>&1 & echo $!"` to launch long-running services (e.g. `go run cmd/main.go`) without blocking Codex. The `nohup` keeps the process alive after the shell exits and the trailing `&` detaches it so subsequent commands can run.
- **PID management**: Capture the printed PID or redirect it to a file so you can later stop the service with `kill <PID>` or `pkill -f <pattern>` once your testing is finished.
- **Log review**: Redirect output to a log file and inspect it with `tail -f /tmp/<name>.log` from a separate command when you need live feedback.
- **Cleanup**: Always terminate background services before finishing your task to avoid orphaned processes or port conflicts.

## Coding Style & Naming Conventions
- **Go**: Run `go fmt ./...` before commits. Restrict exported identifiers, keep persistence code in `database` packages, and follow handler→service→model layering.
- **Frontend**: Use camelCase for functions, PascalCase components, and kebab-case filenames. Scope styles to components or `src/styles`.

## Testing Guidelines
- **Placement**: Co-locate Go tests with implementations using `_test.go`. Frontend checks live beside their modules.
- **Execution**: After backend or auth changes, run `./test.sh` and targeted `go test` commands. Document manual verification if automation is incomplete.

## Commit & Pull Request Guidelines
- **Commits**: Prefer imperative titles (e.g., `Add DNS provider audit log`). Split backend/frontend work when possible and reference issues with `Closes #123`.
- **PRs**: Describe problem, solution, and tests. Surface env var updates, attach relevant screenshots or console output, and call out configuration impacts.

## Security & Configuration Tips
- **Secrets**: Copy `.env.example` to `.env` and supply provider secrets via environment variables only.
- **Rotation**: Rotate credentials in `docker-compose.yml` when modified and document any operational changes in the pull request.

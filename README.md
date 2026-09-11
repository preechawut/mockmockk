# mockapi

A mock API server built with Go, Fiber, and PostgreSQL.

## Quick Start

### Docker Compose
```sh
cp .env.example .env
make up
```

### Local Development
```sh
cp .env.example .env

# Start Postgres & run migrations
make db-up
make migrate-up

# Run application
set -a && . ./.env && set +a
make run
```

## Testing
```sh
make test          # Run unit tests with race detector
make test-coverage # Generate HTML coverage report
make lint          # Run golangci-lint
```

## API Endpoints

### Project & Mock Endpoints Management
- `POST /api/v1/projects` - Create a project
- `GET /api/v1/projects` - List all projects
- `POST /api/v1/projects/:projectId/endpoints` - Create a mock endpoint in project
- `GET /api/v1/projects/:projectId/endpoints` - List all mock endpoints in project
- `GET /api/v1/projects/:projectId/endpoints/:id` - Get a mock endpoint by ID
- `PUT /api/v1/projects/:projectId/endpoints/:id` - Update a mock endpoint
- `POST /api/v1/projects/:projectId/endpoints/:id/duplicate` - Duplicate a mock endpoint
- `DELETE /api/v1/projects/:projectId/endpoints/:id` - Delete a mock endpoint

### Global Mockmocks Management (Compatibility)
- `POST /api/v1/mockmocks` - Create a global mock endpoint
- `GET /api/v1/mockmocks` - List global mock endpoints
- `PUT /api/v1/mockmocks/:id` - Update a global mock endpoint
- `DELETE /api/v1/mockmocks/:id` - Delete a global mock endpoint

### Dynamic Mock Dispatcher
- `ALL /mock/:projectSlug/*` - Dispatches to project-scoped mock endpoint matching method and path, returning the configured status code and custom JSON payload.
- `ALL /mock/*` - Dispatches to global mock endpoints.

### Health Checks
- `GET /healthz` - Liveness check (`{"status":"ok"}`)
- `GET /readyz` - Readiness check with database ping & pool stats

## Local UI

Open `http://localhost:8080/` after `make up`. The UI is plain HTML/CSS/JavaScript embedded in the Go binary (no npm or frontend build step needed).

- **Project Management**: Create and select Projects from the sidebar.
- **Mock Endpoints**: For each project, easily create mock endpoints by selecting HTTP Method (GET, POST, PUT, DELETE, PATCH...), setting Path (e.g. `/users`, `/auth/login`), Status Code, and pasting any custom JSON response.
- **1-Click Copy URL**: Instantly copy full mock URL `http://localhost:8080/mock/:projectSlug/:path` ready to be called from your app or curl.
- **Inline Actions**: Edit, duplicate, and delete mock endpoints with live preview.

UI files: `internal/web/static/`; route/embedding: `internal/web/web.go`; wired in `internal/app/app.go`.

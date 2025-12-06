# Web Admin Panel Implementation Plan

## Overview

Add a web-based admin panel to the Betrayal Bot for operational control and visibility.

## Tech Stack

- **Echo v4** - Go web framework
- **templ** - Type-safe Go templates
- **HTMX 2.0** - Server-driven interactivity (no JavaScript bloat)
- **Tailwind CSS** - Styling (standalone CLI, no Node.js runtime)
- **Railway GraphQL API** - For triggering redeploys

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                     main.go                              │
├──────────────────────────────────────────────────────────┤
│  ┌─────────────────┐         ┌─────────────────────────┐ │
│  │   Discord Bot   │         │      Web Server         │ │
│  │   (Ken/DG)      │         │    (Echo :8080)         │ │
│  │                 │         │                         │ │
│  │  - Commands     │         │  GET  /                 │ │
│  │  - Events       │         │  GET  /health           │ │
│  │                 │         │  GET  /players          │ │
│  │                 │         │  GET  /players/:id      │ │
│  │                 │         │  POST /admin/redeploy   │ │
│  │                 │         │  GET  /admin/audit      │ │
│  └────────┬────────┘         └───────────┬─────────────┘ │
│           │                              │               │
│           │      ┌──────────────┐        │               │
│           └──────┤   Shared DB  ├────────┘               │
│                  │     Pool     │                        │
│                  └──────────────┘                        │
└──────────────────────────────────────────────────────────┘
```

## Environment Variables

| Variable Name | Purpose |
|---------------|---------|
| `WEB_PORT` | Port for web server (8080) |
| `ADMIN_PASSWORD` | Shared password for admin login |
| `RAILWAY_API_TOKEN` | API token from Railway dashboard |
| `RAILWAY_BETRAYAL_PROJECT_ID` | Railway project ID |
| `RAILWAY_BETRAYAL_SERVICE_ID` | Railway service ID |
| `RAILWAY_BETRAYAL_ENVIRONMENT_ID` | Railway environment ID |

## Project Structure

```
betrayal/
├── cmd/
│   └── betrayal-bot/
│       └── main.go                       # MODIFY - Add web server goroutine
├── internal/
│   └── web/                              # NEW DIRECTORY
│       ├── server.go                     # Echo setup, routes
│       ├── handlers/
│       │   ├── dashboard.go              # GET /
│       │   ├── health.go                 # GET /health
│       │   ├── players.go                # GET /players, /players/:id
│       │   ├── admin.go                  # POST /admin/redeploy
│       │   └── auth.go                   # Login/logout
│       ├── middleware/
│       │   └── auth.go                   # Auth middleware
│       ├── templates/
│       │   ├── layouts/
│       │   │   └── base.templ
│       │   ├── pages/
│       │   │   ├── dashboard.templ
│       │   │   ├── login.templ
│       │   │   ├── players.templ
│       │   │   ├── player_detail.templ
│       │   │   └── audit.templ
│       │   ├── partials/
│       │   │   ├── nav.templ
│       │   │   ├── player_row.templ
│       │   │   └── player_table.templ
│       │   └── components/
│       │       ├── button.templ
│       │       ├── badge.templ
│       │       └── card.templ
│       └── railway/
│           └── client.go                 # Railway GraphQL client
├── web/
│   └── static/
│       ├── css/
│       │   ├── input.css                 # Tailwind source
│       │   └── output.css                # Generated (committed)
│       └── js/
│           └── htmx.min.js               # Vendored HTMX 2.0
├── tailwind.config.js                    # NEW
├── .env.example                          # MODIFY
├── Makefile                              # MODIFY
└── Dockerfile                            # MODIFY
```

## Routes

| Method | Path | Handler | Auth | Description |
|--------|------|---------|------|-------------|
| GET | `/` | Dashboard | Yes | Main dashboard |
| GET | `/health` | Health | No | Health check for Railway |
| GET | `/login` | LoginPage | No | Login form |
| POST | `/login` | LoginSubmit | No | Process login |
| POST | `/logout` | Logout | Yes | Clear session |
| GET | `/players` | PlayerList | Yes | Full player list page |
| GET | `/players/table` | PlayerTable | Yes | HTMX partial for player table |
| GET | `/players/:id` | PlayerDetail | Yes | Single player detail |
| POST | `/admin/redeploy` | Redeploy | Yes | Trigger Railway redeploy |
| GET | `/admin/audit` | AuditLogs | Yes | View command audit logs |

## Implementation Phases

### Phase 1: Foundation (P0)
1. Add Echo and templ dependencies to go.mod
2. Create `internal/web/` package structure
3. Set up Echo server with basic middleware (logging, recovery, static files)
4. Create base layout template
5. Implement `/health` endpoint
6. Modify `main.go` to run web server alongside Discord bot
7. Create `web/static/` directory with Tailwind input.css
8. Create tailwind.config.js
9. Download and vendor HTMX 2.0
10. Update Makefile with templ and tailwind targets
11. Update Dockerfile to include templ generate and tailwind build

### Phase 2: Authentication (P0)
1. Implement simple session-based auth with shared password
2. Create login page template
3. Add auth middleware to protect routes
4. Create logout handler

### Phase 3: Dashboard & Players (P0)
1. Create dashboard page showing:
   - Current cycle (day/night + number)
   - Player count (alive/dead)
   - Quick stats
2. Create player list page with HTMX-powered search/filter
3. Create player detail page showing full inventory:
   - Role, alignment, status
   - Items, abilities, perks
   - Coins, luck
   - Notes
4. Create nav component with links

### Phase 4: Railway Integration (P1)
1. Create Railway GraphQL client
2. Implement redeploy button on admin page
3. Add deployment status indicator (optional)

### Phase 5: Audit Logs (P2)
1. Create audit log viewer page
2. Add pagination with HTMX
3. Add filtering by command/user

## Makefile Targets

```makefile
# Templ
templ-generate:
	templ generate

templ-watch:
	templ generate --watch

# Tailwind CSS
tailwind-build:
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css

tailwind-watch:
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --watch

tailwind-minify:
	tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify

# Combined build (for CI/deploy)
web-build: templ-generate tailwind-minify
```

## Dockerfile Changes

```dockerfile
# Install templ CLI
RUN go install github.com/a-h/templ/cmd/templ@latest

# Install Tailwind CSS standalone CLI
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 \
    && chmod +x tailwindcss-linux-x64 \
    && mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

# Generate templ files
RUN templ generate

# Build Tailwind CSS (minified for production)
RUN tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify
```

## Railway API

The Railway API is GraphQL-based. Key operations:

### Get Latest Deployment
```graphql
query deployments {
  deployments(
    first: 1
    input: {
      projectId: "PROJECT_ID"
      environmentId: "ENVIRONMENT_ID"
      serviceId: "SERVICE_ID"
    }
  ) {
    edges {
      node {
        id
        staticUrl
      }
    }
  }
}
```

### Restart Deployment
```graphql
mutation deploymentRestart {
  deploymentRestart(id: "DEPLOYMENT_ID")
}
```

## Security Considerations

1. **Session Management**: Secure cookies with HttpOnly, SameSite flags
2. **CSRF Protection**: Use Echo's CSRF middleware with HTMX
3. **Password**: Compare against ADMIN_PASSWORD env var
4. **Rate Limiting**: Limit /login and /admin/redeploy endpoints
5. **HTTPS**: Railway handles SSL termination

## Dependencies to Add

```
github.com/labstack/echo/v4
github.com/a-h/templ
github.com/gorilla/sessions (for cookie sessions)
```

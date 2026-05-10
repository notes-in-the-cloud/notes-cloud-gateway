# Notes Cloud API Gateway

A reverse proxy that serves as the single entry point for all Notes Cloud platform services. It handles JWT authentication and routes requests to the appropriate backend microservices.

## What it does

1. **Single Entry Point** — All client requests go through the gateway instead of directly to individual services
2. **JWT Authentication** — Validates access tokens for protected endpoints using shared JWT configuration
3. **Request Routing** — Forwards requests to the appropriate backend service based on the URL path

## Architecture

```
                                    ┌─────────────────┐
                                    │  Auth Service   │
                                ┌──▶│   (port 8081)   │
                                │   └─────────────────┘
                                │   ┌─────────────────┐
┌──────────┐   ┌─────────────┐  │   │  Notes Service  │
│  Client  │──▶│ API Gateway │──┼──▶│   (port 8082)   │
│          │   │ (port 8090) │  │   └─────────────────┘
└──────────┘   └─────────────┘  │   ┌─────────────────┐
                                ├──▶│ Sharing Service │
                                │   │   (port 8083)   │
                                │   └─────────────────┘
                                │   ┌─────────────────┐
                                ├──▶│Reminder Service │
                                │   │   (port 8084)   │
                                │   └─────────────────┘
                                │   ┌─────────────────┐
                                └──▶│  Todo Service   │
                                    │   (port 8085)   │
                                    └─────────────────┘
```

## Endpoints

### Health Checks

| Gateway | Method | Description |
|---------|--------|-------------|
| `/api/healthz` | GET | Liveness probe |
| `/api/readyz` | GET | Readiness probe |

### Authentication (Public) → Auth Service

| Gateway | Method | Auth Service |
|---------|--------|--------------|
| `/api/v1/auth/register` | POST | `/authService/api/v1/register` |
| `/api/v1/auth/login` | POST | `/authService/api/v1/login` |
| `/api/v1/auth/logout` | POST | `/authService/api/v1/logout` |
| `/api/v1/auth/refresh` | POST | `/authService/api/v1/refresh` |
| `/api/v1/auth/verify` | POST | `/authService/api/v1/verify` |
| `/api/v1/auth/resend` | POST | `/authService/api/v1/resend` |

### OAuth (Public) → Auth Service

| Gateway | Method | Auth Service |
|---------|--------|--------------|
| `/api/v1/auth/google/start` | GET | `/authService/api/v1/auth/google/start` |
| `/api/v1/auth/google/callback` | GET | `/authService/api/v1/auth/google/callback` |
| `/api/v1/auth/gitlab/start` | GET | `/authService/api/v1/auth/gitlab/start` |
| `/api/v1/auth/gitlab/callback` | GET | `/authService/api/v1/auth/gitlab/callback` |

### User (Protected) → Auth Service

| Gateway | Method | Auth Service |
|---------|--------|--------------|
| `/api/v1/me` | GET | `/authService/api/v1/me` |

### Notes (Protected) → Notes Service

| Gateway | Method | Notes Service |
|---------|--------|---------------|
| `/api/v1/notes` | GET | `/api/v1/notes` |
| `/api/v1/notes` | POST | `/api/v1/notes` |
| `/api/v1/notes/{note_id}` | GET | `/api/v1/notes/{note_id}` |
| `/api/v1/notes/{note_id}` | PUT | `/api/v1/notes/{note_id}` |
| `/api/v1/notes/{note_id}` | DELETE | `/api/v1/notes/{note_id}` |

### Sharing → Sharing Service

| Gateway | Method | Auth | Sharing Service |
|---------|--------|------|-----------------|
| `/api/v1/share-links/{token}` | GET | Public | `/api/v1/share-links/{token}` |
| `/api/v1/notes/{note_id}/share-links` | POST | Protected | `/api/v1/notes/{note_id}/share-links` |

### Todos (Protected) → Todo Service

| Gateway | Method | Todo Service |
|---------|--------|--------------|
| `/api/v1/todos` | GET | `/api/v1/todos` |
| `/api/v1/todos` | POST | `/api/v1/todos` |
| `/api/v1/todos/{todo_id}` | GET | `/api/v1/todos/{todo_id}` |
| `/api/v1/todos/{todo_id}` | PUT | `/api/v1/todos/{todo_id}` |
| `/api/v1/todos/{todo_id}` | DELETE | `/api/v1/todos/{todo_id}` |

### Todo Lists (Protected) → Todo Service

| Gateway | Method | Todo Service |
|---------|--------|--------------|
| `/api/v1/todo-lists` | GET | `/api/v1/todo-lists` |
| `/api/v1/todo-lists` | POST | `/api/v1/todo-lists` |
| `/api/v1/todo-lists/{list_id}` | GET | `/api/v1/todo-lists/{list_id}` |
| `/api/v1/todo-lists/{list_id}` | PUT | `/api/v1/todo-lists/{list_id}` |
| `/api/v1/todo-lists/{list_id}` | DELETE | `/api/v1/todo-lists/{list_id}` |

### Reminders (Protected) → Reminder Service

| Gateway | Method | Reminder Service |
|---------|--------|------------------|
| `/api/v1/reminders` | GET | `/api/users/{userId}/reminders` |
| `/api/v1/reminders` | POST | `/api/users/{userId}/reminders` |
| `/api/v1/reminders` | PUT | `/api/users/{userId}/reminders` |
| `/api/v1/reminders/{reminder_id}` | GET | `/api/users/{userId}/reminders/{reminder_id}` |
| `/api/v1/reminders/{reminder_id}` | DELETE | `/api/users/{userId}/reminders/{reminder_id}` |

### Notifications (Protected) → Reminder Service

| Gateway | Method | Reminder Service |
|---------|--------|------------------|
| `/api/v1/notifications` | GET | `/api/users/{userId}/notifications` |
| `/api/v1/notifications` | DELETE | `/api/users/{userId}/notifications` |
| `/api/v1/notifications/unread-count` | GET | `/api/users/{userId}/notifications/unread-count` |
| `/api/v1/notifications/read-all` | POST | `/api/users/{userId}/notifications/read-all` |
| `/api/v1/notifications/{notification_id}/read` | POST | `/api/users/{userId}/notifications/{notification_id}/read` |

## Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `SERVER_PORT` | Port the gateway listens on | `8090` |
| `AUTH_SERVICE_URL` | Auth service base URL | `http://localhost:8081` |
| `NOTES_SERVICE_URL` | Notes service base URL | `http://localhost:8082` |
| `SHARING_SERVICE_URL` | Sharing service base URL | `http://localhost:8083` |
| `REMINDER_SERVICE_URL` | Reminder service base URL | `http://localhost:8084` |
| `TODO_SERVICE_URL` | Todo service base URL | `http://localhost:8085` |
| `JWT_SECRET` | Secret key for JWT validation | — |
| `JWT_ISSUER` | Expected JWT issuer | — |
| `JWT_AUDIENCE` | Expected JWT audience | — |

## Running Locally

```bash
# Set required environment variables
export JWT_SECRET="your-secret"
export JWT_ISSUER="notes-cloud"
export JWT_AUDIENCE="notes-cloud"

# Run the gateway
go run ./cmd/api-gateway
```

## Docker

```bash
# Build
docker build -t notes-cloud-api-gateway .

# Run
docker run -p 8090:8090 \
  -e JWT_SECRET="your-secret" \
  -e JWT_ISSUER="notes-cloud" \
  -e JWT_AUDIENCE="notes-cloud" \
  -e AUTH_SERVICE_URL="http://auth-service:8081" \
  -e NOTES_SERVICE_URL="http://notes-service:8082" \
  -e SHARING_SERVICE_URL="http://sharing-service:8083" \
  -e REMINDER_SERVICE_URL="http://reminder-service:8084" \
  -e TODO_SERVICE_URL="http://todo-service:8085" \
  notes-cloud-api-gateway
```

## Kubernetes

```bash
# Port forward to access locally
kubectl port-forward -n notes-cloud svc/api-gateway 8090:8090
```

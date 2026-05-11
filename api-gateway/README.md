# Notes Cloud API Gateway

A reverse proxy that serves as the single entry point for all Notes Cloud platform services. It handles JWT authentication and routes requests to the appropriate backend microservices.

## What it does

1. **Single Entry Point** — All client requests go through the gateway instead of directly to individual services
2. **JWT Authentication** — Validates access tokens for protected endpoints using shared JWT configuration
3. **Request Routing** — Forwards requests to the appropriate backend service based on the URL path
4. **CORS Handling** — Manages Cross-Origin Resource Sharing with credential support for cookie-based authentication
5. **Cookie Propagation** — Forwards cookies between frontend and backend services for secure token management

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
| `INTERNAL_SERVICE_TOKEN` | Token for internal service-to-service calls | — |
| `JWT_SECRET` | Secret key for JWT validation | — |
| `JWT_ISSUER` | Expected JWT issuer | — |
| `JWT_AUDIENCE` | Expected JWT audience | — |
| `ALLOWED_ORIGINS` | Comma-separated list of allowed CORS origins | `http://localhost:5173,http://localhost:3000` |

### CORS & Cookie Configuration

The gateway supports **cookie-based authentication** with proper CORS configuration:

- **`Access-Control-Allow-Credentials: true`** — Allows cookies to be sent cross-origin
- **`Access-Control-Allow-Origin`** — Set to the specific requesting origin (not `*` when credentials are used)
- **Allowed Methods:** GET, POST, PUT, DELETE, OPTIONS
- **Allowed Headers:** Accept, Authorization, Content-Type

**Important:** When using cookie-based authentication:
1. Frontend must set `credentials: 'include'` in fetch requests
2. OAuth redirect URLs must point to the gateway (not directly to auth-service) to ensure cookies are set from the gateway's domain/port
3. `ALLOWED_ORIGINS` must include your frontend URL for CORS to work

Example frontend fetch:
```typescript
const res = await fetch('http://localhost:8090/api/v1/notes', {
  method: 'GET',
  credentials: 'include', // Required for cookies
  headers: {
    'Authorization': `Bearer ${accessToken}`
  }
});
```

## Running Locally

```bash
# Set required environment variables
export JWT_SECRET="your-secret"
export JWT_ISSUER="notes-cloud"
export JWT_AUDIENCE="notes-cloud"
export INTERNAL_SERVICE_TOKEN="your-internal-token"
export ALLOWED_ORIGINS="http://localhost:5173,http://localhost:3000"

# Run the gateway
go run ./cmd/api-gateway
```

The gateway will start on `http://localhost:8090`. Make sure all backend services are running on their respective ports.

## Docker

```bash
# Build multi-platform image
docker buildx build --platform linux/amd64,linux/arm64 \
  -t hristo12319/notes-cloud-api-gateway:latest --push .

# Run locally
docker run -p 8090:8090 \
  -e JWT_SECRET="your-secret" \
  -e JWT_ISSUER="notes-cloud" \
  -e JWT_AUDIENCE="notes-cloud" \
  -e INTERNAL_SERVICE_TOKEN="your-internal-token" \
  -e ALLOWED_ORIGINS="http://localhost:5173" \
  -e AUTH_SERVICE_URL="http://auth-service:8081" \
  -e NOTES_SERVICE_URL="http://notes-service:8082" \
  -e SHARING_SERVICE_URL="http://sharing-service:8083" \
  -e REMINDER_SERVICE_URL="http://reminder-service:8084" \
  -e TODO_SERVICE_URL="http://todo-service:8085" \
  hristo12319/notes-cloud-api-gateway:latest
```

## Kubernetes

```bash
# Port forward to access locally
kubectl port-forward -n notes-cloud svc/api-gateway 8090:8090

# Restart deployment after config changes
kubectl rollout restart deployment/api-gateway -n notes-cloud
```

## WebSocket Support

The gateway provides WebSocket support for real-time notifications at `/ws`.

**Important:** Since browser WebSocket API cannot easily send `Authorization` headers, the WebSocket endpoint supports passing the access token as a query parameter:

```javascript
const ws = new WebSocket(`ws://localhost:8090/ws?token=${accessToken}`);
```

The gateway validates the token using the same JWT configuration as HTTP endpoints.

## Internal Service Communication

The gateway exposes internal endpoints for service-to-service communication at `/internal/*`. These endpoints require the `X-Internal-Token` header matching the `INTERNAL_SERVICE_TOKEN` environment variable.

Example: Reminder service pushes notifications via:
```
POST /internal/notifications/{userId}
X-Internal-Token: <internal-token>
```

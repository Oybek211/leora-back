# Leora Server

Go (Fiber v2) backend for the Leora app.

## Local Development

```bash
# Start Postgres, Redis, and the backend
docker-compose up --build
```

The server runs at `http://localhost:9090/api/v1`.

## Deploy to Render

1. Push this repository to GitHub/GitLab.
2. In Render dashboard: **New → Blueprint → select your repo**.
3. Render detects `render.yaml` and provisions:
   - **leora-server** (web service, Docker)
   - **leora-postgres** (PostgreSQL 15)
   - **leora-redis** (Redis / Key Value store)
4. Fill in the secrets when prompted:
   - `JWT_SECRET` — a strong random string
   - `GOOGLE_OAUTH_CLIENT_ID` — your Google OAuth client ID
   - `APPLE_KEY_ID` — your Apple Sign-In key ID
   - `CORS_ORIGINS` — comma-separated allowed origins (e.g. `https://your-app.com`)
5. Click **Apply** and wait for the build to complete.
6. Verify the deploy: `GET https://<your-service>.onrender.com/api/v1/health` should return `{"status":"ok"}`.

## Environment Variables

See [.env.example](.env.example) for the full list.

| Variable | Required | Description |
|---|---|---|
| `PORT` | Set by Render | HTTP listen port (default: 9090) |
| `APP_ENV` | No | `local` or `production` |
| `DATABASE_URL` | Production | Postgres connection string |
| `REDIS_URL` | Production | Redis connection string |
| `JWT_SECRET` | Yes | Secret for signing JWT tokens |
| `GOOGLE_OAUTH_CLIENT_ID` | Yes | Google OAuth2 client ID |
| `APPLE_BUNDLE_ID` | No | iOS bundle ID (default: `com.sarvar.leora`) |
| `APPLE_TEAM_ID` | No | Apple developer team ID |
| `APPLE_KEY_ID` | Yes | Apple Sign-In key ID |
| `CORS_ORIGINS` | No | Comma-separated allowed origins |

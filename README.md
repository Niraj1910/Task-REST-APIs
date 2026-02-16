# Task REST API

Secure, production-grade REST API built in Go for task management with JWT authentication and email verification.

**Live Demo**: https://task-rest-apis.onrender.com  
**Docker Hub Image**: https://hub.docker.com/r/niraj1910/task-rest-api

[![Go Tests](https://github.com/Niraj1910/Task-REST-APIs/actions/workflows/tests.yml/badge.svg)](https://github.com/Niraj1910/Task-REST-APIs/actions)

## Features

- **User Authentication**
  - Registration with email verification (async via Resend)
  - Login → JWT token + HttpOnly cookie
  - Logout
  - Get current user profile (`GET /api/users/me`)
  - Partial profile update (`PATCH /api/users/me`)

- **Task Management** (protected routes)
  - Full CRUD with strict ownership (`user_id` from JWT)
  - List tasks: pagination (`page`, `limit`), filtering (`status`), sorting (`created_at`, `priority`)
  - Default: 10 newest tasks first

- **Security & Reliability**
  - JWT middleware
  - Password hashing (bcrypt)
  - Input validation (Gin binding)
  - Atomic transaction for verification
  - Non-blocking email sending
  - CORS support

- **Observability**
  - Structured logging with zerolog

- **Testing**
  - Unit & integration tests (auth, profile, tasks)
  - Isolated in-memory SQLite per test

- **Deployment**
  - Multi-stage Dockerfile
  - docker-compose for local dev
  - Live on Render (free tier)
  - Cloud Postgres via Supabase (free tier)

## Tech Stack

- Go 1.25
- Gin
- GORM + PostgreSQL (Supabase)
- JWT (golang-jwt/jwt/v5)
- Zerolog
- Resend (email)
- Testify
- Docker & Docker Compose
- GitHub Actions (CI)

## Quick Start (Docker – recommended)

```bash
git clone https://github.com/Niraj1910/Task-REST-APIs
cd Task-REST-APIs

cp .env.example .env
# Edit .env (Supabase URI, JWT_SECRET, Resend key)

docker compose up --build

docker pull niraj1910/task-rest-api:latest

docker run -d -p 4000:4000 \
  -e POSTGRES_URI="your-supabase-uri" \
  -e JWT_SECRET="your-secret" \
  -e RESEND_API_KEY="re_xxx" \
  niraj1910/task-rest-api:latest
  ```

## Important note for IPv4
Some cloud runtimes (e.g. Render free tier) have IPv6 issues.
Use the IPv4 pooled connection URI from Supabase (port 6543) to avoid "network is unreachable".

## API Endpoints
**Public**

- POST /register
- GET /verify?token=...&email=...
- POST /login
- POST /logout

**Protected (JWT required)**

- GET /api/user/profile
- GET /api/user/task (owner's tasks)
- PATCH /api/user/update
- GET /api/task (paginated, filterable, sortable)
- POST /api/task/new
- GET /api/task/:id
- PUT /api/task/:id
- DELETE /api/task/:id

**Testing**
```bash 
go test ./... -v
go test ./handlers -v

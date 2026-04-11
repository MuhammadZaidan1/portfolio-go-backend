# Jay's Portfolio API

A Go-powered backend for a personal portfolio, built with Gin, GORM, and Supabase. This API serves portfolio content like skills, experiences, projects, certificates, CV downloads, and contact messages, while also providing protected admin routes for content management.

## Features

- Public portfolio endpoints for skills, experiences, projects, certificates, CV, and contact form submissions
- Admin CRUD operations for projects, skills, experiences, certificates, CV uploads, and messages
- JWT authentication for admin routes
- Supabase Storage integration for CVs, project images, and certificate files
- PostgreSQL database connection with auto-migrations via GORM
- CORS-ready setup for frontend integration

## Tech stack

- Go 1.25.1
- Gin Web Framework
- GORM ORM
- Supabase PostgreSQL
- Supabase Storage
- JWT auth
- Bcrypt password hashing
- godotenv for local environment loading

## Repository structure

- `cmd/api/main.go` - main API server entrypoint
- `cmd/seed/main.go` - admin user seed tool
- `internal/config/database.go` - environment loading, database connection, auto-migration
- `internal/handlers/` - request handlers for public and admin routes
- `internal/models/models.go` - domain models and database schema definitions
- `internal/storage/supabase.go` - Supabase storage helper utilities
- `internal/utils/` - JWT token and password utilities

## Environment variables

Create a `.env` file or set these variables in your deployment environment:

```env
DATABASE_URL=<your supabase postgres connection string>
JWT_SECRET=<your jwt secret>
SUPABASE_URL=<your supabase project url>
SUPABASE_SERVICE_KEY=<your supabase service role key>
FRONTEND_URL=http://localhost:5173
PORT=8080
```

## Setup

1. Clone this repository:

```bash
git clone https://github.com/<your-username>/<repo-name>.git
cd go-backend
```

2. Install dependencies:

```bash
go mod download
```

3. Create or configure your `.env` file with the required environment variables.

4. Run the API server:

```bash
go run ./cmd/api
```

5. Seed the initial admin account:

```bash
go run ./cmd/seed
```

> The seed command creates a default admin username and hashed password in the database. Update `cmd/seed/main.go` before running if you want custom credentials.

## API Endpoints

### Public routes

- `GET /api/ping`
- `GET /api/skills`
- `GET /api/experiences`
- `GET /api/projects`
- `GET /api/projects/:id`
- `GET /api/project-tech-stacks`
- `GET /api/certificates`
- `POST /api/login`
- `GET /api/cv`
- `POST /api/contact`
- `PATCH /api/cv/download`

### Admin routes

Protected by JWT middleware under `/api/admin`:

- `POST /api/admin/projects`
- `PUT /api/admin/projects/:id`
- `DELETE /api/admin/projects/:id`
- `POST /api/admin/projects/:id/images`
- `DELETE /api/admin/projects/images/:image_id`
- `PATCH /api/admin/projects/images/:image_id/sort`
- `GET /api/admin/project-tech-stacks`
- `POST /api/admin/project-tech-stacks`
- `POST /api/admin/project-tech-stacks/import`
- `PUT /api/admin/project-tech-stacks/:id`
- `DELETE /api/admin/project-tech-stacks/:id`
- `POST /api/admin/experiences`
- `PUT /api/admin/experiences/:id`
- `DELETE /api/admin/experiences/:id`
- `POST /api/admin/certificates`
- `PUT /api/admin/certificates/:id`
- `DELETE /api/admin/certificates/:id`
- `POST /api/admin/certificates/:id/file`
- `POST /api/admin/skill-categories`
- `PUT /api/admin/skill-categories/:id`
- `DELETE /api/admin/skill-categories/:id`
- `POST /api/admin/skills`
- `PUT /api/admin/skills/:id`
- `DELETE /api/admin/skills/:id`
- `POST /api/admin/cv`
- `GET /api/admin/messages`
- `PATCH /api/admin/messages/:id/read`
- `DELETE /api/admin/messages/:id`

## Notes

- The app auto-migrates database tables on startup.
- Supabase Storage is used for CV uploads, project images, and certificate files.
- The admin login returns a JWT token used to access protected admin routes.

## Optional additions

If you want, I can also add:

- usage examples for each endpoint
- a sample `.env.example`
- deployment instructions for Supabase, Railway, or Render
- shields/badges for Go, license, build status

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Weekly Watch is a movie recommendation web application built with Go and MySQL for CS 4604. It serves an HTML UI from a single Go binary with no frontend build step.

## Commands

```bash
# Install dependencies (run from backend/)
cd backend && go mod tidy

# Run the application (from project root — needed so the app can reference frontend/)
go run ./backend

# Or run from backend/ directly
cd backend && go run .

# Build binary
cd backend && go build -o weekly-watch .

# Initialize/reset the database (destructive — drops and recreates everything)
mysql -u root -p < db/schema.sql
```

There are no tests, linters, or CI configured.

## Prerequisites

- Go 1.23+
- MySQL 8.x running locally
- Database must be initialized by running `db/schema.sql` in MySQL before starting the app

## Architecture

Single-file Go application (`backend/main.go`) with everything colocated:

- **Database config**: Hardcoded constants at the top. Each developer must update `dbUser`/`dbPass` to match their local MySQL credentials.
- **Authentication**: Cookie-based sessions via `gorilla/sessions`. Passwords hashed with `bcrypt`. Two roles: admin and regular user. Session cookie name: `weekly-watch-session`.
- **Auth helpers**: `getCurrentUser(r)` reads session, `requireAuth(handler)` and `requireAdmin(handler)` are middleware wrappers used in route registration.
- **Data models**: Go structs map to MySQL tables defined in `db/schema.sql`. Key types: `User` (includes `PasswordHash`, `IsAdmin`), `Movie`, `Genre`, `ViewingHistory`, `WeeklyRecommendation`, `Rating`, `Review`, `PopularMovie`, `UserActivity`, `PageData` (template context).
- **Query functions**: `getUsers`, `getMovies`, `getGenres`, `getUserViewingHistory`, `getUserRecommendations`, `getUserRatings`, `getUserReviews`, `getPopularMovies`, `getUserActivitySummary`.
- **HTTP handlers**: Tab-based navigation with `?tab=` query params and flash messages via `?msg=`/`?msg_type=`. Routes:
  - Public: `GET/POST /login`, `GET/POST /signup`, `GET /logout`, `GET /api/status`
  - Auth-required: `GET /` (dashboard), `POST /change-password`, rating/review CRUD
  - Admin-only: movie CRUD, user CRUD (`/insert/movie`, `/delete/movie`, etc.)
- **HTML templates**: Three inline string constants — `indexHTML` (main app), `loginHTML`, `signupHTML`. The file `frontend/index.html` is an older version and is **not** loaded by the Go server.
- **Reports**: Reports tab with two aggregate queries — "Most Popular Movies" (JOIN + GROUP BY + COUNT) and "User Activity Summary" (LEFT JOIN with subqueries + GROUP BY + SUM).

## Database

`db/schema.sql` creates the `weekly_watch` database with 18 tables in 3NF. The schema file is destructive — it runs `DROP DATABASE IF EXISTS weekly_watch` on every execution.

Key entity tables: User (with `password_hash` and `is_admin`), Movie, Person, Genre, Streaming_Service. Bridge tables: Movie_Genre, Movie_Streaming, Movie_Contributor, Favorites_Movie, Subscription. The schema includes sample data inserts (~20 rows per major table). All seed users have password `password123`. User `tom_w` is the admin.

## SQL Injection Protection

All database queries use parameterized placeholders (`?`) via Go's `database/sql` package. No raw string concatenation is used to build SQL queries. This prevents SQL injection by ensuring user input is always treated as data, never as SQL code.

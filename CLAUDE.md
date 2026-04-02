# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Weekly Watch is a movie recommendation web application built with Go and MySQL. It serves an HTML UI from a single Go binary with no frontend build step. This is a CS 4604 course project.

## Commands

```bash
# Install dependencies
go mod tidy

# Run the application (serves on http://localhost:8080)
go run main.go

# Build binary
go build -o weekly-watch .
```

There are no tests, linters, or CI configured.

## Prerequisites

- Go 1.21+
- MySQL 8.x running locally
- Database must be initialized by running `schema.sql` in MySQL before starting the app

## Architecture

This is a single-file application (`main.go`) with everything colocated:

- **Database config**: Hardcoded constants at the top of `main.go` (lines 36-42). Each developer must update `dbUser`/`dbPass` to match their local MySQL credentials.
- **Data models**: Go structs (lines 50-101) map to MySQL tables defined in `schema.sql`
- **Query functions**: `getUsers`, `getMovies`, `getUserViewingHistory`, `getUserRecommendations`, `getUserRatings` — each returns typed slices from SQL queries
- **HTTP handlers**: Two routes — `/` (HTML dashboard via Go templates) and `/api/status` (JSON health check)
- **HTML template**: Embedded as a raw string constant `indexHTML` at the bottom of `main.go` (not a separate file)

## Database Schema

`schema.sql` creates the `weekly_watch` database with 18 tables in 3NF. Key entity tables: User, Movie, Person, Genre, Streaming_Service. Bridge tables: Movie_Genre, Movie_Streaming, Movie_Contributor, Favorites_Movie, Subscription. The schema includes sample data inserts.

The schema file is destructive — it drops and recreates the entire database on each run (`DROP DATABASE IF EXISTS weekly_watch`).

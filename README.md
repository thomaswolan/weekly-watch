# The Weekly Watch - Phase 4 Setup Guide

## DBMS Information
- **DBMS:** MySQL 8.x
- **Programming Language:** Go (Golang)
- **MySQL Driver:** github.com/go-sql-driver/mysql
- **Interface:** Web-based (HTML served by Go's net/http)

---

## Prerequisites

### 1. Install MySQL (Mac)
```bash
# Using Homebrew
brew install mysql

# Start MySQL
brew services start mysql

# Secure the installation (set root password)
mysql_secure_installation
```

### 2. Install Go (Mac)
```bash
# Using Homebrew
brew install go

# Verify
go version
```

---

## Setup Steps

### Step 1: Create the Database
```bash
# Log into MySQL
mysql -u root -p

# Inside MySQL, run the schema file:
source /path/to/db/schema.sql;

# Verify tables were created:
USE weekly_watch;
SHOW TABLES;

# Verify sample data:
SELECT * FROM User;
SELECT * FROM Movie;
```

### Step 2: Configure the Go App
Open `backend/main.go` and update the database credentials at the top:
```go
const (
    dbUser = "root"        // Your MySQL username
    dbPass = "password"    // Your MySQL password
    dbHost = "127.0.0.1"
    dbPort = "3306"
    dbName = "weekly_watch"
)
```

### Step 3: Install Go Dependencies
```bash
cd /path/to/project/backend
go mod tidy
```

### Step 4: Run the Application
From the repository root (recommended, so the app finds `frontend/index.html`):

```bash
cd /path/to/project
go run ./backend
```

Or from `backend/`:

```bash
cd /path/to/project/backend
go run .
```

You should see:
```
===========================================
  The Weekly Watch - Movie Recommendation
  CS 4604 Phase 4
===========================================

Connecting to MySQL at 127.0.0.1:3306/weekly_watch...
Successfully connected to MySQL!

Tables in database:
  - Favorites_List
  - Favorites_Movie
  - Genre
  - Movie
  - Movie_Contributor
  - Movie_Genre
  - Movie_Streaming
  - Notification
  - Person
  - Rating
  - Review
  - Role
  - Streaming_Service
  - Subscription
  - User
  - User_Preference
  - Viewing_History
  - Weekly_Recommendation

Starting web server on http://localhost:8080
```

### Step 5: Open the App
Navigate to **http://localhost:8080** in your browser.

---

## Deliverable Screenshots Checklist

1. **Normalized Schema** — Use `db/schema.sql` or a screenshot of tables in MySQL Workbench
2. **Tables with Sample Data** — Screenshot the tables in MySQL Workbench showing 5+ rows each
3. **DBMS-Interface Connection** — Screenshot the running Go app in terminal + the web page showing "Connected to MySQL"

---

## Project Structure
```
weekly-watch/
├── README.md
├── docs/              # Project documentation (placeholders for now)
├── db/
│   └── schema.sql     # MySQL schema + sample data
├── backend/
│   ├── main.go        # Go web server and DB access
│   ├── go.mod
│   └── go.sum
├── frontend/
│   └── index.html     # Go html/template for the main page
├── reports/           # Generated or exported reports
└── roles/             # Role-related assets or notes (optional)
```

---

## Normalization Notes

All tables are in **3NF (Third Normal Form)**:

- **1NF**: All attributes are atomic (no repeating groups, no multi-valued attributes)
- **2NF**: No partial dependencies (all non-key attributes depend on the entire primary key; bridge tables use composite keys appropriately)
- **3NF**: No transitive dependencies (non-key attributes depend only on the primary key, not on other non-key attributes)

Key normalization decisions:
- Movie_Genre, Movie_Streaming, Movie_Contributor, and Favorites_Movie are bridge tables resolving M:N relationships
- Subscription table was added in Phase 3 to properly model the User–Streaming_Service relationship with its own attributes
- Rating has a UNIQUE constraint on (user_id, movie_id) to prevent duplicate ratings
- rating_value was added to the Rating table (was missing in Phase 3 schema)

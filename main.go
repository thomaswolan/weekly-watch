// =============================================================
// The Weekly Watch - Go Web Application
// CS 4604 Phase 4: DBMS-Interface Connection
//
// Prerequisites:
//   1. Go installed (https://go.dev/dl/)
//   2. MySQL 8.x installed and running
//   3. Run schema.sql in MySQL first
//
// Setup:
//   go mod init weekly-watch
//   go get github.com/go-sql-driver/mysql
//   go run main.go
//
// Then open http://localhost:8080 in your browser.
// =============================================================

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// ==========================
// Configuration
// ==========================

const (
	dbUser = "root"        // Change to your MySQL username
	dbPass = "Hansolo13!?!"    // Change to your MySQL password
	dbHost = "127.0.0.1"
	dbPort = "3306"
	dbName = "weekly_watch"
)

var db *sql.DB

// ==========================
// Data Models
// ==========================

type User struct {
	UserID    int
	Username  string
	Email     string
	CreatedAt string
	LastLogin sql.NullString
}

type Movie struct {
	MovieID     int
	Title       string
	PlotSummary string
	TrailerURL  string
	TmdbID      string
	Genres      string
}

type ViewingHistory struct {
	ViewingID        int
	MovieTitle       string
	WatchedDate      string
	CompletionStatus string
	WatchCount       int
	Notes            string
}

type WeeklyRecommendation struct {
	RecommendationID int
	MovieTitle       string
	AssignedDate     string
	DueDate          string
	Status           string
}

type Rating struct {
	RatingID    int
	MovieTitle  string
	RatingValue string
	RatedAt     string
}

type PageData struct {
	Users               []User
	Movies              []Movie
	ViewingHistory      []ViewingHistory
	Recommendations     []WeeklyRecommendation
	Ratings             []Rating
	SelectedUser        *User
	Connected           bool
	DBInfo              string
	Error               string
}

// ==========================
// Database Connection
// ==========================

func connectDB() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPass, dbHost, dbPort, dbName)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// ==========================
// Query Functions
// ==========================

func getUsers() ([]User, error) {
	rows, err := db.Query("SELECT user_id, username, email, created_at, last_login FROM User ORDER BY user_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var createdAt time.Time
		err := rows.Scan(&u.UserID, &u.Username, &u.Email, &createdAt, &u.LastLogin)
		if err != nil {
			return nil, err
		}
		u.CreatedAt = createdAt.Format("Jan 2, 2006")
		users = append(users, u)
	}
	return users, nil
}

func getMovies() ([]Movie, error) {
	query := `
		SELECT m.movie_id, m.title,
			COALESCE(m.plot_summary, '') as plot_summary,
			COALESCE(m.trailer_url, '') as trailer_url,
			COALESCE(m.tmdb_id, '') as tmdb_id,
			COALESCE(GROUP_CONCAT(g.genre_name SEPARATOR ', '), 'N/A') as genres
		FROM Movie m
		LEFT JOIN Movie_Genre mg ON m.movie_id = mg.movie_id
		LEFT JOIN Genre g ON mg.genre_id = g.genre_id
		GROUP BY m.movie_id
		ORDER BY m.title
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var m Movie
		err := rows.Scan(&m.MovieID, &m.Title, &m.PlotSummary, &m.TrailerURL, &m.TmdbID, &m.Genres)
		if err != nil {
			return nil, err
		}
		movies = append(movies, m)
	}
	return movies, nil
}

func getUserViewingHistory(userID int) ([]ViewingHistory, error) {
	query := `
		SELECT vh.viewing_id, m.title, vh.watched_date, vh.completion_status,
			vh.watch_count, COALESCE(vh.notes, '') as notes
		FROM Viewing_History vh
		JOIN Movie m ON vh.movie_id = m.movie_id
		WHERE vh.user_id = ?
		ORDER BY vh.watched_date DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []ViewingHistory
	for rows.Next() {
		var vh ViewingHistory
		var watchedDate time.Time
		err := rows.Scan(&vh.ViewingID, &vh.MovieTitle, &watchedDate,
			&vh.CompletionStatus, &vh.WatchCount, &vh.Notes)
		if err != nil {
			return nil, err
		}
		vh.WatchedDate = watchedDate.Format("Jan 2, 2006")
		history = append(history, vh)
	}
	return history, nil
}

func getUserRecommendations(userID int) ([]WeeklyRecommendation, error) {
	query := `
		SELECT wr.recommendation_id, m.title, wr.assigned_date, wr.due_date, wr.status
		FROM Weekly_Recommendation wr
		JOIN Movie m ON wr.movie_id = m.movie_id
		WHERE wr.user_id = ?
		ORDER BY wr.assigned_date DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recs []WeeklyRecommendation
	for rows.Next() {
		var wr WeeklyRecommendation
		var assigned, due time.Time
		err := rows.Scan(&wr.RecommendationID, &wr.MovieTitle, &assigned, &due, &wr.Status)
		if err != nil {
			return nil, err
		}
		wr.AssignedDate = assigned.Format("Jan 2, 2006")
		wr.DueDate = due.Format("Jan 2, 2006")
		recs = append(recs, wr)
	}
	return recs, nil
}

func getUserRatings(userID int) ([]Rating, error) {
	query := `
		SELECT r.rating_id, m.title, r.rating_value, r.rated_at
		FROM Rating r
		JOIN Movie m ON r.movie_id = m.movie_id
		WHERE r.user_id = ?
		ORDER BY r.rated_at DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []Rating
	for rows.Next() {
		var rt Rating
		var ratedAt time.Time
		err := rows.Scan(&rt.RatingID, &rt.MovieTitle, &rt.RatingValue, &ratedAt)
		if err != nil {
			return nil, err
		}
		rt.RatedAt = ratedAt.Format("Jan 2, 2006")
		ratings = append(ratings, rt)
	}
	return ratings, nil
}

// ==========================
// HTTP Handlers
// ==========================

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{Connected: true, DBInfo: fmt.Sprintf("MySQL 8.x @ %s:%s / %s", dbHost, dbPort, dbName)}

	users, err := getUsers()
	if err != nil {
		data.Error = "Error fetching users: " + err.Error()
	}
	data.Users = users

	movies, err := getMovies()
	if err != nil {
		data.Error = "Error fetching movies: " + err.Error()
	}
	data.Movies = movies

	// Check if a user is selected
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr != "" {
		var userID int
		fmt.Sscanf(userIDStr, "%d", &userID)
		for i, u := range users {
			if u.UserID == userID {
				data.SelectedUser = &users[i]
				break
			}
		}
		if data.SelectedUser != nil {
			data.ViewingHistory, _ = getUserViewingHistory(userID)
			data.Recommendations, _ = getUserRecommendations(userID)
			data.Ratings, _ = getUserRatings(userID)
		}
	}

	tmpl, err := template.New("index").Parse(indexHTML)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	tmpl.Execute(w, data)
}

func apiStatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := map[string]interface{}{
		"connected": true,
		"dbms":      "MySQL 8.x",
		"database":  dbName,
		"host":      dbHost,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Quick table count check
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ?", dbName).Scan(&count)
	if err == nil {
		status["table_count"] = count
	}

	json.NewEncoder(w).Encode(status)
}

// ==========================
// HTML Template
// ==========================

const indexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>The Weekly Watch</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f0f0f; color: #e0e0e0; min-height: 100vh;
        }
        .header {
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            padding: 24px 40px; border-bottom: 2px solid #e94560;
            display: flex; justify-content: space-between; align-items: center;
        }
        .header h1 { color: #e94560; font-size: 28px; letter-spacing: 1px; }
        .connection-badge {
            background: #1b4332; color: #95d5b2; padding: 8px 16px;
            border-radius: 20px; font-size: 13px; font-weight: 600;
        }
        .connection-badge.error { background: #4a1520; color: #f4978e; }
        .container { max-width: 1200px; margin: 0 auto; padding: 32px 24px; }

        .db-info {
            background: #1a1a2e; border: 1px solid #2a2a4a; border-radius: 8px;
            padding: 16px 24px; margin-bottom: 32px; font-size: 14px; color: #8888aa;
        }
        .db-info strong { color: #e94560; }

        .section { margin-bottom: 40px; }
        .section h2 {
            color: #e94560; font-size: 20px; margin-bottom: 16px;
            padding-bottom: 8px; border-bottom: 1px solid #2a2a4a;
        }

        table {
            width: 100%; border-collapse: collapse; background: #1a1a2e;
            border-radius: 8px; overflow: hidden;
        }
        th {
            background: #16213e; color: #e94560; padding: 12px 16px;
            text-align: left; font-size: 13px; text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        td { padding: 10px 16px; border-bottom: 1px solid #2a2a4a; font-size: 14px; }
        tr:hover td { background: #1f1f3a; }

        .user-select {
            display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 24px;
        }
        .user-btn {
            background: #1a1a2e; color: #c0c0d0; border: 1px solid #3a3a5a;
            padding: 8px 18px; border-radius: 6px; cursor: pointer;
            text-decoration: none; font-size: 14px; transition: all 0.2s;
        }
        .user-btn:hover { border-color: #e94560; color: #e94560; }
        .user-btn.active { background: #e94560; color: white; border-color: #e94560; }

        .status-badge {
            display: inline-block; padding: 3px 10px; border-radius: 12px;
            font-size: 12px; font-weight: 600;
        }
        .status-completed { background: #1b4332; color: #95d5b2; }
        .status-pending { background: #3d2e00; color: #ffd166; }
        .status-loved { background: #4a1520; color: #f4978e; }
        .status-liked { background: #1b4332; color: #95d5b2; }
        .status-disliked { background: #2a2a4a; color: #8888aa; }

        .empty-state {
            text-align: center; padding: 32px; color: #555;
            font-style: italic;
        }

        .grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; }
        @media (max-width: 768px) { .grid-2 { grid-template-columns: 1fr; } }
    </style>
</head>
<body>
    <div class="header">
        <h1>The Weekly Watch</h1>
        {{if .Connected}}
            <span class="connection-badge">Connected to MySQL</span>
        {{else}}
            <span class="connection-badge error">Disconnected</span>
        {{end}}
    </div>

    <div class="container">
        {{if .Error}}
            <div class="db-info" style="border-color: #e94560;">
                <strong>Error:</strong> {{.Error}}
            </div>
        {{end}}

        <div class="db-info">
            <strong>DBMS:</strong> {{.DBInfo}} &nbsp;|&nbsp;
            <strong>Language:</strong> Go (Golang) &nbsp;|&nbsp;
            <strong>Driver:</strong> go-sql-driver/mysql &nbsp;|&nbsp;
            <strong>Status:</strong> Successfully Connected
        </div>

        <!-- USER SELECTOR -->
        <div class="section">
            <h2>Select a User</h2>
            <div class="user-select">
                {{range .Users}}
                    <a class="user-btn {{if $.SelectedUser}}{{if eq $.SelectedUser.UserID .UserID}}active{{end}}{{end}}"
                       href="/?user_id={{.UserID}}">
                        {{.Username}}
                    </a>
                {{end}}
            </div>
        </div>

        {{if .SelectedUser}}
        <!-- USER DASHBOARD -->
        <div class="section">
            <h2>Dashboard: {{.SelectedUser.Username}}</h2>

            <div class="grid-2">
                <!-- Viewing History -->
                <div>
                    <h2 style="font-size: 16px;">Viewing History</h2>
                    {{if .ViewingHistory}}
                    <table>
                        <tr><th>Movie</th><th>Watched</th><th>Count</th><th>Status</th></tr>
                        {{range .ViewingHistory}}
                        <tr>
                            <td>{{.MovieTitle}}</td>
                            <td>{{.WatchedDate}}</td>
                            <td>{{.WatchCount}}</td>
                            <td><span class="status-badge status-{{.CompletionStatus}}">{{.CompletionStatus}}</span></td>
                        </tr>
                        {{end}}
                    </table>
                    {{else}}<p class="empty-state">No viewing history yet.</p>{{end}}
                </div>

                <!-- Ratings -->
                <div>
                    <h2 style="font-size: 16px;">Ratings</h2>
                    {{if .Ratings}}
                    <table>
                        <tr><th>Movie</th><th>Rating</th><th>Date</th></tr>
                        {{range .Ratings}}
                        <tr>
                            <td>{{.MovieTitle}}</td>
                            <td><span class="status-badge status-{{.RatingValue}}">{{.RatingValue}}</span></td>
                            <td>{{.RatedAt}}</td>
                        </tr>
                        {{end}}
                    </table>
                    {{else}}<p class="empty-state">No ratings yet.</p>{{end}}
                </div>
            </div>
        </div>

        <!-- Weekly Recommendations -->
        <div class="section">
            <h2>Weekly Recommendations</h2>
            {{if .Recommendations}}
            <table>
                <tr><th>Movie</th><th>Assigned</th><th>Due</th><th>Status</th></tr>
                {{range .Recommendations}}
                <tr>
                    <td>{{.MovieTitle}}</td>
                    <td>{{.AssignedDate}}</td>
                    <td>{{.DueDate}}</td>
                    <td><span class="status-badge status-{{.Status}}">{{.Status}}</span></td>
                </tr>
                {{end}}
            </table>
            {{else}}<p class="empty-state">No recommendations yet.</p>{{end}}
        </div>
        {{end}}

        <!-- ALL MOVIES -->
        <div class="section">
            <h2>Movie Catalog</h2>
            <table>
                <tr><th>Title</th><th>Genres</th><th>TMDB ID</th></tr>
                {{range .Movies}}
                <tr>
                    <td><strong>{{.Title}}</strong></td>
                    <td>{{.Genres}}</td>
                    <td>{{.TmdbID}}</td>
                </tr>
                {{end}}
            </table>
        </div>

        <!-- ALL USERS -->
        <div class="section">
            <h2>Registered Users</h2>
            <table>
                <tr><th>ID</th><th>Username</th><th>Email</th><th>Joined</th></tr>
                {{range .Users}}
                <tr>
                    <td>{{.UserID}}</td>
                    <td>{{.Username}}</td>
                    <td>{{.Email}}</td>
                    <td>{{.CreatedAt}}</td>
                </tr>
                {{end}}
            </table>
        </div>
    </div>
</body>
</html>`

// ==========================
// Main
// ==========================

func main() {
	fmt.Println("===========================================")
	fmt.Println("  The Weekly Watch - Movie Recommendation  ")
	fmt.Println("  CS 4604 Phase 4                         ")
	fmt.Println("===========================================")
	fmt.Println()

	fmt.Printf("Connecting to MySQL at %s:%s/%s...\n", dbHost, dbPort, dbName)
	err := connectDB()
	if err != nil {
		log.Fatalf("Database connection failed: %v\n", err)
	}
	defer db.Close()

	fmt.Println("Successfully connected to MySQL!")
	fmt.Println()

	// Print table verification
	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = ? ORDER BY table_name", dbName)
	if err == nil {
		defer rows.Close()
		fmt.Println("Tables in database:")
		for rows.Next() {
			var name string
			rows.Scan(&name)
			fmt.Printf("  - %s\n", name)
		}
	}

	fmt.Println()
	fmt.Println("Starting web server on http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop.")

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/status", apiStatusHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

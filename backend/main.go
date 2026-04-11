// =============================================================
// The Weekly Watch - Go Web Application
// CS 4604 Phase 4: DBMS-Interface Connection
//
// Prerequisites:
//   1. Go installed (https://go.dev/dl/)
//   2. MySQL 8.x installed and running
//   3. Run db/schema.sql in MySQL first
//
// Setup:
//   go mod init weekly-watch
//   go get github.com/go-sql-driver/mysql
//   From repo root: go run ./backend
//   Or from backend/: go run .
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
	"os"
	"path/filepath"
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
var indexTemplate *template.Template

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

func loadIndexTemplate() error {
	candidates := []string{
		filepath.Join("frontend", "index.html"),
		filepath.Join("..", "frontend", "index.html"),
	}
	var lastErr error
	for _, p := range candidates {
		b, err := os.ReadFile(p)
		if err != nil {
			lastErr = err
			continue
		}
		t, err := template.New("index").Parse(string(b))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", p, err)
		}
		indexTemplate = t
		return nil
	}
	return fmt.Errorf("load frontend/index.html: tried %v: %w", candidates, lastErr)
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

	if indexTemplate == nil {
		http.Error(w, "index template not loaded", http.StatusInternalServerError)
		return
	}
	if err := indexTemplate.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	if err := loadIndexTemplate(); err != nil {
		log.Fatalf("Template load failed: %v\n", err)
	}

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

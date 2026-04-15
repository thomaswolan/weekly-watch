// =============================================================
// The Weekly Watch - Go Web Application
// CS 4604 Phase 5: Database Connectivity and UI Operations
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
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// ==========================
// Configuration
// ==========================

const (
	dbUser = "root"            // Change to your MySQL username
	dbPass = "SuperSecretPass!?!"    // Change to your MySQL password
	dbHost = "127.0.0.1"
	dbPort = "3306"
	dbName = "weekly_watch"
)

var db *sql.DB
var store = sessions.NewCookieStore([]byte("weekly-watch-secret-key-cs4604"))

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

// ==========================
// Data Models
// ==========================

type User struct {
	UserID       int
	Username     string
	Email        string
	PasswordHash string
	IsAdmin      bool
	CreatedAt    string
	LastLogin    sql.NullString
}

type Movie struct {
	MovieID     int
	Title       string
	PlotSummary string
	TrailerURL  string
	TmdbID      string
	Genres      string
}

type Genre struct {
	GenreID   int
	GenreName string
}

type ViewingHistory struct {
	ViewingID        int
	MovieID          int
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
	MovieID     int
	MovieTitle  string
	RatingValue string
	RatedAt     string
}

type Review struct {
	ReviewID   int
	MovieID    int
	MovieTitle string
	ReviewText string
	IsSpoiler  bool
	CreatedAt  string
}

type FlashMessage struct {
	Type    string // success, error
	Message string
}

type PopularMovie struct {
	Title       string
	RatingCount int
	LovedCount  int
	LikedCount  int
}

type UserActivity struct {
	Username       string
	WatchCount     int
	TotalRewatches int
	ReviewCount    int
	RatingCount    int
}

type PageData struct {
	Users           []User
	Movies          []Movie
	Genres          []Genre
	ViewingHistory  []ViewingHistory
	Recommendations []WeeklyRecommendation
	Ratings         []Rating
	Reviews         []Review
	SelectedUser    *User
	Connected       bool
	DBInfo          string
	Error           string
	Flash           *FlashMessage
	ActiveTab       string
	LoggedInUser    *User
	IsAdmin         bool
	PopularMovies   []PopularMovie
	UserActivities  []UserActivity
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
// Auth Helpers
// ==========================

func getCurrentUser(r *http.Request) *User {
	session, err := store.Get(r, "weekly-watch-session")
	if err != nil {
		return nil
	}
	userID, ok := session.Values["user_id"].(int)
	if !ok || userID == 0 {
		return nil
	}

	var u User
	var createdAt time.Time
	err = db.QueryRow(
		"SELECT user_id, username, email, password_hash, is_admin, created_at, last_login FROM User WHERE user_id = ?",
		userID,
	).Scan(&u.UserID, &u.Username, &u.Email, &u.PasswordHash, &u.IsAdmin, &createdAt, &u.LastLogin)
	if err != nil {
		return nil
	}
	u.CreatedAt = createdAt.Format("Jan 2, 2006")
	return &u
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if getCurrentUser(r) == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := getCurrentUser(r)
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if !user.IsAdmin {
			http.Redirect(w, r, "/?msg=Admin+access+required&msg_type=error", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// ==========================
// Query Functions
// ==========================

func getUsers() ([]User, error) {
	rows, err := db.Query("SELECT user_id, username, email, password_hash, is_admin, created_at, last_login FROM User ORDER BY user_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var createdAt time.Time
		err := rows.Scan(&u.UserID, &u.Username, &u.Email, &u.PasswordHash, &u.IsAdmin, &createdAt, &u.LastLogin)
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

func getGenres() ([]Genre, error) {
	rows, err := db.Query("SELECT genre_id, genre_name FROM Genre ORDER BY genre_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []Genre
	for rows.Next() {
		var g Genre
		err := rows.Scan(&g.GenreID, &g.GenreName)
		if err != nil {
			return nil, err
		}
		genres = append(genres, g)
	}
	return genres, nil
}

func getUserViewingHistory(userID int) ([]ViewingHistory, error) {
	query := `
		SELECT vh.viewing_id, vh.movie_id, m.title, vh.watched_date, vh.completion_status,
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
		err := rows.Scan(&vh.ViewingID, &vh.MovieID, &vh.MovieTitle, &watchedDate,
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
		SELECT r.rating_id, r.movie_id, m.title, r.rating_value, r.rated_at
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
		err := rows.Scan(&rt.RatingID, &rt.MovieID, &rt.MovieTitle, &rt.RatingValue, &ratedAt)
		if err != nil {
			return nil, err
		}
		rt.RatedAt = ratedAt.Format("Jan 2, 2006")
		ratings = append(ratings, rt)
	}
	return ratings, nil
}

func getUserReviews(userID int) ([]Review, error) {
	query := `
		SELECT r.review_id, r.movie_id, m.title, r.review_text, r.is_spoiler, r.created_at
		FROM Review r
		JOIN Movie m ON r.movie_id = m.movie_id
		WHERE r.user_id = ?
		ORDER BY r.created_at DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var rv Review
		var createdAt time.Time
		err := rows.Scan(&rv.ReviewID, &rv.MovieID, &rv.MovieTitle, &rv.ReviewText, &rv.IsSpoiler, &createdAt)
		if err != nil {
			return nil, err
		}
		rv.CreatedAt = createdAt.Format("Jan 2, 2006")
		reviews = append(reviews, rv)
	}
	return reviews, nil
}

// ==========================
// Report Queries
// ==========================

func getPopularMovies() ([]PopularMovie, error) {
	query := `
		SELECT m.title,
			COUNT(r.rating_id) AS rating_count,
			SUM(CASE WHEN r.rating_value = 'loved' THEN 1 ELSE 0 END) AS loved_count,
			SUM(CASE WHEN r.rating_value = 'liked' THEN 1 ELSE 0 END) AS liked_count
		FROM Movie m
		JOIN Rating r ON m.movie_id = r.movie_id
		GROUP BY m.movie_id, m.title
		ORDER BY rating_count DESC, loved_count DESC
		LIMIT 15
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PopularMovie
	for rows.Next() {
		var pm PopularMovie
		err := rows.Scan(&pm.Title, &pm.RatingCount, &pm.LovedCount, &pm.LikedCount)
		if err != nil {
			return nil, err
		}
		results = append(results, pm)
	}
	return results, nil
}

func getUserActivitySummary() ([]UserActivity, error) {
	query := `
		SELECT u.username,
			COALESCE(vh.watch_count, 0) AS watch_count,
			COALESCE(vh.total_rewatches, 0) AS total_rewatches,
			COALESCE(rev.review_count, 0) AS review_count,
			COALESCE(rat.rating_count, 0) AS rating_count
		FROM User u
		LEFT JOIN (
			SELECT user_id, COUNT(*) AS watch_count, SUM(watch_count) AS total_rewatches
			FROM Viewing_History
			GROUP BY user_id
		) vh ON u.user_id = vh.user_id
		LEFT JOIN (
			SELECT user_id, COUNT(*) AS review_count
			FROM Review
			GROUP BY user_id
		) rev ON u.user_id = rev.user_id
		LEFT JOIN (
			SELECT user_id, COUNT(*) AS rating_count
			FROM Rating
			GROUP BY user_id
		) rat ON u.user_id = rat.user_id
		ORDER BY watch_count DESC, rating_count DESC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []UserActivity
	for rows.Next() {
		var ua UserActivity
		err := rows.Scan(&ua.Username, &ua.WatchCount, &ua.TotalRewatches,
			&ua.ReviewCount, &ua.RatingCount)
		if err != nil {
			return nil, err
		}
		results = append(results, ua)
	}
	return results, nil
}

// ==========================
// HTTP Handlers
// ==========================

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	loggedInUser := getCurrentUser(r)
	if loggedInUser == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := PageData{
		Connected:    true,
		DBInfo:       fmt.Sprintf("MySQL 8.x @ %s:%s / %s", dbHost, dbPort, dbName),
		ActiveTab:    r.URL.Query().Get("tab"),
		LoggedInUser: loggedInUser,
		IsAdmin:      loggedInUser.IsAdmin,
	}
	if data.ActiveTab == "" {
		data.ActiveTab = "dashboard"
	}

	// Flash message from query params
	if msg := r.URL.Query().Get("msg"); msg != "" {
		msgType := r.URL.Query().Get("msg_type")
		if msgType == "" {
			msgType = "success"
		}
		data.Flash = &FlashMessage{Type: msgType, Message: msg}
	}

	// Admins get the full user list for user selector and manage tab
	if loggedInUser.IsAdmin {
		users, err := getUsers()
		if err != nil {
			data.Error = "Error fetching users: " + err.Error()
		}
		data.Users = users
	}

	movies, err := getMovies()
	if err != nil {
		data.Error = "Error fetching movies: " + err.Error()
	}
	data.Movies = movies

	genres, err := getGenres()
	if err != nil {
		data.Error = "Error fetching genres: " + err.Error()
	}
	data.Genres = genres

	// Determine which user's dashboard to show
	// Admins can select any user; regular users always see their own
	selectedUserID := loggedInUser.UserID
	if loggedInUser.IsAdmin {
		if uidStr := r.URL.Query().Get("user_id"); uidStr != "" {
			if uid, err := strconv.Atoi(uidStr); err == nil && uid > 0 {
				selectedUserID = uid
			}
		}
	}

	if loggedInUser.IsAdmin && len(data.Users) > 0 {
		for i, u := range data.Users {
			if u.UserID == selectedUserID {
				data.SelectedUser = &data.Users[i]
				break
			}
		}
	} else {
		data.SelectedUser = loggedInUser
	}

	if data.SelectedUser != nil {
		data.ViewingHistory, _ = getUserViewingHistory(data.SelectedUser.UserID)
		data.Recommendations, _ = getUserRecommendations(data.SelectedUser.UserID)
		data.Ratings, _ = getUserRatings(data.SelectedUser.UserID)
		data.Reviews, _ = getUserReviews(data.SelectedUser.UserID)
	}

	// Load reports data when on reports tab
	if data.ActiveTab == "reports" {
		data.PopularMovies, _ = getPopularMovies()
		data.UserActivities, _ = getUserActivitySummary()
	}

	tmpl, err := template.New("index").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).Parse(indexHTML)
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
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ?", dbName).Scan(&count)
	if err == nil {
		status["table_count"] = count
	}
	json.NewEncoder(w).Encode(status)
}

// ============================
// INSERT Handlers
// ============================

func insertMovieHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	title := strings.TrimSpace(r.FormValue("title"))
	plotSummary := strings.TrimSpace(r.FormValue("plot_summary"))
	trailerURL := strings.TrimSpace(r.FormValue("trailer_url"))
	tmdbID := strings.TrimSpace(r.FormValue("tmdb_id"))
	genreIDs := r.Form["genre_ids"]

	if title == "" {
		http.Redirect(w, r, "/?tab=manage&msg=Movie+title+is+required&msg_type=error", http.StatusSeeOther)
		return
	}

	var tmdbVal interface{}
	if tmdbID == "" {
		tmdbVal = nil
	} else {
		tmdbVal = tmdbID
	}

	result, err := db.Exec(
		"INSERT INTO Movie (title, plot_summary, trailer_url, tmdb_id) VALUES (?, ?, ?, ?)",
		title, plotSummary, trailerURL, tmdbVal,
	)
	if err != nil {
		http.Redirect(w, r, "/?tab=manage&msg=Error+inserting+movie:+"+template.URLQueryEscaper(err.Error())+"&msg_type=error", http.StatusSeeOther)
		return
	}

	movieID, _ := result.LastInsertId()

	for _, gidStr := range genreIDs {
		gid, err := strconv.Atoi(gidStr)
		if err == nil {
			db.Exec("INSERT INTO Movie_Genre (movie_id, genre_id) VALUES (?, ?)", movieID, gid)
		}
	}

	http.Redirect(w, r, "/?tab=manage&msg=Movie+%22"+template.URLQueryEscaper(title)+"%22+added+successfully", http.StatusSeeOther)
}

func insertUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if username == "" || email == "" || password == "" {
		http.Redirect(w, r, "/?tab=manage&msg=Username,+email,+and+password+are+required&msg_type=error", http.StatusSeeOther)
		return
	}
	if len(password) < 6 {
		http.Redirect(w, r, "/?tab=manage&msg=Password+must+be+at+least+6+characters&msg_type=error", http.StatusSeeOther)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	_, err := db.Exec("INSERT INTO User (username, email, password_hash) VALUES (?, ?, ?)",
		username, email, string(hash))
	if err != nil {
		http.Redirect(w, r, "/?tab=manage&msg=Error:+"+template.URLQueryEscaper(err.Error())+"&msg_type=error", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/?tab=manage&msg=User+%22"+template.URLQueryEscaper(username)+"%22+registered+successfully", http.StatusSeeOther)
}

func insertRatingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	userID, _ := strconv.Atoi(r.FormValue("user_id"))
	movieID, _ := strconv.Atoi(r.FormValue("movie_id"))
	ratingValue := r.FormValue("rating_value")

	if userID == 0 || movieID == 0 || ratingValue == "" {
		http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=All+fields+are+required&msg_type=error", userID), http.StatusSeeOther)
		return
	}

	_, err := db.Exec("INSERT INTO Rating (user_id, movie_id, rating_value) VALUES (?, ?, ?)", userID, movieID, ratingValue)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=Rating+already+exists+for+this+movie.+Use+update+instead.&msg_type=error", userID), http.StatusSeeOther)
		} else {
			http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=Error:+%s&msg_type=error", userID, template.URLQueryEscaper(err.Error())), http.StatusSeeOther)
		}
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=Rating+added+successfully", userID), http.StatusSeeOther)
}

func insertReviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	userID, _ := strconv.Atoi(r.FormValue("user_id"))
	movieID, _ := strconv.Atoi(r.FormValue("movie_id"))
	reviewText := strings.TrimSpace(r.FormValue("review_text"))
	isSpoiler := r.FormValue("is_spoiler") == "1"

	if userID == 0 || movieID == 0 || reviewText == "" {
		http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=All+fields+are+required&msg_type=error", userID), http.StatusSeeOther)
		return
	}

	_, err := db.Exec("INSERT INTO Review (user_id, movie_id, review_text, is_spoiler) VALUES (?, ?, ?, ?)",
		userID, movieID, reviewText, isSpoiler)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=Error:+%s&msg_type=error", userID, template.URLQueryEscaper(err.Error())), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=Review+added+successfully", userID), http.StatusSeeOther)
}

// ============================
// DELETE Handlers
// ============================

func deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	movieID, _ := strconv.Atoi(r.FormValue("movie_id"))
	if movieID == 0 {
		http.Redirect(w, r, "/?tab=manage&msg=Invalid+movie+ID&msg_type=error", http.StatusSeeOther)
		return
	}

	_, err := db.Exec("DELETE FROM Movie WHERE movie_id = ?", movieID)
	if err != nil {
		http.Redirect(w, r, "/?tab=manage&msg=Error:+"+template.URLQueryEscaper(err.Error())+"&msg_type=error", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/?tab=manage&msg=Movie+deleted+successfully", http.StatusSeeOther)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	userID, _ := strconv.Atoi(r.FormValue("user_id"))
	if userID == 0 {
		http.Redirect(w, r, "/?tab=manage&msg=Invalid+user+ID&msg_type=error", http.StatusSeeOther)
		return
	}

	_, err := db.Exec("DELETE FROM User WHERE user_id = ?", userID)
	if err != nil {
		http.Redirect(w, r, "/?tab=manage&msg=Error:+"+template.URLQueryEscaper(err.Error())+"&msg_type=error", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/?tab=manage&msg=User+deleted+successfully", http.StatusSeeOther)
}

func deleteRatingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	ratingID, _ := strconv.Atoi(r.FormValue("rating_id"))
	userID, _ := strconv.Atoi(r.FormValue("user_id"))

	_, err := db.Exec("DELETE FROM Rating WHERE rating_id = ?", ratingID)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=Error:+%s&msg_type=error", userID, template.URLQueryEscaper(err.Error())), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=Rating+deleted+successfully", userID), http.StatusSeeOther)
}

func deleteReviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	reviewID, _ := strconv.Atoi(r.FormValue("review_id"))
	userID, _ := strconv.Atoi(r.FormValue("user_id"))

	_, err := db.Exec("DELETE FROM Review WHERE review_id = ?", reviewID)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=Error:+%s&msg_type=error", userID, template.URLQueryEscaper(err.Error())), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=Review+deleted+successfully", userID), http.StatusSeeOther)
}

// ============================
// UPDATE Handlers
// ============================

func updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	movieID, _ := strconv.Atoi(r.FormValue("movie_id"))
	title := strings.TrimSpace(r.FormValue("title"))
	plotSummary := strings.TrimSpace(r.FormValue("plot_summary"))
	trailerURL := strings.TrimSpace(r.FormValue("trailer_url"))

	if movieID == 0 || title == "" {
		http.Redirect(w, r, "/?tab=manage&msg=Movie+ID+and+title+are+required&msg_type=error", http.StatusSeeOther)
		return
	}

	_, err := db.Exec(
		"UPDATE Movie SET title = ?, plot_summary = ?, trailer_url = ? WHERE movie_id = ?",
		title, plotSummary, trailerURL, movieID,
	)
	if err != nil {
		http.Redirect(w, r, "/?tab=manage&msg=Error:+"+template.URLQueryEscaper(err.Error())+"&msg_type=error", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/?tab=manage&msg=Movie+%22"+template.URLQueryEscaper(title)+"%22+updated+successfully", http.StatusSeeOther)
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	userID, _ := strconv.Atoi(r.FormValue("user_id"))
	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))

	if userID == 0 || username == "" || email == "" {
		http.Redirect(w, r, "/?tab=manage&msg=All+fields+are+required&msg_type=error", http.StatusSeeOther)
		return
	}

	_, err := db.Exec("UPDATE User SET username = ?, email = ? WHERE user_id = ?", username, email, userID)
	if err != nil {
		http.Redirect(w, r, "/?tab=manage&msg=Error:+"+template.URLQueryEscaper(err.Error())+"&msg_type=error", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/?tab=manage&msg=User+updated+successfully", http.StatusSeeOther)
}

func updateRatingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	ratingID, _ := strconv.Atoi(r.FormValue("rating_id"))
	userID, _ := strconv.Atoi(r.FormValue("user_id"))
	ratingValue := r.FormValue("rating_value")

	_, err := db.Exec("UPDATE Rating SET rating_value = ? WHERE rating_id = ?", ratingValue, ratingID)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=Error:+%s&msg_type=error", userID, template.URLQueryEscaper(err.Error())), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=Rating+updated+successfully", userID), http.StatusSeeOther)
}

func updateReviewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	r.ParseForm()
	reviewID, _ := strconv.Atoi(r.FormValue("review_id"))
	userID, _ := strconv.Atoi(r.FormValue("user_id"))
	reviewText := strings.TrimSpace(r.FormValue("review_text"))
	isSpoiler := r.FormValue("is_spoiler") == "1"

	_, err := db.Exec("UPDATE Review SET review_text = ?, is_spoiler = ? WHERE review_id = ?",
		reviewText, isSpoiler, reviewID)
	if err != nil {
		http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=Error:+%s&msg_type=error", userID, template.URLQueryEscaper(err.Error())), http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/?user_id=%d&tab=dashboard&msg=Review+updated+successfully", userID), http.StatusSeeOther)
}

// ==========================
// Auth Handlers
// ==========================

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if getCurrentUser(r) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		var flash *FlashMessage
		if msg := r.URL.Query().Get("msg"); msg != "" {
			msgType := r.URL.Query().Get("msg_type")
			if msgType == "" {
				msgType = "error"
			}
			flash = &FlashMessage{Type: msgType, Message: msg}
		}
		tmpl, _ := template.New("login").Parse(loginHTML)
		tmpl.Execute(w, map[string]interface{}{"Flash": flash})
		return
	}

	r.ParseForm()
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")

	var user User
	var createdAt time.Time
	err := db.QueryRow(
		"SELECT user_id, username, email, password_hash, is_admin, created_at FROM User WHERE username = ?",
		username,
	).Scan(&user.UserID, &user.Username, &user.Email, &user.PasswordHash, &user.IsAdmin, &createdAt)
	if err != nil {
		http.Redirect(w, r, "/login?msg=Invalid+username+or+password&msg_type=error", http.StatusSeeOther)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/login?msg=Invalid+username+or+password&msg_type=error", http.StatusSeeOther)
		return
	}

	db.Exec("UPDATE User SET last_login = NOW() WHERE user_id = ?", user.UserID)

	session, _ := store.Get(r, "weekly-watch-session")
	session.Values["user_id"] = user.UserID
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if getCurrentUser(r) != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		var flash *FlashMessage
		if msg := r.URL.Query().Get("msg"); msg != "" {
			msgType := r.URL.Query().Get("msg_type")
			if msgType == "" {
				msgType = "error"
			}
			flash = &FlashMessage{Type: msgType, Message: msg}
		}
		tmpl, _ := template.New("signup").Parse(signupHTML)
		tmpl.Execute(w, map[string]interface{}{"Flash": flash})
		return
	}

	r.ParseForm()
	username := strings.TrimSpace(r.FormValue("username"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if username == "" || email == "" || password == "" {
		http.Redirect(w, r, "/signup?msg=All+fields+are+required&msg_type=error", http.StatusSeeOther)
		return
	}
	if len(password) < 6 {
		http.Redirect(w, r, "/signup?msg=Password+must+be+at+least+6+characters&msg_type=error", http.StatusSeeOther)
		return
	}
	if password != confirmPassword {
		http.Redirect(w, r, "/signup?msg=Passwords+do+not+match&msg_type=error", http.StatusSeeOther)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Redirect(w, r, "/signup?msg=Error+creating+account&msg_type=error", http.StatusSeeOther)
		return
	}

	_, err = db.Exec("INSERT INTO User (username, email, password_hash) VALUES (?, ?, ?)",
		username, email, string(hash))
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			http.Redirect(w, r, "/signup?msg=Username+or+email+already+exists&msg_type=error", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/signup?msg=Error:+"+template.URLQueryEscaper(err.Error())+"&msg_type=error", http.StatusSeeOther)
		}
		return
	}

	http.Redirect(w, r, "/login?msg=Account+created+successfully.+Please+log+in.&msg_type=success", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "weekly-watch-session")
	session.Values["user_id"] = nil
	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/login?msg=Logged+out+successfully&msg_type=success", http.StatusSeeOther)
}

func changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	user := getCurrentUser(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/?tab=dashboard", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	currentPassword := r.FormValue("current_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
	if err != nil {
		http.Redirect(w, r, "/?tab=dashboard&msg=Current+password+is+incorrect&msg_type=error", http.StatusSeeOther)
		return
	}

	if len(newPassword) < 6 {
		http.Redirect(w, r, "/?tab=dashboard&msg=New+password+must+be+at+least+6+characters&msg_type=error", http.StatusSeeOther)
		return
	}
	if newPassword != confirmPassword {
		http.Redirect(w, r, "/?tab=dashboard&msg=New+passwords+do+not+match&msg_type=error", http.StatusSeeOther)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	_, err = db.Exec("UPDATE User SET password_hash = ? WHERE user_id = ?", string(hash), user.UserID)
	if err != nil {
		http.Redirect(w, r, "/?tab=dashboard&msg=Error+updating+password&msg_type=error", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/?tab=dashboard&msg=Password+changed+successfully&msg_type=success", http.StatusSeeOther)
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
        .container { max-width: 1200px; margin: 0 auto; padding: 32px 24px; }

        .db-info {
            background: #1a1a2e; border: 1px solid #2a2a4a; border-radius: 8px;
            padding: 16px 24px; margin-bottom: 24px; font-size: 14px; color: #8888aa;
        }
        .db-info strong { color: #e94560; }

        .flash {
            padding: 14px 20px; border-radius: 8px; margin-bottom: 24px;
            font-size: 14px; font-weight: 500;
        }
        .flash-success { background: #1b4332; color: #95d5b2; border: 1px solid #2d6a4f; }
        .flash-error { background: #4a1520; color: #f4978e; border: 1px solid #6b2030; }

        .tab-nav {
            display: flex; gap: 4px; margin-bottom: 32px;
            border-bottom: 2px solid #2a2a4a; padding-bottom: 0;
        }
        .tab-link {
            background: transparent; color: #8888aa; border: none;
            padding: 12px 24px; font-size: 14px; font-weight: 600;
            cursor: pointer; border-bottom: 2px solid transparent;
            margin-bottom: -2px; transition: all 0.2s;
            text-decoration: none;
        }
        .tab-link:hover { color: #e0e0e0; }
        .tab-link.active { color: #e94560; border-bottom-color: #e94560; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }

        .section { margin-bottom: 40px; }
        .section h2 {
            color: #e94560; font-size: 20px; margin-bottom: 16px;
            padding-bottom: 8px; border-bottom: 1px solid #2a2a4a;
        }
        .section h3 { color: #c0c0d0; font-size: 16px; margin-bottom: 12px; }

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

        .user-select { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 24px; }
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
            text-align: center; padding: 32px; color: #555; font-style: italic;
        }

        .grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 24px; }
        @media (max-width: 768px) { .grid-2 { grid-template-columns: 1fr; } }

        .form-card {
            background: #1a1a2e; border: 1px solid #2a2a4a; border-radius: 8px;
            padding: 24px; margin-bottom: 24px;
        }
        .form-card h3 {
            color: #e94560; font-size: 16px; margin-bottom: 16px;
            padding-bottom: 8px; border-bottom: 1px solid #2a2a4a;
        }
        .form-row { display: flex; gap: 16px; margin-bottom: 12px; flex-wrap: wrap; }
        .form-group { display: flex; flex-direction: column; flex: 1; min-width: 200px; }
        .form-group label {
            font-size: 12px; color: #8888aa; margin-bottom: 4px;
            text-transform: uppercase; letter-spacing: 0.5px;
        }
        .form-group input, .form-group select, .form-group textarea {
            background: #0f0f0f; border: 1px solid #3a3a5a; color: #e0e0e0;
            padding: 10px 12px; border-radius: 6px; font-size: 14px; font-family: inherit;
        }
        .form-group input:focus, .form-group select:focus, .form-group textarea:focus {
            outline: none; border-color: #e94560;
        }
        .form-group textarea { resize: vertical; min-height: 60px; }
        .checkbox-group { display: flex; align-items: center; gap: 8px; padding-top: 8px; }
        .checkbox-group input[type="checkbox"] { width: 16px; height: 16px; accent-color: #e94560; }

        .btn {
            padding: 10px 20px; border: none; border-radius: 6px; cursor: pointer;
            font-size: 14px; font-weight: 600; transition: all 0.2s;
        }
        .btn-primary { background: #e94560; color: white; }
        .btn-primary:hover { background: #d63851; }
        .btn-small { padding: 5px 12px; font-size: 12px; }
        .btn-inline {
            display: inline-block; padding: 4px 10px; font-size: 11px;
            border: none; border-radius: 4px; cursor: pointer; font-weight: 600;
        }
        .btn-inline.edit { background: #1b3a5c; color: #7eb8e0; }
        .btn-inline.delete { background: #4a1520; color: #f4978e; }
        .btn-inline.edit:hover { background: #254d73; }
        .btn-inline.delete:hover { background: #6b2030; }

        .actions-cell { white-space: nowrap; }
        .actions-cell form { display: inline; }

        .edit-row { display: none; }
        .edit-row.visible { display: table-row; }
        .edit-row td { background: #16213e; padding: 16px; }
        .edit-form-inline {
            display: flex; gap: 12px; align-items: center; flex-wrap: wrap;
        }
        .edit-form-inline input, .edit-form-inline select, .edit-form-inline textarea {
            background: #0f0f0f; border: 1px solid #3a3a5a; color: #e0e0e0;
            padding: 6px 10px; border-radius: 4px; font-size: 13px; font-family: inherit;
        }
        .genre-checkboxes { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 4px; }
        .genre-checkboxes label {
            display: flex; align-items: center; gap: 4px;
            font-size: 13px; color: #c0c0d0; cursor: pointer;
        }
        .genre-checkboxes input[type="checkbox"] { accent-color: #e94560; }
    </style>
</head>
<body>
    <div class="header">
        <h1>The Weekly Watch</h1>
        <div style="display:flex;align-items:center;gap:16px;">
            {{if .Connected}}
                <span class="connection-badge">Connected to MySQL</span>
            {{end}}
            {{if .LoggedInUser}}
                <span style="color:#c0c0d0;font-size:14px;">
                    {{.LoggedInUser.Username}}
                    {{if .IsAdmin}}<span style="color:#ffd166;font-size:11px;margin-left:4px;">(admin)</span>{{end}}
                </span>
                <a href="/logout" style="color:#e94560;font-size:13px;text-decoration:none;font-weight:600;">Logout</a>
            {{end}}
        </div>
    </div>

    <div class="container">
        {{if .Error}}
            <div class="flash flash-error">{{.Error}}</div>
        {{end}}

        {{if .Flash}}
            <div class="flash flash-{{.Flash.Type}}">{{.Flash.Message}}</div>
        {{end}}

        <div class="db-info">
            <strong>DBMS:</strong> {{.DBInfo}} &nbsp;|&nbsp;
            <strong>Language:</strong> Go (Golang) &nbsp;|&nbsp;
            <strong>Driver:</strong> go-sql-driver/mysql &nbsp;|&nbsp;
            <strong>Status:</strong> Successfully Connected
        </div>

        <!-- USER SELECTOR (admin only) -->
        {{if .IsAdmin}}
        <div class="section">
            <h2>Select a User</h2>
            <div class="user-select">
                {{range .Users}}
                    <a class="user-btn {{if $.SelectedUser}}{{if eq $.SelectedUser.UserID .UserID}}active{{end}}{{end}}"
                       href="/?user_id={{.UserID}}&tab={{$.ActiveTab}}">
                        {{.Username}}
                    </a>
                {{end}}
            </div>
        </div>
        {{end}}

        <!-- TAB NAVIGATION -->
        <div class="tab-nav">
            <a class="tab-link {{if eq .ActiveTab "dashboard"}}active{{end}}"
               href="/?{{if .SelectedUser}}user_id={{.SelectedUser.UserID}}&{{end}}tab=dashboard">Dashboard</a>
            {{if .IsAdmin}}
            <a class="tab-link {{if eq .ActiveTab "manage"}}active{{end}}"
               href="/?{{if .SelectedUser}}user_id={{.SelectedUser.UserID}}&{{end}}tab=manage">Manage Records</a>
            {{end}}
            <a class="tab-link {{if eq .ActiveTab "browse"}}active{{end}}"
               href="/?{{if .SelectedUser}}user_id={{.SelectedUser.UserID}}&{{end}}tab=browse">Browse All</a>
            <a class="tab-link {{if eq .ActiveTab "reports"}}active{{end}}"
               href="/?tab=reports">Reports</a>
        </div>

        <!-- ==================== DASHBOARD TAB ==================== -->
        <div class="tab-content {{if eq .ActiveTab "dashboard"}}active{{end}}">
            {{if .SelectedUser}}
            <div class="section">
                <h2>Dashboard: {{.SelectedUser.Username}}</h2>

                <div class="grid-2">
                    <!-- Viewing History -->
                    <div>
                        <h3>Viewing History</h3>
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

                    <!-- Ratings with edit/delete -->
                    <div>
                        <h3>Ratings</h3>
                        {{if .Ratings}}
                        <table>
                            <tr><th>Movie</th><th>Rating</th><th>Date</th><th>Actions</th></tr>
                            {{range .Ratings}}
                            <tr>
                                <td>{{.MovieTitle}}</td>
                                <td><span class="status-badge status-{{.RatingValue}}">{{.RatingValue}}</span></td>
                                <td>{{.RatedAt}}</td>
                                <td class="actions-cell">
                                    <button class="btn-inline edit" onclick="toggleEdit('rating', {{.RatingID}})">Edit</button>
                                    <form action="/delete/rating" method="POST" onsubmit="return confirm('Delete this rating?')">
                                        <input type="hidden" name="rating_id" value="{{.RatingID}}">
                                        <input type="hidden" name="user_id" value="{{$.SelectedUser.UserID}}">
                                        <button type="submit" class="btn-inline delete">Delete</button>
                                    </form>
                                </td>
                            </tr>
                            <tr class="edit-row" id="edit-rating-{{.RatingID}}">
                                <td colspan="4">
                                    <form action="/update/rating" method="POST" class="edit-form-inline">
                                        <input type="hidden" name="rating_id" value="{{.RatingID}}">
                                        <input type="hidden" name="user_id" value="{{$.SelectedUser.UserID}}">
                                        <label style="color:#8888aa;font-size:12px;">New rating:</label>
                                        <select name="rating_value">
                                            <option value="loved" {{if eq .RatingValue "loved"}}selected{{end}}>Loved</option>
                                            <option value="liked" {{if eq .RatingValue "liked"}}selected{{end}}>Liked</option>
                                            <option value="disliked" {{if eq .RatingValue "disliked"}}selected{{end}}>Disliked</option>
                                        </select>
                                        <button type="submit" class="btn btn-primary btn-small">Save</button>
                                        <button type="button" class="btn btn-small" style="background:#2a2a4a;color:#aaa;" onclick="toggleEdit('rating', {{.RatingID}})">Cancel</button>
                                    </form>
                                </td>
                            </tr>
                            {{end}}
                        </table>
                        {{else}}<p class="empty-state">No ratings yet.</p>{{end}}

                        <div class="form-card" style="margin-top: 16px;">
                            <h3>Add Rating</h3>
                            <form action="/insert/rating" method="POST">
                                <input type="hidden" name="user_id" value="{{.SelectedUser.UserID}}">
                                <div class="form-row">
                                    <div class="form-group">
                                        <label>Movie</label>
                                        <select name="movie_id" required>
                                            <option value="">-- Select Movie --</option>
                                            {{range $.Movies}}
                                            <option value="{{.MovieID}}">{{.Title}}</option>
                                            {{end}}
                                        </select>
                                    </div>
                                    <div class="form-group">
                                        <label>Rating</label>
                                        <select name="rating_value" required>
                                            <option value="loved">Loved</option>
                                            <option value="liked">Liked</option>
                                            <option value="disliked">Disliked</option>
                                        </select>
                                    </div>
                                </div>
                                <button type="submit" class="btn btn-primary">Add Rating</button>
                            </form>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Reviews -->
            <div class="section">
                <h2>Reviews</h2>
                {{if .Reviews}}
                <table>
                    <tr><th>Movie</th><th>Review</th><th>Spoiler</th><th>Date</th><th>Actions</th></tr>
                    {{range .Reviews}}
                    <tr>
                        <td>{{.MovieTitle}}</td>
                        <td style="max-width:400px;">{{.ReviewText}}</td>
                        <td>{{if .IsSpoiler}}<span class="status-badge status-pending">Yes</span>{{else}}No{{end}}</td>
                        <td>{{.CreatedAt}}</td>
                        <td class="actions-cell">
                            <button class="btn-inline edit" onclick="toggleEdit('review', {{.ReviewID}})">Edit</button>
                            <form action="/delete/review" method="POST" onsubmit="return confirm('Delete this review?')">
                                <input type="hidden" name="review_id" value="{{.ReviewID}}">
                                <input type="hidden" name="user_id" value="{{$.SelectedUser.UserID}}">
                                <button type="submit" class="btn-inline delete">Delete</button>
                            </form>
                        </td>
                    </tr>
                    <tr class="edit-row" id="edit-review-{{.ReviewID}}">
                        <td colspan="5">
                            <form action="/update/review" method="POST" class="edit-form-inline">
                                <input type="hidden" name="review_id" value="{{.ReviewID}}">
                                <input type="hidden" name="user_id" value="{{$.SelectedUser.UserID}}">
                                <textarea name="review_text" style="width:400px;min-height:50px;">{{.ReviewText}}</textarea>
                                <label style="color:#c0c0d0;font-size:13px;"><input type="checkbox" name="is_spoiler" value="1" {{if .IsSpoiler}}checked{{end}} style="accent-color:#e94560;"> Spoiler</label>
                                <button type="submit" class="btn btn-primary btn-small">Save</button>
                                <button type="button" class="btn btn-small" style="background:#2a2a4a;color:#aaa;" onclick="toggleEdit('review', {{.ReviewID}})">Cancel</button>
                            </form>
                        </td>
                    </tr>
                    {{end}}
                </table>
                {{else}}<p class="empty-state">No reviews yet.</p>{{end}}

                <div class="form-card" style="margin-top: 16px;">
                    <h3>Write a Review</h3>
                    <form action="/insert/review" method="POST">
                        <input type="hidden" name="user_id" value="{{.SelectedUser.UserID}}">
                        <div class="form-row">
                            <div class="form-group">
                                <label>Movie</label>
                                <select name="movie_id" required>
                                    <option value="">-- Select Movie --</option>
                                    {{range $.Movies}}
                                    <option value="{{.MovieID}}">{{.Title}}</option>
                                    {{end}}
                                </select>
                            </div>
                        </div>
                        <div class="form-row">
                            <div class="form-group">
                                <label>Review Text</label>
                                <textarea name="review_text" required placeholder="Write your review..."></textarea>
                            </div>
                        </div>
                        <div class="checkbox-group" style="margin-bottom:12px;">
                            <input type="checkbox" name="is_spoiler" value="1" id="spoiler-check">
                            <label for="spoiler-check" style="font-size:13px;">Contains spoilers</label>
                        </div>
                        <button type="submit" class="btn btn-primary">Submit Review</button>
                    </form>
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

            <!-- Change Password (only for own account) -->
            {{if .LoggedInUser}}{{if eq .SelectedUser.UserID .LoggedInUser.UserID}}
            <div class="section">
                <h2>Change Password</h2>
                <div class="form-card">
                    <form action="/change-password" method="POST">
                        <div class="form-row">
                            <div class="form-group">
                                <label>Current Password</label>
                                <input type="password" name="current_password" required>
                            </div>
                            <div class="form-group">
                                <label>New Password</label>
                                <input type="password" name="new_password" required minlength="6">
                            </div>
                            <div class="form-group">
                                <label>Confirm New Password</label>
                                <input type="password" name="confirm_password" required minlength="6">
                            </div>
                        </div>
                        <button type="submit" class="btn btn-primary">Change Password</button>
                    </form>
                </div>
            </div>
            {{end}}{{end}}

            {{else}}
            <div class="empty-state" style="padding: 60px 20px;">
                <p style="font-size: 18px; color: #8888aa;">Select a user above to view their dashboard.</p>
            </div>
            {{end}}
        </div>

        <!-- ==================== MANAGE TAB (admin only) ==================== -->
        {{if .IsAdmin}}
        <div class="tab-content {{if eq .ActiveTab "manage"}}active{{end}}">

            <div class="form-card">
                <h3>Add New Movie</h3>
                <form action="/insert/movie" method="POST">
                    <div class="form-row">
                        <div class="form-group">
                            <label>Title *</label>
                            <input type="text" name="title" required placeholder="e.g. The Matrix">
                        </div>
                        <div class="form-group">
                            <label>TMDB ID</label>
                            <input type="text" name="tmdb_id" placeholder="e.g. 603">
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label>Plot Summary</label>
                            <textarea name="plot_summary" placeholder="Brief plot description..."></textarea>
                        </div>
                    </div>
                    <div class="form-row">
                        <div class="form-group">
                            <label>Trailer URL</label>
                            <input type="text" name="trailer_url" placeholder="https://youtube.com/watch?v=...">
                        </div>
                    </div>
                    <div class="form-group" style="margin-bottom: 12px;">
                        <label>Genres</label>
                        <div class="genre-checkboxes">
                            {{range .Genres}}
                            <label><input type="checkbox" name="genre_ids" value="{{.GenreID}}"> {{.GenreName}}</label>
                            {{end}}
                        </div>
                    </div>
                    <button type="submit" class="btn btn-primary">Add Movie</button>
                </form>
            </div>

            <div class="form-card">
                <h3>Register New User (Admin)</h3>
                <form action="/insert/user" method="POST">
                    <div class="form-row">
                        <div class="form-group">
                            <label>Username *</label>
                            <input type="text" name="username" required placeholder="e.g. john_d">
                        </div>
                        <div class="form-group">
                            <label>Email *</label>
                            <input type="email" name="email" required placeholder="e.g. john@example.com">
                        </div>
                        <div class="form-group">
                            <label>Password *</label>
                            <input type="password" name="password" required placeholder="Min 6 characters" minlength="6">
                        </div>
                    </div>
                    <button type="submit" class="btn btn-primary">Register User</button>
                </form>
            </div>

            <div class="section">
                <h2>All Movies (Edit / Delete)</h2>
                <table>
                    <tr><th>ID</th><th>Title</th><th>Genres</th><th>TMDB</th><th>Actions</th></tr>
                    {{range .Movies}}
                    <tr>
                        <td>{{.MovieID}}</td>
                        <td><strong>{{.Title}}</strong></td>
                        <td>{{.Genres}}</td>
                        <td>{{.TmdbID}}</td>
                        <td class="actions-cell">
                            <button class="btn-inline edit" onclick="toggleEdit('movie', {{.MovieID}})">Edit</button>
                            <form action="/delete/movie" method="POST" onsubmit="return confirm('Delete this movie and all associated data?')">
                                <input type="hidden" name="movie_id" value="{{.MovieID}}">
                                <button type="submit" class="btn-inline delete">Delete</button>
                            </form>
                        </td>
                    </tr>
                    <tr class="edit-row" id="edit-movie-{{.MovieID}}">
                        <td colspan="5">
                            <form action="/update/movie" method="POST" class="edit-form-inline">
                                <input type="hidden" name="movie_id" value="{{.MovieID}}">
                                <input type="text" name="title" value="{{.Title}}" placeholder="Title" style="width:200px;" required>
                                <input type="text" name="plot_summary" value="{{.PlotSummary}}" placeholder="Plot summary" style="width:250px;">
                                <input type="text" name="trailer_url" value="{{.TrailerURL}}" placeholder="Trailer URL" style="width:200px;">
                                <button type="submit" class="btn btn-primary btn-small">Save</button>
                                <button type="button" class="btn btn-small" style="background:#2a2a4a;color:#aaa;" onclick="toggleEdit('movie', {{.MovieID}})">Cancel</button>
                            </form>
                        </td>
                    </tr>
                    {{end}}
                </table>
            </div>

            <div class="section">
                <h2>All Users (Edit / Delete)</h2>
                <table>
                    <tr><th>ID</th><th>Username</th><th>Email</th><th>Joined</th><th>Actions</th></tr>
                    {{range .Users}}
                    <tr>
                        <td>{{.UserID}}</td>
                        <td>{{.Username}}</td>
                        <td>{{.Email}}</td>
                        <td>{{.CreatedAt}}</td>
                        <td class="actions-cell">
                            <button class="btn-inline edit" onclick="toggleEdit('user', {{.UserID}})">Edit</button>
                            <form action="/delete/user" method="POST" onsubmit="return confirm('Delete this user and all their data?')">
                                <input type="hidden" name="user_id" value="{{.UserID}}">
                                <button type="submit" class="btn-inline delete">Delete</button>
                            </form>
                        </td>
                    </tr>
                    <tr class="edit-row" id="edit-user-{{.UserID}}">
                        <td colspan="5">
                            <form action="/update/user" method="POST" class="edit-form-inline">
                                <input type="hidden" name="user_id" value="{{.UserID}}">
                                <input type="text" name="username" value="{{.Username}}" placeholder="Username" required>
                                <input type="email" name="email" value="{{.Email}}" placeholder="Email" required>
                                <button type="submit" class="btn btn-primary btn-small">Save</button>
                                <button type="button" class="btn btn-small" style="background:#2a2a4a;color:#aaa;" onclick="toggleEdit('user', {{.UserID}})">Cancel</button>
                            </form>
                        </td>
                    </tr>
                    {{end}}
                </table>
            </div>
        </div>
        {{end}}

        <!-- ==================== BROWSE TAB ==================== -->
        <div class="tab-content {{if eq .ActiveTab "browse"}}active{{end}}">
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

            {{if .IsAdmin}}
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
            {{end}}
        </div>

        <!-- ==================== REPORTS TAB ==================== -->
        <div class="tab-content {{if eq .ActiveTab "reports"}}active{{end}}">
            <div class="section">
                <h2>Most Popular Movies</h2>
                <p style="color:#8888aa;font-size:13px;margin-bottom:12px;">
                    Movies ranked by total ratings received (uses JOIN on Rating + Movie, GROUP BY movie, COUNT/SUM aggregates).
                </p>
                {{if .PopularMovies}}
                <table>
                    <tr><th>#</th><th>Movie</th><th>Total Ratings</th><th>Loved</th><th>Liked</th></tr>
                    {{range $i, $m := .PopularMovies}}
                    <tr>
                        <td>{{add $i 1}}</td>
                        <td><strong>{{$m.Title}}</strong></td>
                        <td>{{$m.RatingCount}}</td>
                        <td><span class="status-badge status-loved">{{$m.LovedCount}}</span></td>
                        <td><span class="status-badge status-liked">{{$m.LikedCount}}</span></td>
                    </tr>
                    {{end}}
                </table>
                {{else}}<p class="empty-state">No rating data available.</p>{{end}}
            </div>

            <div class="section">
                <h2>User Activity Summary</h2>
                <p style="color:#8888aa;font-size:13px;margin-bottom:12px;">
                    Activity breakdown per user (uses LEFT JOIN with subqueries on Viewing_History, Review, and Rating tables with GROUP BY and COUNT/SUM aggregates).
                </p>
                {{if .UserActivities}}
                <table>
                    <tr><th>User</th><th>Movies Watched</th><th>Total Views (incl. rewatches)</th><th>Reviews Written</th><th>Ratings Given</th></tr>
                    {{range .UserActivities}}
                    <tr>
                        <td>{{.Username}}</td>
                        <td>{{.WatchCount}}</td>
                        <td>{{.TotalRewatches}}</td>
                        <td>{{.ReviewCount}}</td>
                        <td>{{.RatingCount}}</td>
                    </tr>
                    {{end}}
                </table>
                {{else}}<p class="empty-state">No activity data available.</p>{{end}}
            </div>
        </div>
    </div>

    <script>
        function toggleEdit(type, id) {
            var row = document.getElementById('edit-' + type + '-' + id);
            if (row) {
                row.classList.toggle('visible');
            }
        }
    </script>
</body>
</html>`

// ==========================
// Login Template
// ==========================

const loginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Login - The Weekly Watch</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f0f0f; color: #e0e0e0; min-height: 100vh;
            display: flex; justify-content: center; align-items: center;
        }
        .login-card {
            background: #1a1a2e; border: 1px solid #2a2a4a; border-radius: 12px;
            padding: 40px; width: 100%; max-width: 420px;
        }
        .login-card h1 { color: #e94560; text-align: center; margin-bottom: 8px; font-size: 28px; }
        .login-card .subtitle { color: #8888aa; text-align: center; margin-bottom: 24px; font-size: 14px; }
        .form-group { margin-bottom: 16px; }
        .form-group label {
            display: block; font-size: 12px; color: #8888aa; margin-bottom: 4px;
            text-transform: uppercase; letter-spacing: 0.5px;
        }
        .form-group input {
            width: 100%; background: #0f0f0f; border: 1px solid #3a3a5a;
            color: #e0e0e0; padding: 10px 12px; border-radius: 6px; font-size: 14px;
        }
        .form-group input:focus { outline: none; border-color: #e94560; }
        .btn-primary {
            width: 100%; padding: 12px; background: #e94560; color: white;
            border: none; border-radius: 6px; font-size: 14px; font-weight: 600; cursor: pointer;
            margin-top: 8px;
        }
        .btn-primary:hover { background: #d63851; }
        .flash { padding: 12px 16px; border-radius: 8px; margin-bottom: 20px; font-size: 14px; }
        .flash-error { background: #4a1520; color: #f4978e; border: 1px solid #6b2030; }
        .flash-success { background: #1b4332; color: #95d5b2; border: 1px solid #2d6a4f; }
        .link-row { text-align: center; margin-top: 16px; font-size: 14px; color: #8888aa; }
        .link-row a { color: #e94560; text-decoration: none; }
        .link-row a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="login-card">
        <h1>The Weekly Watch</h1>
        <p class="subtitle">Sign in to your account</p>
        {{if .Flash}}
            <div class="flash flash-{{.Flash.Type}}">{{.Flash.Message}}</div>
        {{end}}
        <form method="POST" action="/login">
            <div class="form-group">
                <label>Username</label>
                <input type="text" name="username" required autofocus placeholder="Enter your username">
            </div>
            <div class="form-group">
                <label>Password</label>
                <input type="password" name="password" required placeholder="Enter your password">
            </div>
            <button type="submit" class="btn-primary">Sign In</button>
        </form>
        <div class="link-row">
            Don&#39;t have an account? <a href="/signup">Sign up</a>
        </div>
    </div>
</body>
</html>`

// ==========================
// Signup Template
// ==========================

const signupHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sign Up - The Weekly Watch</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #0f0f0f; color: #e0e0e0; min-height: 100vh;
            display: flex; justify-content: center; align-items: center;
        }
        .signup-card {
            background: #1a1a2e; border: 1px solid #2a2a4a; border-radius: 12px;
            padding: 40px; width: 100%; max-width: 420px;
        }
        .signup-card h1 { color: #e94560; text-align: center; margin-bottom: 8px; font-size: 28px; }
        .signup-card .subtitle { color: #8888aa; text-align: center; margin-bottom: 24px; font-size: 14px; }
        .form-group { margin-bottom: 16px; }
        .form-group label {
            display: block; font-size: 12px; color: #8888aa; margin-bottom: 4px;
            text-transform: uppercase; letter-spacing: 0.5px;
        }
        .form-group input {
            width: 100%; background: #0f0f0f; border: 1px solid #3a3a5a;
            color: #e0e0e0; padding: 10px 12px; border-radius: 6px; font-size: 14px;
        }
        .form-group input:focus { outline: none; border-color: #e94560; }
        .btn-primary {
            width: 100%; padding: 12px; background: #e94560; color: white;
            border: none; border-radius: 6px; font-size: 14px; font-weight: 600; cursor: pointer;
            margin-top: 8px;
        }
        .btn-primary:hover { background: #d63851; }
        .flash { padding: 12px 16px; border-radius: 8px; margin-bottom: 20px; font-size: 14px; }
        .flash-error { background: #4a1520; color: #f4978e; border: 1px solid #6b2030; }
        .flash-success { background: #1b4332; color: #95d5b2; border: 1px solid #2d6a4f; }
        .link-row { text-align: center; margin-top: 16px; font-size: 14px; color: #8888aa; }
        .link-row a { color: #e94560; text-decoration: none; }
        .link-row a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="signup-card">
        <h1>The Weekly Watch</h1>
        <p class="subtitle">Create a new account</p>
        {{if .Flash}}
            <div class="flash flash-{{.Flash.Type}}">{{.Flash.Message}}</div>
        {{end}}
        <form method="POST" action="/signup">
            <div class="form-group">
                <label>Username</label>
                <input type="text" name="username" required autofocus placeholder="Choose a username">
            </div>
            <div class="form-group">
                <label>Email</label>
                <input type="email" name="email" required placeholder="Enter your email">
            </div>
            <div class="form-group">
                <label>Password</label>
                <input type="password" name="password" required placeholder="Min 6 characters" minlength="6">
            </div>
            <div class="form-group">
                <label>Confirm Password</label>
                <input type="password" name="confirm_password" required placeholder="Confirm your password" minlength="6">
            </div>
            <button type="submit" class="btn-primary">Create Account</button>
        </form>
        <div class="link-row">
            Already have an account? <a href="/login">Sign in</a>
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
	fmt.Println("  CS 4604 Phase 6                         ")
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

	// Public routes (no auth required)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/api/status", apiStatusHandler)

	// Auth-required routes
	http.HandleFunc("/", requireAuth(homeHandler))
	http.HandleFunc("/change-password", requireAuth(changePasswordHandler))

	// User data routes (auth required)
	http.HandleFunc("/insert/rating", requireAuth(insertRatingHandler))
	http.HandleFunc("/insert/review", requireAuth(insertReviewHandler))
	http.HandleFunc("/delete/rating", requireAuth(deleteRatingHandler))
	http.HandleFunc("/delete/review", requireAuth(deleteReviewHandler))
	http.HandleFunc("/update/rating", requireAuth(updateRatingHandler))
	http.HandleFunc("/update/review", requireAuth(updateReviewHandler))

	// Admin-only routes
	http.HandleFunc("/insert/movie", requireAdmin(insertMovieHandler))
	http.HandleFunc("/insert/user", requireAdmin(insertUserHandler))
	http.HandleFunc("/delete/movie", requireAdmin(deleteMovieHandler))
	http.HandleFunc("/delete/user", requireAdmin(deleteUserHandler))
	http.HandleFunc("/update/movie", requireAdmin(updateMovieHandler))
	http.HandleFunc("/update/user", requireAdmin(updateUserHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
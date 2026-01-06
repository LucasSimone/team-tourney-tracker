package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Player struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PlayerStat struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	GamesPlayed   int     `json:"games_played"`
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	WinPercentage float64 `json:"win_percentage"`
}

type Season struct {
	ID   int `json:"id"`
	Year int `json:"year"`
}

type Match struct {
	ID        int    `json:"id"`
	SeasonID  int    `json:"season_id"`
	TeamAID   int    `json:"team_a_id"`
	TeamBID   int    `json:"team_b_id"`
	ScoreA    int    `json:"score_a"`
	ScoreB    int    `json:"score_b"`
	WinnerID  int    `json:"winner_id"`
	CreatedAt string `json:"created_at"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

var db *sql.DB

func getCORSOrigin() string {
	// Get allowed origin from environment, default to localhost for dev
	origin := os.Getenv("CORS_ORIGIN")
	if origin == "" {
		origin = "http://localhost:5173"
	}
	return origin
}

func setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	allowedOrigin := getCORSOrigin()

	// Check if origin matches allowed origin
	if origin == allowedOrigin {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
}

func getIDFromPath(path string) (string, bool) {
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")

	// We need at least 3 parts: ["", "resource", "id"]
	// e.g. "/players/1" splits to ["", "players", "1"]
	if len(parts) < 3 {
		return "", false
	}

	lastPart := parts[len(parts)-1]
	// Check if last part is not empty and not a special word
	if lastPart != "" && lastPart != "role" {
		return lastPart, true
	}
	return "", false
}

func main() {
	var err error
	os.MkdirAll("/db", 0755)
	db, err = sql.Open("sqlite3", "/db/sports.db")
	if err != nil {
		log.Fatal(err)
	}
	initDB()
	initJWTSecret()

	// Register all handlers - use a catch-all pattern and handle paths in each handler
	http.HandleFunc("/teams/", teamsHandler)
	http.HandleFunc("/teams", teamsHandler)
	http.HandleFunc("/seasons/", seasonsHandler)
	http.HandleFunc("/seasons", seasonsHandler)
	http.HandleFunc("/matches/", matchesHandler)
	http.HandleFunc("/matches", matchesHandler)
	http.HandleFunc("/players/stats", playerStatsHandler)
	http.HandleFunc("/players/", playersHandler)
	http.HandleFunc("/players", playersHandler)
	http.HandleFunc("/auth/login", loginHandler)
	http.HandleFunc("/auth/register", registerHandler)
	http.HandleFunc("/auth/users/", usersHandler)
	http.HandleFunc("/auth/users", usersHandler)
	http.HandleFunc("/admin/backup", backupHandler)
	http.HandleFunc("/admin/backup/download", backupDownloadHandler)

	// Setup weekly database backups
	backupEmail := os.Getenv("BACKUP_EMAIL")
	if backupEmail != "" {
		setupWeeklyBackup(backupEmail)
	} else {
		log.Println("Weekly email backups disabled (BACKUP_EMAIL not set), using local backup system")
	}

	log.Println("Backend running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS teams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);
	CREATE TABLE IF NOT EXISTS seasons (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL UNIQUE
	);
	CREATE TABLE IF NOT EXISTS matches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		season_id INTEGER,
		team_a_id INTEGER,
		team_b_id INTEGER,
		score_a INTEGER,
		score_b INTEGER,
		winner_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(season_id) REFERENCES seasons(id),
		FOREIGN KEY(team_a_id) REFERENCES teams(id),
		FOREIGN KEY(team_b_id) REFERENCES teams(id),
		FOREIGN KEY(winner_id) REFERENCES teams(id)
	);
	CREATE TABLE IF NOT EXISTS players (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);
	CREATE TABLE IF NOT EXISTS team_players (
		team_id INTEGER,
		player_id INTEGER,
		PRIMARY KEY(team_id, player_id),
		FOREIGN KEY(team_id) REFERENCES teams(id),
		FOREIGN KEY(player_id) REFERENCES players(id)
	);
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'user'
	);
	`)
	if err != nil {
		log.Fatal(err)
	}
}

// Handler for /teams
func teamsHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check if this is a /teams/{id}/players request
	if strings.Contains(r.URL.Path, "/players") {
		handleTeamPlayers(w, r)
		return
	}

	// Check if this is an operation on a specific team
	id, hasID := getIDFromPath(r.URL.Path)

	switch r.Method {
	case "GET":
		if hasID {
			// Get specific team
			var t Team
			err := db.QueryRow("SELECT id, name FROM teams WHERE id = ?", id).Scan(&t.ID, &t.Name)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(t)
		} else {
			// Get all teams
			rows, err := db.Query("SELECT id, name FROM teams")
			if err != nil {
				w.WriteHeader(500)
				return
			}
			var teams []Team
			for rows.Next() {
				var t Team
				rows.Scan(&t.ID, &t.Name)
				teams = append(teams, t)
			}
			if teams == nil {
				teams = []Team{}
			}
			json.NewEncoder(w).Encode(teams)
		}
	case "POST":
		// Require admin authentication to create teams
		if claims := requireAdmin(w, r); claims == nil {
			return
		}
		// Accept optional players list when creating a team
		var payload struct {
			Name    string `json:"name"`
			Players []int  `json:"players"`
		}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		res, err := db.Exec("INSERT INTO teams (name) VALUES (?)", payload.Name)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		teamID, _ := res.LastInsertId()
		// insert team_players if provided
		for _, pid := range payload.Players {
			_, _ = db.Exec("INSERT OR IGNORE INTO team_players (team_id, player_id) VALUES (?, ?)", teamID, pid)
		}
		w.WriteHeader(201)
	case "PUT":
		// Require admin authentication to update teams
		if claims := requireAdmin(w, r); claims == nil {
			return
		}
		if !hasID {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var payload struct {
			Name    string `json:"name"`
			Players []int  `json:"players"`
		}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		_, err = db.Exec("UPDATE teams SET name = ? WHERE id = ?", payload.Name, id)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		// Update team_players: first delete all existing, then add new ones
		_, _ = db.Exec("DELETE FROM team_players WHERE team_id = ?", id)
		for _, pid := range payload.Players {
			_, _ = db.Exec("INSERT OR IGNORE INTO team_players (team_id, player_id) VALUES (?, ?)", id, pid)
		}
		w.WriteHeader(http.StatusOK)
	case "DELETE":
		// Require admin authentication to delete teams
		if claims := requireAdmin(w, r); claims == nil {
			return
		}
		if !hasID {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// Delete team_players first (due to foreign key)
		_, _ = db.Exec("DELETE FROM team_players WHERE team_id = ?", id)
		_, err := db.Exec("DELETE FROM teams WHERE id = ?", id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// Handler for /teams/{id}/players
func handleTeamPlayers(w http.ResponseWriter, r *http.Request) {
	// Parse team ID from path like /teams/1/players
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	teamID := parts[2]

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get all players for this team
	rows, err := db.Query(`
		SELECT p.id, p.name FROM players p
		JOIN team_players tp ON p.id = tp.player_id
		WHERE tp.team_id = ?
		ORDER BY p.name
	`, teamID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var p Player
		rows.Scan(&p.ID, &p.Name)
		players = append(players, p)
	}
	if players == nil {
		players = []Player{}
	}
	json.NewEncoder(w).Encode(players)
}

// Handler for /players
func playersHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check if this is an operation on a specific player
	id, hasID := getIDFromPath(r.URL.Path)

	switch r.Method {
	case "GET":
		if hasID {
			// Get specific player
			var p Player
			err := db.QueryRow("SELECT id, name FROM players WHERE id = ?", id).Scan(&p.ID, &p.Name)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(p)
		} else {
			// Get all players
			rows, err := db.Query("SELECT id, name FROM players")
			if err != nil {
				w.WriteHeader(500)
				return
			}
			var players []Player
			for rows.Next() {
				var p Player
				rows.Scan(&p.ID, &p.Name)
				players = append(players, p)
			}
			if players == nil {
				players = []Player{}
			}
			json.NewEncoder(w).Encode(players)
		}
	case "POST":
		// Require admin authentication to create players
		if claims := requireAdmin(w, r); claims == nil {
			return
		}
		var p Player
		json.NewDecoder(r.Body).Decode(&p)
		_, err := db.Exec("INSERT INTO players (name) VALUES (?)", p.Name)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(201)
	case "PUT":
		// Require admin authentication to update players
		if claims := requireAdmin(w, r); claims == nil {
			return
		}
		if !hasID {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var p Player
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		_, err = db.Exec("UPDATE players SET name = ? WHERE id = ?", p.Name, id)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(http.StatusOK)
	case "DELETE":
		// Require admin authentication to delete players
		if claims := requireAdmin(w, r); claims == nil {
			return
		}
		if !hasID {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err := db.Exec("DELETE FROM players WHERE id = ?", id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// Handler for /players/stats
func playerStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get season_id from query parameter (0 means all seasons)
	seasonID := r.URL.Query().Get("season_id")
	if seasonID == "" {
		seasonID = "0"
	}

	var rows *sql.Rows
	var err error

	// Build query to get player stats
	// For each player, count how many matches they participated in (as a member of team_a or team_b)
	// and count wins/losses
	if seasonID == "0" {
		// All seasons
		rows, err = db.Query(`
			SELECT 
				p.id,
				p.name,
				COUNT(DISTINCT m.id) as games_played,
				SUM(CASE WHEN m.winner_id = tp.team_id THEN 1 ELSE 0 END) as wins,
				COUNT(DISTINCT m.id) - SUM(CASE WHEN m.winner_id = tp.team_id THEN 1 ELSE 0 END) as losses
			FROM players p
			JOIN team_players tp ON p.id = tp.player_id
			JOIN matches m ON (m.team_a_id = tp.team_id OR m.team_b_id = tp.team_id)
			GROUP BY p.id, p.name
			ORDER BY wins DESC
		`)
	} else {
		// Specific season
		rows, err = db.Query(`
			SELECT 
				p.id,
				p.name,
				COUNT(DISTINCT m.id) as games_played,
				SUM(CASE WHEN m.winner_id = tp.team_id THEN 1 ELSE 0 END) as wins,
				COUNT(DISTINCT m.id) - SUM(CASE WHEN m.winner_id = tp.team_id THEN 1 ELSE 0 END) as losses
			FROM players p
			JOIN team_players tp ON p.id = tp.player_id
			JOIN matches m ON (m.team_a_id = tp.team_id OR m.team_b_id = tp.team_id)
			WHERE m.season_id = ?
			GROUP BY p.id, p.name
			ORDER BY wins DESC
		`, seasonID)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stats []PlayerStat
	for rows.Next() {
		var ps PlayerStat
		var wins, losses, gamesPlayed int
		err := rows.Scan(&ps.ID, &ps.Name, &gamesPlayed, &wins, &losses)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ps.GamesPlayed = gamesPlayed
		ps.Wins = wins
		ps.Losses = losses
		if gamesPlayed > 0 {
			ps.WinPercentage = float64(wins) / float64(gamesPlayed) * 100
		} else {
			ps.WinPercentage = 0
		}
		stats = append(stats, ps)
	}

	if stats == nil {
		stats = []PlayerStat{}
	}

	json.NewEncoder(w).Encode(stats)
}

// Handler for /seasons
func seasonsHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check if this is an operation on a specific season
	id, hasID := getIDFromPath(r.URL.Path)

	switch r.Method {
	case "GET":
		if hasID {
			// Get specific season
			var s Season
			err := db.QueryRow("SELECT id, year FROM seasons WHERE id = ?", id).Scan(&s.ID, &s.Year)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(s)
		} else {
			// Get all seasons
			rows, err := db.Query("SELECT id, year FROM seasons")
			if err != nil {
				w.WriteHeader(500)
				return
			}
			var seasons []Season
			for rows.Next() {
				var s Season
				rows.Scan(&s.ID, &s.Year)
				seasons = append(seasons, s)
			}
			if seasons == nil {
				seasons = []Season{}
			}
			json.NewEncoder(w).Encode(seasons)
		}
	case "POST":
		// Require admin authentication to create seasons
		if claims := requireAdmin(w, r); claims == nil {
			return
		}
		var s Season
		json.NewDecoder(r.Body).Decode(&s)
		_, err := db.Exec("INSERT INTO seasons (year) VALUES (?)", s.Year)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(201)
	case "PUT":
		// Require admin authentication to update seasons
		if claims := requireAdmin(w, r); claims == nil {
			return
		}
		if !hasID {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var s Season
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		_, err = db.Exec("UPDATE seasons SET year = ? WHERE id = ?", s.Year, id)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(http.StatusOK)
	case "DELETE":
		// Require admin authentication to delete seasons
		if claims := requireAdmin(w, r); claims == nil {
			return
		}
		if !hasID {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err := db.Exec("DELETE FROM seasons WHERE id = ?", id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// Handler for /matches
func matchesHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check if this is an operation on a specific match
	id, hasID := getIDFromPath(r.URL.Path)

	switch r.Method {
	case "GET":
		if hasID {
			// Get specific match
			var m Match
			err := db.QueryRow("SELECT id, season_id, team_a_id, team_b_id, score_a, score_b, winner_id, created_at FROM matches WHERE id = ?", id).
				Scan(&m.ID, &m.SeasonID, &m.TeamAID, &m.TeamBID, &m.ScoreA, &m.ScoreB, &m.WinnerID, &m.CreatedAt)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(m)
		} else {
			// Get all matches
			rows, err := db.Query("SELECT id, season_id, team_a_id, team_b_id, score_a, score_b, winner_id, created_at FROM matches")
			if err != nil {
				w.WriteHeader(500)
				return
			}
			var matches []Match
			for rows.Next() {
				var m Match
				rows.Scan(&m.ID, &m.SeasonID, &m.TeamAID, &m.TeamBID, &m.ScoreA, &m.ScoreB, &m.WinnerID, &m.CreatedAt)
				matches = append(matches, m)
			}
			if matches == nil {
				matches = []Match{}
			}
			json.NewEncoder(w).Encode(matches)
		}
	case "POST":
		// Require authenticated user to create matches
		if claims := requireAuth(w, r); claims == nil {
			return
		}
		var m Match
		json.NewDecoder(r.Body).Decode(&m)
		var createdAt interface{} = nil
		if m.CreatedAt != "" {
			createdAt = m.CreatedAt
		}
		_, err := db.Exec("INSERT INTO matches (season_id, team_a_id, team_b_id, score_a, score_b, winner_id, created_at) VALUES (?, ?, ?, ?, ?, ?, COALESCE(?, CURRENT_TIMESTAMP))",
			m.SeasonID, m.TeamAID, m.TeamBID, m.ScoreA, m.ScoreB, m.WinnerID, createdAt)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(201)
	case "PUT":
		// Require admin authentication to update matches
		if claims := requireAdmin(w, r); claims == nil {
			return
		}
		if !hasID {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var m Match
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		_, err = db.Exec("UPDATE matches SET season_id = ?, team_a_id = ?, team_b_id = ?, score_a = ?, score_b = ?, winner_id = ?, created_at = ? WHERE id = ?",
			m.SeasonID, m.TeamAID, m.TeamBID, m.ScoreA, m.ScoreB, m.WinnerID, m.CreatedAt, id)
		if err != nil {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(http.StatusOK)
	case "DELETE":
		// Require admin authentication to delete matches
		if claims := requireAdmin(w, r); claims == nil {
			return
		}
		if !hasID {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err := db.Exec("DELETE FROM matches WHERE id = ?", id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

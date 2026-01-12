package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TeamWithPlayers struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Players []int  `json:"players"`
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
	ID             int    `json:"id"`
	SeasonID       int    `json:"season_id"`
	MatchType      string `json:"match_type"`
	ParticipantAID int    `json:"participant_a_id"`
	ParticipantBID int    `json:"participant_b_id"`
	ScoreA         int    `json:"score_a"`
	ScoreB         int    `json:"score_b"`
	WinnerID       int    `json:"winner_id"`
	CreatedAt      string `json:"created_at"`
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

	// Only set Content-Type if not already set (don't override multipart/form-data)
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
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
	setupWeeklyBackup()

	log.Println("Backend running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initDB() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS teams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		image_path TEXT
	);
	CREATE TABLE IF NOT EXISTS seasons (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		year INTEGER NOT NULL UNIQUE
	);
	CREATE TABLE IF NOT EXISTS matches (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		season_id INTEGER,
		match_type TEXT DEFAULT 'team',
		participant_a_id INTEGER,
		participant_b_id INTEGER,
		score_a INTEGER,
		score_b INTEGER,
		winner_id INTEGER,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(season_id) REFERENCES seasons(id),
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

	// Check if this is a /teams/{id}/image request
	if strings.Contains(r.URL.Path, "/image") {
		handleTeamImage(w, r)
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
			// Get all teams with their players, sorted by games played (most to least)
			rows, err := db.Query(`
				SELECT t.id, t.name
				FROM teams t
				LEFT JOIN (
					SELECT 
						CASE 
							WHEN match_type = 'team' THEN participant_a_id
						END as team_id
					FROM matches
					WHERE match_type = 'team'
					UNION ALL
					SELECT 
						CASE 
							WHEN match_type = 'team' THEN participant_b_id
						END as team_id
					FROM matches
					WHERE match_type = 'team'
				) m ON t.id = m.team_id
				GROUP BY t.id, t.name
				ORDER BY COUNT(m.team_id) DESC, t.name ASC
			`)
			if err != nil {
				w.WriteHeader(500)
				return
			}
			defer rows.Close()

			var teams []TeamWithPlayers
			for rows.Next() {
				var t TeamWithPlayers
				rows.Scan(&t.ID, &t.Name)

				// Get players for this team
				playerRows, err := db.Query(`
					SELECT p.id FROM players p
					JOIN team_players tp ON p.id = tp.player_id
					WHERE tp.team_id = ?
					ORDER BY p.id
				`, t.ID)
				if err != nil {
					continue
				}
				defer playerRows.Close()

				var playerIDs []int
				for playerRows.Next() {
					var pid int
					playerRows.Scan(&pid)
					playerIDs = append(playerIDs, pid)
				}
				playerRows.Close()

				t.Players = playerIDs
				if t.Players == nil {
					t.Players = []int{}
				}
				teams = append(teams, t)
			}
			if teams == nil {
				teams = []TeamWithPlayers{}
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
		if err != nil || payload.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request - name is required"})
			return
		}
		// Trim and validate name
		payload.Name = strings.TrimSpace(payload.Name)
		if len(payload.Name) == 0 || len(payload.Name) > 255 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Team name must be 1-255 characters"})
			return
		}
		res, err := db.Exec("INSERT INTO teams (name) VALUES (?)", payload.Name)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Team name already exists"})
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
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}
		// Trim and validate name
		payload.Name = strings.TrimSpace(payload.Name)
		if len(payload.Name) == 0 || len(payload.Name) > 255 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Team name must be 1-255 characters"})
			return
		}
		_, err = db.Exec("UPDATE teams SET name = ? WHERE id = ?", payload.Name, id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update team"})
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
		// Get and delete team image if it exists
		var imagePath string
		db.QueryRow("SELECT COALESCE(image_path, '') FROM teams WHERE id = ?", id).Scan(&imagePath)
		if imagePath != "" {
			os.Remove(imagePath)
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

// Handler for /teams/{id}/image
func handleTeamImage(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse team ID from path like /teams/1/image
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	teamIDStr := parts[2]
	teamID, err := strconv.ParseInt(teamIDStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		// Get image for team
		var imagePath string
		err := db.QueryRow("SELECT COALESCE(image_path, '') FROM teams WHERE id = ?", teamID).Scan(&imagePath)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if imagePath == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Check if file exists
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Serve the image
		http.ServeFile(w, r, imagePath)

	case "POST":
		// Upload image for team (requires admin)
		if claims := requireAdmin(w, r); claims == nil {
			return
		}

		// Parse multipart form with max 10MB
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			log.Printf("ParseMultipartForm error: %v, Content-Type: %s", err, r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to parse form: %v", err)})
			return
		}

		file, handler, err := r.FormFile("image")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "No image file provided"})
			return
		}
		defer file.Close()

		// Validate file type
		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/webp": true,
		}
		if !allowedTypes[handler.Header.Get("Content-Type")] {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid file type. Only JPEG, PNG, and WebP are allowed"})
			return
		}

		// Create uploads directory if it doesn't exist
		uploadsDir := "/db/uploads/teams"
		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create upload directory"})
			return
		}

		// Get extension from content type
		var ext string
		switch handler.Header.Get("Content-Type") {
		case "image/jpeg":
			ext = ".jpg"
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		}

		// Generate filename
		filename := fmt.Sprintf("team_%d%s", teamID, ext)
		filepath := filepath.Join(uploadsDir, filename)

		// Delete old image if it exists
		var oldImagePath string
		db.QueryRow("SELECT COALESCE(image_path, '') FROM teams WHERE id = ?", teamID).Scan(&oldImagePath)
		if oldImagePath != "" && oldImagePath != filepath {
			os.Remove(oldImagePath)
		}

		// Create the file
		dst, err := os.Create(filepath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to save image"})
			return
		}
		defer dst.Close()

		// Copy file contents
		if _, err := io.Copy(dst, file); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to save image"})
			return
		}

		// Update database with image path
		_, err = db.Exec("UPDATE teams SET image_path = ? WHERE id = ?", filepath, teamID)
		if err != nil {
			os.Remove(filepath)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update team"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Image uploaded successfully"})

	case "DELETE":
		// Delete image for team (requires admin)
		if claims := requireAdmin(w, r); claims == nil {
			return
		}

		var imagePath string
		err := db.QueryRow("SELECT COALESCE(image_path, '') FROM teams WHERE id = ?", teamID).Scan(&imagePath)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if imagePath != "" {
			os.Remove(imagePath)
			_, err = db.Exec("UPDATE teams SET image_path = NULL WHERE id = ?", teamID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Image deleted successfully"})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
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

// generateDoublesPairs creates doubles pair teams for all combinations of players
func generateDoublesPairs(newPlayerID int) error {
	// Get all players
	rows, err := db.Query("SELECT id, name FROM players ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var p Player
		rows.Scan(&p.ID, &p.Name)
		players = append(players, p)
	}

	// Generate pairs with the new player
	for i := 0; i < len(players); i++ {
		if players[i].ID == newPlayerID {
			continue // Skip pairing with itself
		}

		// Create team name (alphabetically sorted for consistency - prevents duplicates like "A & B" and "B & A")
		var team1, team2 string
		player1ID := players[i].ID
		player2ID := newPlayerID

		// Ensure team name is always created in the same order
		if player1ID < player2ID {
			team1 = players[i].Name
			team2 = getPlayerName(newPlayerID)
		} else {
			team1 = getPlayerName(newPlayerID)
			team2 = players[i].Name
		}
		teamName := team1 + " & " + team2

		// Check if team already exists
		var existingID int
		err := db.QueryRow("SELECT id FROM teams WHERE name = ?", teamName).Scan(&existingID)
		if err == sql.ErrNoRows {
			// Team doesn't exist, create it
			res, err := db.Exec("INSERT INTO teams (name) VALUES (?)", teamName)
			if err != nil {
				log.Printf("Error creating doubles pair team: %v", err)
				continue
			}
			teamID, _ := res.LastInsertId()

			// Add both players to the team in consistent order
			minID := player1ID
			maxID := player2ID
			if minID > maxID {
				minID, maxID = maxID, minID
			}

			db.Exec("INSERT OR IGNORE INTO team_players (team_id, player_id) VALUES (?, ?)", teamID, minID)
			db.Exec("INSERT OR IGNORE INTO team_players (team_id, player_id) VALUES (?, ?)", teamID, maxID)
		}
	}
	return nil
}

// getPlayerName returns the name of a player by ID
func getPlayerName(playerID int) string {
	var name string
	db.QueryRow("SELECT name FROM players WHERE id = ?", playerID).Scan(&name)
	return name
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
		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil || p.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request - name is required"})
			return
		}
		// Trim and validate name
		p.Name = strings.TrimSpace(p.Name)
		if len(p.Name) == 0 || len(p.Name) > 255 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Player name must be 1-255 characters"})
			return
		}
		res, err := db.Exec("INSERT INTO players (name) VALUES (?)", p.Name)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Player name already exists"})
			return
		}
		// Get the inserted player ID
		playerID, _ := res.LastInsertId()

		// Generate doubles pairs for the new player
		if err := generateDoublesPairs(int(playerID)); err != nil {
			log.Printf("Error generating doubles pairs: %v", err)
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
		if err != nil || p.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request - name is required"})
			return
		}
		// Trim and validate name
		p.Name = strings.TrimSpace(p.Name)
		if len(p.Name) == 0 || len(p.Name) > 255 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Player name must be 1-255 characters"})
			return
		}
		_, err = db.Exec("UPDATE players SET name = ? WHERE id = ?", p.Name, id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update player"})
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

	// Build query to get player stats - include both team matches (via teams) and singles matches
	// Use UNION to avoid duplicate counting from LEFT JOINs
	if seasonID == "0" {
		// All seasons
		rows, err = db.Query(`
			SELECT 
				p.id,
				p.name,
				COUNT(DISTINCT m.id) as games_played,
				COALESCE(SUM(CASE WHEN m.winner_id = p.id AND m.match_type = 'single' THEN 1 
						 WHEN m.winner_id IN (SELECT team_id FROM team_players WHERE player_id = p.id) AND m.match_type = 'team' THEN 1
						 ELSE 0 END), 0) as wins,
				COALESCE(COUNT(DISTINCT m.id), 0) - COALESCE(SUM(CASE WHEN m.winner_id = p.id AND m.match_type = 'single' THEN 1 
													 WHEN m.winner_id IN (SELECT team_id FROM team_players WHERE player_id = p.id) AND m.match_type = 'team' THEN 1
													 ELSE 0 END), 0) as losses
			FROM players p
			LEFT JOIN matches m ON (
				(m.participant_a_id = p.id OR m.participant_b_id = p.id) AND m.match_type = 'single'
			) OR (
				m.match_type = 'team' AND (
					m.participant_a_id IN (SELECT team_id FROM team_players WHERE player_id = p.id) OR
					m.participant_b_id IN (SELECT team_id FROM team_players WHERE player_id = p.id)
				)
			)
			GROUP BY p.id, p.name
			ORDER BY wins DESC, p.name ASC
		`)
	} else {
		// Specific season
		rows, err = db.Query(`
			SELECT 
				p.id,
				p.name,
				COUNT(DISTINCT m.id) as games_played,
				COALESCE(SUM(CASE WHEN m.winner_id = p.id AND m.match_type = 'single' THEN 1 
						 WHEN m.winner_id IN (SELECT team_id FROM team_players WHERE player_id = p.id) AND m.match_type = 'team' THEN 1
						 ELSE 0 END), 0) as wins,
				COALESCE(COUNT(DISTINCT m.id), 0) - COALESCE(SUM(CASE WHEN m.winner_id = p.id AND m.match_type = 'single' THEN 1 
													 WHEN m.winner_id IN (SELECT team_id FROM team_players WHERE player_id = p.id) AND m.match_type = 'team' THEN 1
													 ELSE 0 END), 0) as losses
			FROM players p
			LEFT JOIN matches m ON (
				(m.participant_a_id = p.id OR m.participant_b_id = p.id) AND m.match_type = 'single'
			) OR (
				m.match_type = 'team' AND (
					m.participant_a_id IN (SELECT team_id FROM team_players WHERE player_id = p.id) OR
					m.participant_b_id IN (SELECT team_id FROM team_players WHERE player_id = p.id)
				)
			)
			WHERE m.id IS NULL OR m.season_id = ?
			GROUP BY p.id, p.name
			ORDER BY wins DESC, p.name ASC
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
		err := json.NewDecoder(r.Body).Decode(&s)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}
		// Validate season year is reasonable (1900-2100)
		if s.Year < 1900 || s.Year > 2100 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Season year must be between 1900 and 2100"})
			return
		}
		_, err = db.Exec("INSERT INTO seasons (year) VALUES (?)", s.Year)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Season year already exists"})
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
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}
		// Validate season year is reasonable (1900-2100)
		if s.Year < 1900 || s.Year > 2100 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Season year must be between 1900 and 2100"})
			return
		}
		_, err = db.Exec("UPDATE seasons SET year = ? WHERE id = ?", s.Year, id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update season"})
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
			err := db.QueryRow("SELECT id, season_id, match_type, participant_a_id, participant_b_id, score_a, score_b, winner_id, created_at FROM matches WHERE id = ?", id).
				Scan(&m.ID, &m.SeasonID, &m.MatchType, &m.ParticipantAID, &m.ParticipantBID, &m.ScoreA, &m.ScoreB, &m.WinnerID, &m.CreatedAt)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(m)
		} else {
			// Get all matches
			rows, err := db.Query("SELECT id, season_id, match_type, participant_a_id, participant_b_id, score_a, score_b, winner_id, created_at FROM matches")
			if err != nil {
				w.WriteHeader(500)
				return
			}
			var matches []Match
			for rows.Next() {
				var m Match
				rows.Scan(&m.ID, &m.SeasonID, &m.MatchType, &m.ParticipantAID, &m.ParticipantBID, &m.ScoreA, &m.ScoreB, &m.WinnerID, &m.CreatedAt)
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
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		// Validate required fields
		if m.SeasonID == 0 || m.ParticipantAID == 0 || m.ParticipantBID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing required fields: season_id, participant_a_id, participant_b_id"})
			return
		}

		// Validate match type
		if m.MatchType == "" {
			m.MatchType = "team"
		}
		if m.MatchType != "team" && m.MatchType != "single" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid match_type: must be 'team' or 'single'"})
			return
		}

		// Validate participants are different
		if m.ParticipantAID == m.ParticipantBID {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Participant A and B must be different"})
			return
		}

		// Validate scores are non-negative
		if m.ScoreA < 0 || m.ScoreB < 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Scores cannot be negative"})
			return
		}

		// If winner is specified, validate it matches one of the participants
		if m.WinnerID != 0 && m.WinnerID != m.ParticipantAID && m.WinnerID != m.ParticipantBID {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Winner must be one of the participants"})
			return
		}

		var createdAt interface{} = nil
		if m.CreatedAt != "" {
			createdAt = m.CreatedAt
		}

		_, err = db.Exec("INSERT INTO matches (season_id, match_type, participant_a_id, participant_b_id, score_a, score_b, winner_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, COALESCE(?, CURRENT_TIMESTAMP))",
			m.SeasonID, m.MatchType, m.ParticipantAID, m.ParticipantBID, m.ScoreA, m.ScoreB, m.WinnerID, createdAt)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create match"})
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
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		// Validate required fields
		if m.SeasonID == 0 || m.ParticipantAID == 0 || m.ParticipantBID == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing required fields: season_id, participant_a_id, participant_b_id"})
			return
		}

		// Validate match type
		if m.MatchType != "team" && m.MatchType != "single" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid match_type: must be 'team' or 'single'"})
			return
		}

		// Validate participants are different
		if m.ParticipantAID == m.ParticipantBID {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Participant A and B must be different"})
			return
		}

		// Validate scores are non-negative
		if m.ScoreA < 0 || m.ScoreB < 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Scores cannot be negative"})
			return
		}

		// If winner is specified, validate it matches one of the participants
		if m.WinnerID != 0 && m.WinnerID != m.ParticipantAID && m.WinnerID != m.ParticipantBID {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Winner must be one of the participants"})
			return
		}

		_, err = db.Exec("UPDATE matches SET season_id = ?, match_type = ?, participant_a_id = ?, participant_b_id = ?, score_a = ?, score_b = ?, winner_id = ?, created_at = ? WHERE id = ?",
			m.SeasonID, m.MatchType, m.ParticipantAID, m.ParticipantBID, m.ScoreA, m.ScoreB, m.WinnerID, m.CreatedAt, id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update match"})
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

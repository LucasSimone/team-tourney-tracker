package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func initJWTSecret() {
	// Try to load JWT secret from environment
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("WARNING: JWT_SECRET not set in environment. Using insecure default key.")
		log.Println("WARNING: In production, set JWT_SECRET environment variable to a strong random string.")
		secret = "your-secret-key-change-this-in-production"
	}
	jwtSecret = []byte(secret)
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Default users stored in code - not in database
var defaultUsers = map[string]struct {
	PasswordHash string
	Role         string
	ID           int
}{
	"lucas":    {PasswordHash: "0e0e253150db68d7fed1999bde814c9c9cfd796abf3dae5d7ca38476943c4020", Role: "admin", ID: 1},
	"baddyboi": {PasswordHash: "9f168976c19f6db8bb9e8d83e6f054ab98572ef620111e09cff59c419cd56c9d", Role: "user", ID: 2},
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func verifyPassword(hash, password string) bool {
	return hash == hashPassword(password)
}

func generateToken(user User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func verifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// requireAuth validates the JWT token from the Authorization header
// Returns claims if valid, or nil if invalid. Writes error response if invalid.
func requireAuth(w http.ResponseWriter, r *http.Request) *Claims {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("missing authorization header"))
		return nil
	}

	// Extract token from "Bearer <token>" format
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		// No "Bearer " prefix found
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid authorization header format"))
		return nil
	}

	claims, err := verifyToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("invalid token"))
		return nil
	}

	return claims
}

// requireAdmin validates token and checks for admin role
func requireAdmin(w http.ResponseWriter, r *http.Request) *Claims {
	claims := requireAuth(w, r)
	if claims == nil {
		return nil
	}

	if claims.Role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("admin access required"))
		return nil
	}

	return claims
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	var user User

	// Check hardcoded default users first
	if defaultUser, exists := defaultUsers[payload.Username]; exists {
		if !verifyPassword(defaultUser.PasswordHash, payload.Password) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
			return
		}
		user = User{
			ID:       defaultUser.ID,
			Username: payload.Username,
			Role:     defaultUser.Role,
		}
	} else {
		// Check database for admin-created users
		var passwordHash string
		err = db.QueryRow("SELECT id, username, role, password FROM users WHERE username = ?", payload.Username).
			Scan(&user.ID, &user.Username, &user.Role, &passwordHash)

		if err != nil || !verifyPassword(passwordHash, payload.Password) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid credentials"})
			return
		}
	}

	token, err := generateToken(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
		return
	}

	passwordHash := hashPassword(payload.Password)
	res, err := db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", payload.Username, passwordHash, "user")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Username already exists"})
		return
	}

	userID, _ := res.LastInsertId()
	user := User{
		ID:       int(userID),
		Username: payload.Username,
		Role:     "user",
	}

	token, err := generateToken(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing authorization"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := verifyToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token"})
		return
	}

	// Only admins can manage users
	if claims.Role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Admin access required"})
		return
	}

	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, username, role FROM users")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var users []User
		for rows.Next() {
			var u User
			rows.Scan(&u.ID, &u.Username, &u.Role)
			users = append(users, u)
		}
		if users == nil {
			users = []User{}
		}
		json.NewEncoder(w).Encode(users)

	case "POST":
		var payload struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if payload.Role == "" {
			payload.Role = "user"
		}

		passwordHash := hashPassword(payload.Password)
		_, err = db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", payload.Username, passwordHash, payload.Role)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Username already exists"})
			return
		}
		w.WriteHeader(http.StatusCreated)

	case "PUT":
		// Handle user updates - both role and password
		// Parse the URL for the user ID pattern: /auth/users/{id}
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userID := parts[3]
		var payload struct {
			Role     string `json:"role"`
			Password string `json:"password"`
		}
		json.NewDecoder(r.Body).Decode(&payload)

		// Build dynamic update query
		if payload.Password != "" && payload.Role != "" {
			passwordHash := hashPassword(payload.Password)
			_, err := db.Exec("UPDATE users SET role = ?, password = ? WHERE id = ?", payload.Role, passwordHash, userID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if payload.Role != "" {
			_, err := db.Exec("UPDATE users SET role = ? WHERE id = ?", payload.Role, userID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else if payload.Password != "" {
			passwordHash := hashPassword(payload.Password)
			_, err := db.Exec("UPDATE users SET password = ? WHERE id = ?", passwordHash, userID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusOK)

	case "DELETE":
		// Handle user deletion
		// Parse the URL for the user ID pattern: /auth/users/{id}
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err := db.Exec("DELETE FROM users WHERE id = ?", parts[3])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

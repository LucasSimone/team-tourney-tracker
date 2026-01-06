package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// createLocalBackup creates a backup file in the backups directory
func createLocalBackup() (string, error) {
	// Create backups directory
	backupDir := "/db/backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Read the database file
	dbPath := "/db/sports.db"
	dbData, err := os.ReadFile(dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to read database file: %w", err)
	}

	// Create backup file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("sports_backup_%s.db", timestamp))

	if err := os.WriteFile(backupPath, dbData, 0644); err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}

	log.Printf("Local backup created: %s", backupPath)

	// Clean up old backups (keep only last 3)
	cleanupOldBackups(backupDir)

	return backupPath, nil
}

// cleanupOldBackups keeps only the 3 most recent backups
func cleanupOldBackups(backupDir string) {
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return
	}

	// Filter to only .db files and sort by modification time (newest first)
	type fileInfo struct {
		name    string
		modTime time.Time
	}

	var backups []fileInfo
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".db") {
			continue
		}

		info, _ := file.Info()
		backups = append(backups, fileInfo{
			name:    file.Name(),
			modTime: info.ModTime(),
		})
	}

	// Sort by modification time (newest first)
	for i := 0; i < len(backups); i++ {
		for j := i + 1; j < len(backups); j++ {
			if backups[j].modTime.After(backups[i].modTime) {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	// Delete backups beyond the 3 most recent
	for i := 3; i < len(backups); i++ {
		filepath := filepath.Join(backupDir, backups[i].name)
		if err := os.Remove(filepath); err != nil {
			log.Printf("Failed to delete backup %s: %v", backups[i].name, err)
		} else {
			log.Printf("Deleted old backup: %s", backups[i].name)
		}
	}
}

// setupWeeklyBackup starts a goroutine that backs up the database weekly
func setupWeeklyBackup() {
	go func() {
		// Run once immediately at startup
		log.Println("Starting weekly database backup scheduler...")
		if _, err := createLocalBackup(); err != nil {
			log.Printf("Initial backup failed: %v", err)
		}

		// Run weekly
		ticker := time.NewTicker(7 * 24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			if _, err := createLocalBackup(); err != nil {
				log.Printf("Weekly backup failed: %v", err)
			}
		}
	}()
}

// backupHandler handles manual backup requests from the admin panel
func backupHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	setCORSHeaders(w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check authentication and authorization
	if r.Method == "POST" {
		// Verify user is authenticated and is an admin
		claims := requireAdmin(w, r)
		if claims == nil {
			return
		}

		// Create local backup
		backupPath, err := createLocalBackup()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("backup failed: %v", err)))
			log.Printf("Backup error: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"status":"success","message":"backup created at %s"}`, backupPath)))
		return
	}

	if r.Method == "GET" {
		// Verify user is authenticated and is an admin
		claims := requireAdmin(w, r)
		if claims == nil {
			return
		}

		backupDir := "/db/backups"
		files, err := os.ReadDir(backupDir)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"backups":[]}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"backups":[`))

		first := true
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".db") {
				continue
			}

			info, _ := file.Info()
			if !first {
				w.Write([]byte(","))
			}
			w.Write([]byte(fmt.Sprintf(`{"name":"%s","size":%d,"modified":"%s"}`,
				file.Name(),
				info.Size(),
				info.ModTime().Format(time.RFC3339))))
			first = false
		}

		w.Write([]byte(`]}`))
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// backupDownloadHandler serves backup files for download
func backupDownloadHandler(w http.ResponseWriter, r *http.Request) {
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

	// Verify user is authenticated and is an admin
	claims := requireAdmin(w, r)
	if claims == nil {
		return
	}

	// Get filename from query parameter
	filename := r.URL.Query().Get("file")
	if filename == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing file parameter"))
		return
	}

	// Prevent directory traversal attacks
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid filename"))
		return
	}

	backupPath := filepath.Join("/db/backups", filename)

	// Verify file exists and is readable
	fileInfo, err := os.Stat(backupPath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("backup file not found"))
		return
	}

	if !fileInfo.Mode().IsRegular() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid file"))
		return
	}

	// Serve the file
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	fileData, err := os.ReadFile(backupPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to read backup file"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(fileData)
}

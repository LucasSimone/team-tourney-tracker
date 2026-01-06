package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	mail "github.com/go-mail/mail/v2"
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

// emailDatabaseBackup sends the current database as an email attachment
func emailDatabaseBackup(recipientEmail string) error {
	// Get SMTP configuration from environment
	smtpHost := os.Getenv("SMTP_HOST")
	senderEmail := os.Getenv("SENDER_EMAIL")
	appPassword := os.Getenv("SENDER_APP_PASSWORD")

	// If email not configured, just do local backup
	if smtpHost == "" || senderEmail == "" || appPassword == "" {
		log.Println("Email not configured, creating local backup only")
		_, err := createLocalBackup()
		return err
	}

	// Extract host and port from SMTP_HOST (format: "smtp.gmail.com" or "smtp.gmail.com:465")
	smtpPort := 587 // Default port for TLS
	if strings.Contains(smtpHost, ":") {
		parts := strings.Split(smtpHost, ":")
		smtpHost = parts[0]
		if p, err := strconv.Atoi(parts[1]); err == nil {
			smtpPort = p
		}
	}

	// Allow override via SMTP_PORT environment variable
	if envPort := os.Getenv("SMTP_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			smtpPort = p
		}
	}

	log.Printf("Connecting to SMTP server: %s:%d", smtpHost, smtpPort)

	// Read the database file
	dbPath := "/db/sports.db"
	dbData, err := os.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf("failed to read database file: %w", err)
	}

	// Create email message
	m := mail.NewMessage()
	m.SetHeader("From", senderEmail)
	m.SetHeader("To", recipientEmail)
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	m.SetHeader("Subject", fmt.Sprintf("Tournament Tracker Database Backup - %s", timestamp))

	// Email body
	body := fmt.Sprintf(`
Database Backup Report
======================

Backup Time: %s
Database Size: %d bytes

This is an automated weekly backup of the Tournament Tracker database.

To restore, simply replace the sports.db file with this backup.

Best regards,
Tournament Tracker System
`, time.Now().Format("2006-01-02 15:04:05 MST"), len(dbData))

	m.SetBody("text/plain", body)

	// Attach database file
	filename := fmt.Sprintf("sports_backup_%s.db", timestamp)
	m.Attach(dbPath, mail.SetHeader(map[string][]string{
		"Content-Disposition": {fmt.Sprintf("attachment; filename=%q", filename)},
	}))

	// Send email via Gmail SMTP
	dialer := mail.NewDialer(smtpHost, smtpPort, senderEmail, appPassword)

	// Port 465 uses implicit TLS (SMTPS), port 587 uses explicit TLS (STARTTLS)
	if smtpPort == 465 {
		dialer.SSL = true
	}

	// Configure TLS
	dialer.TLSConfig = &tls.Config{
		ServerName:         smtpHost,
		InsecureSkipVerify: false,
	}

	if err := dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Database backup email sent successfully to %s", recipientEmail)
	return nil
}

// setupWeeklyBackup starts a goroutine that backs up the database weekly
func setupWeeklyBackup(recipientEmail string) {
	go func() {
		// Calculate time until next Monday at 2 AM
		ticker := time.NewTicker(7 * 24 * time.Hour)
		defer ticker.Stop()

		// Run once immediately at startup
		log.Println("Starting weekly database backup scheduler...")
		if err := emailDatabaseBackup(recipientEmail); err != nil {
			log.Printf("Initial backup failed: %v", err)
		}

		// Run weekly
		for range ticker.C {
			if err := emailDatabaseBackup(recipientEmail); err != nil {
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

		// Try to send email if configured
		backupEmail := os.Getenv("BACKUP_EMAIL")
		if backupEmail != "" {
			if err := emailDatabaseBackup(backupEmail); err != nil {
				log.Printf("Email backup failed (local backup still created): %v", err)
			}
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

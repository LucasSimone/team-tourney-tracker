package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	mail "github.com/go-mail/mail/v2"
)

// emailDatabaseBackup sends the current database as an email attachment
func emailDatabaseBackup(recipientEmail string) error {
	// Get SMTP configuration from environment
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 587 // Gmail uses port 587 for TLS
	senderEmail := os.Getenv("SENDER_EMAIL")
	appPassword := os.Getenv("SENDER_APP_PASSWORD")

	// Validate required environment variables
	if smtpHost == "" || senderEmail == "" || appPassword == "" {
		return fmt.Errorf("missing required environment variables: SMTP_HOST, SENDER_EMAIL, SENDER_APP_PASSWORD")
	}

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
	// In production, proper TLS will be used. InsecureSkipVerify is disabled for security.
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: false}

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

		backupEmail := os.Getenv("BACKUP_EMAIL")
		if backupEmail == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("backup configuration error"))
			log.Printf("Error: BACKUP_EMAIL environment variable not set")
			return
		}

		if err := emailDatabaseBackup(backupEmail); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("backup failed"))
			log.Printf("Backup error: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("backup completed successfully"))
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

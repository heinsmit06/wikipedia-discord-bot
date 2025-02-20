package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"discord-wiki-bot/internal/wiki"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/r3labs/sse/v2"
)

func Connect() (*sql.DB, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	return sql.Open("postgres", connStr)
}

// collects events from Wikipedia stream
func CollectEvents(eventChan chan<- wiki.WikiEvent) error {
	client := sse.NewClient("https://stream.wikimedia.org/v2/stream/recentchange")
	fmt.Println("Stats collector started")

	for {
		err := client.Subscribe("recent-changes", func(msg *sse.Event) {
			var event wiki.WikiEvent
			if err := json.Unmarshal(msg.Data, &event); err != nil {
				log.Printf("Error unmarshalling event: %v", err)
				return
			}

			select {
			case eventChan <- event:
			case <-time.After(time.Second):
				log.Printf("Warning: event channel blocked for 1s, dropping event")
			}
		})

		if err != nil {
			return fmt.Errorf("SSE subscription error: %v", err)
		}
	}
}

// processes collected events and updates the database
func ProcessEvents(db *sql.DB, eventChan chan wiki.WikiEvent) error {
	counts := make(map[string]int)
	today := time.Now().UTC().Format("2006-01-02")

	// Process events in batches with a shorter timeout
	timeout := time.After(2 * time.Second)
	processed := 0
	batchSize := 1000 // Process up to 1000 events per batch

	for processed < batchSize {
		select {
		case event := <-eventChan:
			lang := event.DetectLanguage()
			counts[lang]++
			processed++
		case <-timeout:
			goto ProcessCounts
		default:
			if len(eventChan) == 0 {
				goto ProcessCounts
			}
		}
	}

ProcessCounts:
	if len(counts) == 0 {
		return nil // No events to process
	}

	// Update database in a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Prepare statement for better performance
	stmt, err := tx.Prepare(`
		INSERT INTO daily_stats (date, language, change_count)
		VALUES ($1, $2, $3)
		ON CONFLICT (date, language)
		DO UPDATE SET 
			change_count = daily_stats.change_count + $3,
			last_updated = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for lang, count := range counts {
		if _, err := stmt.Exec(today, lang, count); err != nil {
			return fmt.Errorf("failed to update stats for %s: %v", lang, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Successfully processed %d events for %d languages", processed, len(counts))
	return nil
}

// GetStats retrieves statistics for a specific date and language
func GetStats(db *sql.DB, date string, language string) (int, time.Time, error) {
	var count int
	var lastUpdated time.Time

	query := `
		SELECT change_count, last_updated
		FROM daily_stats
		WHERE date = $1 AND language = $2
	`

	err := db.QueryRow(query, date, language).Scan(&count, &lastUpdated)
	if err == sql.ErrNoRows {
		return 0, time.Time{}, nil
	}
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("failed to get stats: %v", err)
	}

	return count, lastUpdated, nil
}

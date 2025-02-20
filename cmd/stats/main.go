package main

import (
	"discord-wiki-bot/internal/db"
	"discord-wiki-bot/internal/wiki"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	conn, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	// Increase buffer size to handle more events between processing
	eventChan := make(chan wiki.WikiEvent, 10000)

	// Process more frequently - every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("Starting stats collector...")

	// Start collector in background
	go func() {
		for {
			if err := db.CollectEvents(eventChan); err != nil {
				log.Printf("Collector error: %v, restarting...", err)
				time.Sleep(time.Second) // Wait before retry
			}
		}
	}()

	// Wait for interrupt signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Process events continuously until interrupted
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := db.ProcessEvents(conn, eventChan); err != nil {
					log.Printf("Processing error: %v", err)
				}
			case <-sc:
				return
			}
		}
	}()

	log.Println("Stats collector is running. Press Ctrl+C to stop.")
	<-sc
	log.Println("Stats collector is shutting down...")
}

package wiki

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/r3labs/sse/v2"
)

type WikiEvent struct {
	Title     string `json:"title"`
	TitleURL  string `json:"title_url"`
	User      string `json:"user"`
	Bot       bool   `json:"bot"`
	Timestamp int    `json:"timestamp"`
	Meta      struct {
		Domain string `json:"domain"`
	} `json:"meta"`
	ServerName string `json:"server_name"`
}

// returns the language code for the event
func (e *WikiEvent) DetectLanguage() string {
	// Wikidata case
	if e.Meta.Domain == "www.wikidata.org" || e.ServerName == "www.wikidata.org" {
		return "en"
	}

	// from domain
	if e.Meta.Domain != "" {
		if code := extractLanguageCode(e.Meta.Domain); code != "" {
			return code
		}
	}

	// server_name as fallback
	if e.ServerName != "" {
		if code := extractLanguageCode(e.ServerName); code != "" {
			return code
		}
	}

	return "en"
}

// gets language code from domain ("es.wikipedia.org" -> "es")
func extractLanguageCode(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) > 0 && len(parts[0]) == 2 {
		return parts[0]
	}
	return ""
}

// returns events filtered by the specified language with a timeout
func ParseEvents(language string) []WikiEvent {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := sse.NewClient("https://stream.wikimedia.org/v2/stream/recentchange")
	fmt.Println("Client started")

	var wg sync.WaitGroup
	var events []WikiEvent
	count := 0
	maxEvents := 10

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer fmt.Println("Client shutting down...")

		errChan := make(chan error, 1)

		go func() {
			err := client.SubscribeWithContext(ctx, "recent-changes", func(msg *sse.Event) {
				if count < maxEvents {
					var event WikiEvent
					err := json.Unmarshal(msg.Data, &event)
					if err != nil {
						log.Printf("Error unmarshalling event: %v", err)
						return
					}

					// Only add events matching the requested language
					if event.DetectLanguage() == language {
						events = append(events, event)
						count++
						if count >= maxEvents {
							cancel()
						}
					}
				}
			})
			errChan <- err
		}()

		select {
		case err := <-errChan:
			if err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}()

	wg.Wait()
	return events
}

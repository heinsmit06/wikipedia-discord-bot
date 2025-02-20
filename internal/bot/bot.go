package bot

import (
	"database/sql"
	"discord-wiki-bot/internal/db"
	"discord-wiki-bot/internal/wiki"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Bot struct {
	prefix           string
	channelLanguages map[string]string
	db               *sql.DB
	session          *discordgo.Session
}

func NewBot() (*Bot, error) {
	err := godotenv.Load(".env")
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	// Initialize database connection
	dbConn, err := db.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN is empty")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %v", err)
	}

	bot := &Bot{
		prefix:           "!wikibot",
		channelLanguages: make(map[string]string),
		db:               dbConn,
		session:          session,
	}

	session.AddHandler(bot.handleMessage)
	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	return bot, nil
}

func (b *Bot) getChannelLanguage(channelID string) string {
	if lang, exists := b.channelLanguages[channelID]; exists {
		return lang
	}
	return "en"
}

func (b *Bot) Close() {
	if b.db != nil {
		b.db.Close()
	}
	if b.session != nil {
		b.session.Close()
	}
}

func DiscordBotRun() error {
	bot, err := NewBot()
	if err != nil {
		return fmt.Errorf("failed to create bot: %v", err)
	}
	defer bot.Close()

	err = bot.session.Open()
	if err != nil {
		return fmt.Errorf("error opening connection to Discord: %v", err)
	}

	fmt.Println("Bot is running...")

	// Wait for interrupt signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("Bot is shutting down...")
	return nil
}

func (b *Bot) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	args := strings.Split(m.Content, " ")
	if args[0] != b.prefix {
		return
	}

	if len(args) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Please provide a command. Use `!wikibot help` for usage information.")
		return
	}

	if args[1] == "help" {
		s.ChannelMessageSend(m.ChannelID, `Usage: !wikibot [args]
			
args:
	recent           - show recent changes in english
	recent [code]    - show recent changes in specified language
	setLang [code]   - set channel language
	getLang          - show current channel language
	stats [yyyy-mm-dd] - show statistics for the specified date in current language
	stats [yyyy-mm-dd] [code] - show statistics for the specified date and language
		`)
		return
	}

	// check for the correct number of arguments
	if len(args) > 3 {
		s.ChannelMessageSend(m.ChannelID, "Incorrect number of arguments, please refer to `!wikibot help`")
		return
	}

	// getting language from command arguments first, then fallback to channel settings
	language := "en"
	if len(args) == 3 {
		language = args[2]
	} else {
		language = b.getChannelLanguage(m.ChannelID)
	}

	if args[1] == "setLang" {
		b.channelLanguages[m.ChannelID] = language
		s.ChannelMessageSend(m.ChannelID, "Setting channel language to "+language)
		return
	}

	if args[1] == "getLang" {
		s.ChannelMessageSend(m.ChannelID, "Current channel language is "+b.getChannelLanguage(m.ChannelID))
		return
	}

	if args[1] == "stats" {
		if len(args) < 3 {
			s.ChannelMessageSend(m.ChannelID, "Please provide a date in yyyy-mm-dd format")
			return
		}

		// Validate date format
		date := args[2]
		_, err := time.Parse("2006-01-02", date)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Invalid date format. Please use yyyy-mm-dd")
			return
		}

		// Get language from the third argument if provided, otherwise use channel language
		language := b.getChannelLanguage(m.ChannelID)
		if len(args) == 4 { // If we have a language argument
			language = args[3]
		}

		count, lastUpdated, err := db.GetStats(b.db, date, language)
		if err != nil {
			log.Printf("Error getting stats: %v", err)
			s.ChannelMessageSend(m.ChannelID, "Failed to get statistics")
			return
		}

		// If no stats found for this date/language combination
		if count == 0 && lastUpdated.IsZero() {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No statistics found for language '%s' on %s", language, date))
			return
		}

		embed := &discordgo.MessageEmbed{
			Title: fmt.Sprintf("Wikipedia Changes Stats"),
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "Date",
					Value:  date,
					Inline: true,
				},
				{
					Name:   "Language",
					Value:  strings.ToUpper(language),
					Inline: true,
				},
				{
					Name:   "Change Count",
					Value:  fmt.Sprintf("%d changes", count),
					Inline: true,
				},
			},
		}

		if !lastUpdated.IsZero() {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   "Last Updated",
				Value:  lastUpdated.Format("2006-01-02 15:04:05 MST"),
				Inline: false,
			})
		}

		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		return
	}

	if args[1] == "recent" {
		events := wiki.ParseEvents(language)

		// checking if all events are empty
		allEmpty := true
		for _, event := range events {
			if event.TitleURL != "" || event.Title != "" {
				allEmpty = false
				break
			}
		}

		if allEmpty {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No changes found for language '%s' within 5 seconds. There might be no such language.", language))
			return
		}

		embed := discordgo.MessageEmbed{
			URL:         "https://stream.wikimedia.org/v2/stream/recentchange",
			Title:       fmt.Sprintf("recent changes (%s)", language),
			Description: "",
		}

		for i, event := range events {
			embed.Description = fmt.Sprintf("%d. title: %s\n title_url: %s\n user: %s\n bot: %v\n timestamp: %s\n\n",
				i+1,
				event.Title,
				event.TitleURL,
				event.User,
				event.Bot,
				time.Unix(int64(event.Timestamp), 0).Format("02 Jan 2006 15:04:05 MST"))
			s.ChannelMessageSendEmbed(m.ChannelID, &embed)
		}
		return
	}
}

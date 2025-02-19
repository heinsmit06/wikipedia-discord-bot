package bot

import (
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

var (
	prefix = "!wikibot"
	// stores channel language preferences
	channelLanguages = make(map[string]string)
)

func getChannelLanguage(channelID string) string {
	if lang, exists := channelLanguages[channelID]; exists {
		return lang
	}
	return "en"
}

func DiscordBotRun() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is empty")
	}

	session, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	session.AddHandler(handleMessage)

	session.Identify.Intents = discordgo.IntentsAllWithoutPrivileged
	err = session.Open()
	if err != nil {
		log.Fatalf("Error opening connection to Discord: %v", err)
	}
	defer session.Close()

	fmt.Println("Bot is running...")

	// waiting here until one of the signals is received
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("Bot is shutting down...")
}

func handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	args := strings.Split(m.Content, " ")
	if args[0] != prefix {
		return
	}

	if args[1] == "help" {
		s.ChannelMessageSend(m.ChannelID, `Usage: !wikibot [args]
			
args:
	recent           - show recent changes in english
	recent [code]    - show recent changes in specified language
	setLang [code]   - set channel language
	getLang          - show current channel language
		`)
		return
	}

	// check for the correct number of arguments
	if len(args) < 2 || len(args) > 3 {
		s.ChannelMessageSend(m.ChannelID, "Incorrect number of arguments, please refer to `!wikibot help`")
		return
	}

	// getting language from command arguments first, then fallback to channel settings
	language := "en"
	if len(args) == 3 { // && len(args[2]) == 2
		language = args[2]
	} else {
		language = getChannelLanguage(m.ChannelID)
	}

	if args[1] == "setLang" {
		channelLanguages[m.ChannelID] = language
		s.ChannelMessageSend(m.ChannelID, "Setting channel language to "+language)
		return
	}

	if args[1] == "getLang" {
		s.ChannelMessageSend(m.ChannelID, "Current channel language is "+getChannelLanguage(m.ChannelID))
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

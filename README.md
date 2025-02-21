# Wikipedia Discord Bot

A Discord bot that tracks and provides statistics about Wikipedia changes in real-time, with language-specific filtering capabilities. The bot connects to Wikipedia's Recent Changes Stream and allows users to view recent changes and daily statistics for different languages.

## Features

- Real-time monitoring of Wikipedia changes
- Language-specific filtering
- Daily statistics tracking
- Discord commands for interaction
- PostgreSQL for persistent storage

## Prerequisites

- Go 1.24 or higher
- PostgreSQL
- Discord Bot Token

## Setup

1. Clone the repository:
```bash
git clone git@github.com:heinsmit06/wikipedia-discord-bot.git
cd wikipedia-discord-bot
```

2. Create a `.env` file in the root directory with the following variables:
```env
DISCORD_TOKEN=your_discord_bot_token
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=your_postgres_user
POSTGRES_PASSWORD=your_postgres_password
POSTGRES_DB=your_database_name
```

3. Set up the PostgreSQL database:
```sql
CREATE TABLE daily_stats (
    date DATE NOT NULL,
    language VARCHAR(10) NOT NULL,
    change_count INTEGER NOT NULL DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (date, language)
);
```

## Running the Bot

The bot consists of two services that need to be run separately:

1. **Stats Collector Service** (collects and stores statistics):
```bash
go run cmd/stats/main.go
```

2. **Discord Bot Service** (handles Discord interactions):
```bash
go run cmd/bot/main.go
```

**Important Note**: The stats collector service needs to be running to collect daily statistics. If you want accurate statistics, you should deploy this service to run 24/7 on a server. Otherwise, statistics will only be collected while the service is running.

## Discord Commands

- `!wikibot help` - Show available commands
- `!wikibot recent` - Show recent changes in English
- `!wikibot recent [code]` - Show recent changes in specified language
- `!wikibot setLang [code]` - Set channel default language
- `!wikibot getLang` - Show current channel language
- `!wikibot stats [yyyy-mm-dd]` - Show statistics for the specified date in current language
- `!wikibot stats [yyyy-mm-dd] [code]` - Show statistics for the specified date and language

## Architecture

The project uses a two-service architecture:
1. Stats Service: Connects to Wikipedia SSE stream, processes events, and stores statistics in PostgreSQL
2. Discord Bot: Handles user interactions and queries the database for statistics

### Potential Scalability with Kafka

While the current architecture is sufficient for basic usage, it could be enhanced using Kafka for better scalability. Here are two possible Kafka integration approaches:

1. **Simple Kafka Integration**:
```
[SSE Stream from Wikipedia] -> [Bot] -> [Kafka Topic] -> [Analytics Service]
                                                     \-> [Discord Notifications]
```

2. **Multi-Source Kafka Integration**:
```
[Wikipedia EN]    \
[Wikipedia RU]     -> [Bot] -> [Kafka Topic] -> [Discord Bot Instance 1]
[Other source]    /                         \-> [Discord Bot Instance 2]
                                           \-> [Analytics Service]
                                           \-> [Monitoring Service]
```

In this architecture:
- Multiple sources (Wikipedia EN, RU, Other) send events to your bot
- Bot acts as a producer and sends all events to Kafka
- Multiple consumers (bot instances, analytics, monitoring) independently read events from Kafka

This enhanced architecture would provide:
- Better scalability for high-volume events
- Fault tolerance and data replay capabilities
- More sophisticated analytics possibilities
- Easier integration with other services

💡 **When to Use Kafka**

Currently, there is no need for Kafka because the event volume is low, and the setup is simple.

Consider implementing Kafka if:
- You start adding more event sources
- You need more than just your bot to consume the data
- You face scalability issues, such as Discord API rate limits

## Contributing

This project is a work in progress and will be updated as time goes on, specifically the architecture and organization of the code.

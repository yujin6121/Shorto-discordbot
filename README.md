# Shorto-discordbot Discord Bot

A Discord bot written in Go that integrates with the Shorto URL shortener API. It supports slash commands for shortening URLs, checking domains, and viewing statistics.

## ✨ Features

- **Slash Commands**: Modern Discord interaction.
- **API Integration**: Connects to the Shorto backend.
- **Rich Formatting**: Clean JSON output in Discord messages.

## 📌 Commands

| Command | Description |
| :--- | :--- |
| `/shorten <url> [custom_code]` | Shortens a long URL. Optionally provide a custom code. |
| `/domains` | Lists all available domains for shortening. |
| `/stats` | Shows usage statistics for the service. |

## 🚀 Setup

### Prerequisites

- [Go](https://golang.org/doc/install) 1.21 or later.
- A Discord Bot Token and Application ID (from the [Discord Developer Portal](https://discord.com/developers/applications)).
- The Shorto API server running.

### Installation

1. Install dependencies:
   ```bash
   go mod tidy
   ```

2. Configure environment variables:
   ```bash
   cp .env.example .env
   ```
   Edit `.env` and fill in your credentials:
   - `DISCORD_TOKEN`: Your bot's secret token.
   - `APP_ID`: Your Discord application ID.
   - `API_BASE_URL`: The URL of your Shorto API (e.g., `http://localhost:3000`).
   - `API_KEY`: Your API key if required by the server.

### 🏃‍♂️ Running the Bot

**Run directly:**
```bash
go run main.go
```

**Or build and run:**
```bash
go build -o bot main.go
./bot
```

### 🐳 Using Docker

You can also run the bot using Docker and Docker Compose.

**Build and run with Docker Compose:**
```bash
docker-compose up -d
```

**Run with pure Docker:**
```bash
docker build -t urlshortener-bot .
docker run -d --env-file .env --name discord-bot urlshortener-bot
```

## 🧠 How it works

The bot uses the `discordgo` library to interact with the Discord API. Upon startup, it registers the slash commands globally for your application. When a command is triggered, it makes an HTTP request to the Shorto API (as seen in `scripts/curl_examples.sh`) and formats the response back to the Discord channel.

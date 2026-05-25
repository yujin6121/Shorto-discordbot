package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// Config holds environment variables
type Config struct {
	Token      string
	AppID      string
	APIBaseURL string
	APIKey     string
}

var config Config

func init() {
	_ = godotenv.Load()
	config = Config{
		Token:      os.Getenv("DISCORD_TOKEN"),
		AppID:      os.Getenv("APP_ID"),
		APIBaseURL: os.Getenv("API_BASE_URL"),
		APIKey:     os.Getenv("API_KEY"),
	}

	if config.Token == "" || config.AppID == "" {
		log.Fatal("DISCORD_TOKEN and APP_ID must be set")
	}
	if config.APIBaseURL == "" {
		config.APIBaseURL = "http://localhost:3000" // Default if not set
	}
}

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "shorten",
			Description: "Shorten a URL",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "The long URL to shorten",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "custom_code",
					Description: "Optional custom code for the short link",
					Required:    false,
				},
			},
		},
		{
			Name:        "domains",
			Description: "List available domains",
		},
		{
			Name:        "stats",
			Description: "Get service statistics",
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"shorten": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			longURL := optionMap["url"].StringValue()
			customCode := ""
			if opt, ok := optionMap["custom_code"]; ok {
				customCode = opt.StringValue()
			}

			respData, err := callShortenAPI(longURL, customCode)
			if err != nil {
				respondWithError(s, i, err)
				return
			}

			respondWithJSON(s, i, "URL Shortened!", respData)
		},
		"domains": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			respData, err := callGETAPI("/api/domains")
			if err != nil {
				respondWithError(s, i, err)
				return
			}
			respondWithJSON(s, i, "Available Domains", respData)
		},
		"stats": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			respData, err := callGETAPI("/api/stats")
			if err != nil {
				respondWithError(s, i, err)
				return
			}
			respondWithJSON(s, i, "Service Statistics", respData)
		},
	}
)

func main() {
	dg, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for index, v := range commands {
		cmd, err := dg.ApplicationCommandCreate(config.AppID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[index] = cmd
	}

	defer dg.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Bot is now running. Press CTRL-C to exit.")
	<-stop

	log.Println("Removing commands...")
	for _, v := range registeredCommands {
		err := dg.ApplicationCommandDelete(config.AppID, "", v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

	log.Println("Gracefully shutting down.")
}

func callShortenAPI(longURL, customCode string) (interface{}, error) {
	payload := map[string]string{"url": longURL}
	if customCode != "" {
		payload["customCode"] = customCode
	}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", config.APIBaseURL+"/api/shorten", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if config.APIKey != "" {
		req.Header.Set("x-api-key", config.APIKey)
	}

	return doRequest(req)
}

func callGETAPI(endpoint string) (interface{}, error) {
	req, err := http.NewRequest("GET", config.APIBaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	if config.APIKey != "" {
		req.Header.Set("x-api-key", config.APIKey)
	}

	return doRequest(req)
}

func doRequest(req *http.Request) (interface{}, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return string(body), nil
	}

	return result, nil
}

func respondWithError(s *discordgo.Session, i *discordgo.InteractionCreate, err error) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("❌ Error: %v", err),
		},
	})
}

func respondWithJSON(s *discordgo.Session, i *discordgo.InteractionCreate, title string, data interface{}) {
	jsonStr, _ := json.MarshalIndent(data, "", "  ")
	content := fmt.Sprintf("**%s**\n```json\n%s\n```", title, string(jsonStr))
	
	// Discord message limit is 2000 characters
	if len(content) > 1900 {
		content = content[:1900] + "\n... (truncated)" + "```"
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type weatherData struct {
	Weather []struct {
		Description string `json:"description"`
	} `json:"weather"`
	Main struct {
		Temp float32 `json:"temp"`
	} `json:"main"`
}

type memeData struct {
	URL string `json:"url"`
}

type giphyData struct {
	Data []struct {
		URL string `json:"url"`
	} `json:"data"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	dg, err := discordgo.New("Bot " + os.Getenv("TOKEN"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	defer dg.Close()

	fmt.Println("Bot is now running. Press CTRL-C to exit." )

	select {}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "!weather") {
		location := strings.TrimSpace(strings.TrimPrefix(m.Content, "!weather"))

		url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&appid=%s", location, os.Getenv("OPENWEATHERMAP_API_KEY"))
		resp, err := http.Get(url)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to get weather data.")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			s.ChannelMessageSend(m.ChannelID, "Failed to get weather data.")
			return
		}

		var data weatherData
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to parse weather data.")
			return
		}

		description := data.Weather[0].Description
		temp := data.Main.Temp

		response := fmt.Sprintf("Current weather in %s: %s, temperature: %.1f°C", location, description, temp)

		s.ChannelMessageSend(m.ChannelID, response)
	} else if strings.HasPrefix(m.Content, "!meme") {
		url, err := getMeme()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to get meme data.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, url)
	} else if strings.HasPrefix(m.Content, "!gif") {
		gifType := strings.TrimSpace(strings.TrimPrefix(m.Content, "!gif"))

		url, err := getGif(gifType)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "Failed to get GIF data.")
			return
		}

		s.ChannelMessageSend(m.ChannelID, url)
	}
}

func getMeme() (string, error) {
	resp, err := http.Get("https://meme-api.com/gimme")
	if err != nil {
		return "", fmt.Errorf("failed to get meme data: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get meme data: unexpected status code %d", resp.StatusCode)
	}

	var data memeData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", fmt.Errorf("failed to parse meme data: %s", err)
	}

	return data.URL, nil
}

func getGif(gifType string) (string, error) {
	apiKey := os.Getenv("GIPHY_API_KEY")
	url := fmt.Sprintf("https://api.giphy.com/v1/gifs/search?q=%s&api_key=%s&limit=25", gifType, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get GIF, status code: %d", resp.StatusCode)
	}

	var data giphyData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	if len(data.Data) == 0 {
		return "", fmt.Errorf("no GIFs found for type: %s", gifType)
	}

	gifURL := data.Data[0].URL
	return gifURL, nil
}
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Input struct {
	Params []struct {
		InputName string `json:"inputname"`
		CompValue string `json:"compvalue"`
	} `json:"params"`
}

type Output struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
}

func main() {
	var input Input
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("failed to decode input: %v", err)})
		return
	}

	var (
		token      string
		action     = "send_message"
		chatID     int64
		text       string
		parseMode  string
		offset     int
		webhookURL string
	)

	// Extract parameters
	for _, p := range input.Params {
		val := strings.TrimSpace(p.CompValue)
		switch strings.ToLower(p.InputName) {
		case "token":
			token = val
		case "action":
			if val != "" {
				action = strings.ToLower(val)
			}
		case "chat_id":
			fmt.Sscanf(val, "%d", &chatID)
		case "text":
			text = val
		case "parse_mode":
			parseMode = val
		case "offset":
			fmt.Sscanf(val, "%d", &offset)
		case "webhook_url":
			webhookURL = val
		}
	}

	// Validate required parameters
	if token == "" {
		json.NewEncoder(os.Stdout).Encode(Output{Error: "token is required"})
		return
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("failed to create bot: %v", err)})
		return
	}

	switch action {
	case "send_message":
		if chatID == 0 {
			json.NewEncoder(os.Stdout).Encode(Output{Error: "chat_id is required for send_message"})
			return
		}
		if text == "" {
			json.NewEncoder(os.Stdout).Encode(Output{Error: "text is required for send_message"})
			return
		}

		msg := tgbotapi.NewMessage(chatID, text)
		if parseMode != "" {
			msg.ParseMode = parseMode
		}

		sentMsg, err := bot.Send(msg)
		if err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("failed to send message: %v", err)})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Result: sentMsg})

	case "get_updates":
		u := tgbotapi.NewUpdate(offset)
		u.Timeout = 60

		updates, err := bot.GetUpdates(u)
		if err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("failed to get updates: %v", err)})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Result: updates})

	case "set_webhook":
		if webhookURL == "" {
			json.NewEncoder(os.Stdout).Encode(Output{Error: "webhook_url is required for set_webhook"})
			return
		}
		wh, err := tgbotapi.NewWebhook(webhookURL)
		if err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("failed to create webhook config: %v", err)})
			return
		}
		resp, err := bot.Request(wh)
		if err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("failed to set webhook: %v", err)})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Result: resp})

	case "delete_webhook":
		resp, err := bot.Request(tgbotapi.DeleteWebhookConfig{})
		if err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("failed to delete webhook: %v", err)})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Result: resp})

	case "get_webhook_info":
		info, err := bot.GetWebhookInfo()
		if err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("failed to get webhook info: %v", err)})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Result: info})

	case "get_me":
		user, err := bot.GetMe()
		if err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("failed to get bot info: %v", err)})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Result: user})

	default:
		json.NewEncoder(os.Stdout).Encode(Output{Error: "invalid action"})
	}
}

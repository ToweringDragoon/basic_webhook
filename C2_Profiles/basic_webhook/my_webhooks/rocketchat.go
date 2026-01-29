package my_webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MythicMeta/MythicContainer/webhookstructs"
)

// RocketChat structures
type RocketChatAttachment struct {
	Title     string `json:"title,omitempty"`
	TitleLink string `json:"title_link,omitempty"`
	Text      string `json:"text,omitempty"`
	Color     string `json:"color,omitempty"`
}

type RocketChatPayload struct {
	Text        string                 `json:"text,omitempty"`
	Channel     string                 `json:"channel,omitempty"`
	Attachments []RocketChatAttachment `json:"attachments,omitempty"`
}

func sendRocketChatMessage(webhookURL string, msg webhookstructs.SlackWebhookMessage) error {
	var attachments []RocketChatAttachment

	for _, att := range msg.Attachments {
		rcAttachment := RocketChatAttachment{
			Title: att.Title,
			Color: att.Color, // RocketChat uses hex colors like "#00FF00"
		}

		// Convert Slack blocks to text for RocketChat
		var textParts []string
		if att.Blocks != nil {
			for _, block := range *att.Blocks {
				// Handle text blocks
				if block.Text != nil && block.Text.Text != "" {
					textParts = append(textParts, block.Text.Text)
				}

				// Handle field blocks
				if block.Fields != nil {
					for _, f := range *block.Fields {
						if f.Text != "" {
							textParts = append(textParts, f.Text)
						}
					}
				}
			}
		}

		// Join all text parts with newlines
		if len(textParts) > 0 {
			rcAttachment.Text = ""
			for i, part := range textParts {
				if i > 0 {
					rcAttachment.Text += "\n"
				}
				rcAttachment.Text += part
			}
		}

		attachments = append(attachments, rcAttachment)
	}

	payload := RocketChatPayload{
		Attachments: attachments,
	}

	// Set channel if provided
	if msg.Channel != "" {
		payload.Channel = msg.Channel
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("rocketchat webhook returned status %s", resp.Status)
	}

	return nil
}

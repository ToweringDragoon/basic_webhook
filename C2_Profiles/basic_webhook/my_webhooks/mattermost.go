package my_webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MythicMeta/MythicContainer/webhookstructs"
)

// Mattermost structures (Slack-compatible format)
type MattermostAttachmentField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type MattermostAttachment struct {
	Fallback string                      `json:"fallback,omitempty"`
	Color    string                      `json:"color,omitempty"`
	Pretext  string                      `json:"pretext,omitempty"`
	Text     string                      `json:"text,omitempty"`
	Title    string                      `json:"title,omitempty"`
	Fields   []MattermostAttachmentField `json:"fields,omitempty"`
}

type MattermostPayload struct {
	Channel     string                 `json:"channel,omitempty"`
	Username    string                 `json:"username,omitempty"`
	Text        string                 `json:"text,omitempty"`
	Attachments []MattermostAttachment `json:"attachments,omitempty"`
}

func sendMattermostMessage(webhookURL string, msg webhookstructs.SlackWebhookMessage) error {
	var attachments []MattermostAttachment

	for _, att := range msg.Attachments {
		mmAttachment := MattermostAttachment{
			Title:    att.Title,
			Color:    att.Color, // Mattermost uses hex colors like "#00FF00"
			Fallback: att.Title, // Required field - use title as fallback
		}

		// Convert Slack blocks to Mattermost format
		var textParts []string
		var fields []MattermostAttachmentField

		if att.Blocks != nil {
			for _, block := range *att.Blocks {
				// Handle text blocks - add to main text
				if block.Text != nil && block.Text.Text != "" {
					textParts = append(textParts, block.Text.Text)
				}

				// Handle field blocks - convert to Mattermost fields
				if block.Fields != nil {
					for _, f := range *block.Fields {
						if f.Text != "" {
							fields = append(fields, MattermostAttachmentField{
								Title: "",
								Value: f.Text,
								Short: true,
							})
						}
					}
				}
			}
		}

		// Set text from collected parts
		if len(textParts) > 0 {
			for i, part := range textParts {
				if i > 0 {
					mmAttachment.Text += "\n"
				}
				mmAttachment.Text += part
			}
		}

		// Set fields if any
		if len(fields) > 0 {
			mmAttachment.Fields = fields
		}

		attachments = append(attachments, mmAttachment)
	}

	payload := MattermostPayload{
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
		return fmt.Errorf("mattermost webhook returned status %s", resp.Status)
	}

	return nil
}

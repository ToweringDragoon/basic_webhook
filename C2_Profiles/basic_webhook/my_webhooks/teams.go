package my_webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/MythicMeta/MythicContainer/webhookstructs"
)

// Teams Adaptive Card structures for Power Automate Workflows
// Uses the standard Adaptive Card format required by the "When a Teams webhook request is received" trigger
type TeamsAdaptiveCard struct {
	Type    string                 `json:"type"`
	Schema  string                 `json:"$schema"`
	Version string                 `json:"version"`
	Body    []TeamsAdaptiveElement `json:"body"`
}

type TeamsAdaptiveElement struct {
	Type      string              `json:"type"`
	Text      string              `json:"text,omitempty"`
	Weight    string              `json:"weight,omitempty"`
	Size      string              `json:"size,omitempty"`
	Color     string              `json:"color,omitempty"`
	Wrap      bool                `json:"wrap,omitempty"`
	Separator bool                `json:"separator,omitempty"`
	Facts     []TeamsAdaptiveFact `json:"facts,omitempty"`
}

type TeamsAdaptiveFact struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

type TeamsAttachment struct {
	ContentType string            `json:"contentType"`
	ContentUrl  *string           `json:"contentUrl"`
	Content     TeamsAdaptiveCard `json:"content"`
}

type TeamsPayload struct {
	Type        string            `json:"type"`
	Attachments []TeamsAttachment `json:"attachments"`
}

func sendTeamsMessage(webhookURL string, msg webhookstructs.SlackWebhookMessage) error {
	var bodyElements []TeamsAdaptiveElement

	for _, att := range msg.Attachments {
		// Add title as header
		if att.Title != "" {
			bodyElements = append(bodyElements, TeamsAdaptiveElement{
				Type:   "TextBlock",
				Text:   att.Title,
				Weight: "Bolder",
				Size:   "Medium",
				Color:  slackColorToTeamsColor(att.Color),
			})
		}

		// Convert Slack blocks
		if att.Blocks != nil {
			var facts []TeamsAdaptiveFact

			for _, block := range *att.Blocks {
				// Handle text blocks
				if block.Text != nil && block.Text.Text != "" {
					bodyElements = append(bodyElements, TeamsAdaptiveElement{
						Type: "TextBlock",
						Text: cleanMarkdown(block.Text.Text),
						Wrap: true,
					})
				}

				// Handle field blocks - collect as facts
				if block.Fields != nil {
					for _, f := range *block.Fields {
						if f.Text != "" {
							// Parse "**Label**\nValue" format
							title, value := parseSlackField(f.Text)
							facts = append(facts, TeamsAdaptiveFact{
								Title: title,
								Value: value,
							})
						}
					}
				}

				// Handle dividers
				if block.Type == "divider" {
					bodyElements = append(bodyElements, TeamsAdaptiveElement{
						Type:      "TextBlock",
						Text:      "",
						Separator: true,
					})
				}
			}

			// Add collected facts as a FactSet
			if len(facts) > 0 {
				bodyElements = append(bodyElements, TeamsAdaptiveElement{
					Type:  "FactSet",
					Facts: facts,
				})
			}
		}
	}

	// Build Adaptive Card payload for Power Automate Workflows
	adaptiveCard := TeamsAdaptiveCard{
		Type:    "AdaptiveCard",
		Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
		Version: "1.3",
		Body:    bodyElements,
	}

	payload := TeamsPayload{
		Type: "message",
		Attachments: []TeamsAttachment{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				ContentUrl:  nil,
				Content:     adaptiveCard,
			},
		},
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
		return fmt.Errorf("teams webhook returned status %s", resp.Status)
	}

	return nil
}

// slackColorToTeamsColor converts hex colors to Teams Adaptive Card color names
func slackColorToTeamsColor(hexColor string) string {
	// Teams Adaptive Cards support: default, dark, light, accent, good, warning, attention
	switch strings.ToLower(hexColor) {
	case "#ff0000", "#ee0000": // red
		return "Attention"
	case "#00ff00", "#00ee00", "#85b089": // green
		return "Good"
	case "#ffff00", "#ffa500": // yellow/orange
		return "Warning"
	case "#b366ff": // purple (callback)
		return "Accent"
	case "#84b4dc": // blue (custom)
		return "Accent"
	default:
		return "Default"
	}
}

// parseSlackField extracts title and value from Slack's "*Title*\nValue" format
func parseSlackField(text string) (string, string) {
	// Handle "**Label**\nValue" or "*Label*\nValue" format
	text = strings.ReplaceAll(text, "**", "*")
	parts := strings.SplitN(text, "\n", 2)
	if len(parts) == 2 {
		title := strings.Trim(parts[0], "*")
		return title, parts[1]
	}
	return "", text
}

// cleanMarkdown removes Slack-specific markdown that Teams doesn't support
func cleanMarkdown(text string) string {
	// Teams Adaptive Cards support basic markdown
	// Convert Slack's *bold* to **bold** for Teams
	return text
}

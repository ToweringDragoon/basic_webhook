package my_webhooks

import (
	"errors"
	"os"
	"strings"

	"github.com/MythicMeta/MythicContainer/webhookstructs"
)

// sendMessage routes webhook messages to the appropriate platform handler.
// Platform is determined by WEBHOOK_DEFAULT_PLATFORM env var, or auto-detected from URL if not set.
// Supported platforms: slack, discord, mattermost, rocketchat, teams
func sendMessage(webhookURL string, newMessage webhookstructs.SlackWebhookMessage) error {
	if webhookURL == "" {
		return errors.New("no basic_webhook URL provided")
	}

	// Check for explicit platform configuration first
	platform := strings.ToLower(os.Getenv("WEBHOOK_DEFAULT_PLATFORM"))
	if platform != "" {
		return sendToPlatform(platform, webhookURL, newMessage)
	}

	// Fall back to auto-detection from URL
	return autoDetectAndSend(webhookURL, newMessage)
}

// sendToPlatform sends the message to the specified platform
func sendToPlatform(platform, webhookURL string, newMessage webhookstructs.SlackWebhookMessage) error {
	switch platform {
	case "discord":
		return sendDiscordMessage(webhookURL, newMessage)
	case "slack":
		_, _, err := webhookstructs.SubmitWebRequest("POST", webhookURL, newMessage)
		return err
	case "mattermost":
		return sendMattermostMessage(webhookURL, newMessage)
	case "rocketchat":
		return sendRocketChatMessage(webhookURL, newMessage)
	case "teams":
		return sendTeamsMessage(webhookURL, newMessage)
	default:
		return errors.New("unsupported WEBHOOK_DEFAULT_PLATFORM: " + platform + " (supported: slack, discord, mattermost, rocketchat, teams)")
	}
}

// autoDetectAndSend attempts to detect the platform from the webhook URL
func autoDetectAndSend(webhookURL string, newMessage webhookstructs.SlackWebhookMessage) error {
	// Detect Discord
	if strings.Contains(webhookURL, "discord.com") || strings.Contains(webhookURL, "discordapp.com") {
		return sendDiscordMessage(webhookURL, newMessage)
	}

	// Detect Slack
	if strings.Contains(webhookURL, "slack.com") {
		_, _, err := webhookstructs.SubmitWebRequest("POST", webhookURL, newMessage)
		return err
	}

	// Detect Microsoft Teams
	if strings.Contains(webhookURL, "office.com") || strings.Contains(webhookURL, "microsoft.com") {
		return sendTeamsMessage(webhookURL, newMessage)
	}

	// Detect Mattermost (common hosted patterns)
	if strings.Contains(webhookURL, "mattermost") {
		return sendMattermostMessage(webhookURL, newMessage)
	}

	// For other URLs, require explicit platform configuration
	return errors.New("could not auto-detect webhook platform from URL. Set WEBHOOK_DEFAULT_PLATFORM env var to: slack, discord, mattermost, rocketchat, or teams")
}

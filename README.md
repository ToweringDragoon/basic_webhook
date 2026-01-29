# basic_webhook

This is a basic webhook container for Mythic v3.0.0. This container connects up to RabbitMQ queues for webhooks and posts those messages and additional data to your team chat platform.

## Supported Platforms

- **Slack** - Native webhook format
- **Discord** - Converted to Discord embeds
- **Microsoft Teams** - Adaptive Cards via Power Automate Workflows
- **Mattermost** - Slack-compatible attachments with fields
- **RocketChat** - Attachments with title/text/color

## Installation

Within Mythic you can run the `mythic-cli` binary to install this in one of three ways:

* `sudo ./mythic-cli install github https://github.com/user/repo` to install the main branch
* `sudo ./mythic-cli install github https://github.com/user/repo branchname` to install a specific branch of that repo
* `sudo ./mythic-cli install folder /path/to/local/folder/cloned/from/github` to install from an already cloned down version of an agent repo

Now, you might be wondering _when_ should you or a user do this to properly add your profile to their Mythic instance. There's no wrong answer here, just depends on your preference. The three options are:

* Mythic is already up and going, then you can run the install script and just direct that profile's containers to start (i.e. `sudo ./mythic-cli start profileName`.
* Mythic is already up and going, but you want to minimize your steps, you can just install the profile and run `sudo ./mythic-cli start`. That script will first _stop_ all of your containers, then start everything back up again. This will also bring in the new profile you just installed.
* Mythic isn't running, you can install the script and just run `sudo ./mythic-cli start`.

## Configuration

### Setting the Webhook URL

You have three options for configuring the webhook URL:

1. **Per-operation (UI)**: Click the operation name at the top of Mythic's UI and click edit for your current operation. You can add the webhook URL there.

2. **Environment variables (recommended)**: Set the following in your `.env` file, then run `sudo ./mythic-cli start basic_webhook`:
   - `WEBHOOK_DEFAULT_URL` - Your webhook endpoint URL
   - `WEBHOOK_DEFAULT_PLATFORM` - Platform type (see below)
   - `WEBHOOK_DEFAULT_CHANNEL` - Default channel for notifications
   - `WEBHOOK_DEFAULT_CALLBACK_CHANNEL` - Channel for new callbacks
   - `WEBHOOK_DEFAULT_FEEDBACK_CHANNEL` - Channel for feedback
   - `WEBHOOK_DEFAULT_STARTUP_CHANNEL` - Channel for startup notifications

3. **Hardcoded**: Edit the code directly (requires rebuild with `sudo ./mythic-cli config set basic_webhook_use_build_context true` and `sudo ./mythic-cli config set basic_webhook_use_volume false` then `sudo ./mythic-cli build basic_webhook`)

### Platform Selection

Set `WEBHOOK_DEFAULT_PLATFORM` to specify your platform. This is **required for self-hosted instances** where auto-detection won't work.

| Platform | Value | Auto-detected URLs |
|----------|-------|-------------------|
| Slack | `slack` | `slack.com` |
| Discord | `discord` | `discord.com`, `discordapp.com` |
| Microsoft Teams | `teams` | `office.com`, `microsoft.com` |
| Mattermost | `mattermost` | `mattermost` in URL, or `/hooks/{key}` (1 path segment) |
| RocketChat | `rocketchat` | `/hooks/{id}/{token}` (2 path segments) |

## Platform Setup Instructions

### Slack

1. Go to your Slack workspace settings → Apps → Manage → Build
2. Create a new app → From scratch
3. Enable Incoming Webhooks
4. Add a new webhook to your workspace and select a channel
5. Copy the webhook URL (format: `https://hooks.slack.com/services/T.../B.../...`)

```
WEBHOOK_DEFAULT_PLATFORM=slack
WEBHOOK_DEFAULT_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

### Discord

1. Open your Discord server settings → Integrations → Webhooks
2. Create a new webhook and select a channel
3. Copy the webhook URL (format: `https://discord.com/api/webhooks/.../...`)

```
WEBHOOK_DEFAULT_PLATFORM=discord
WEBHOOK_DEFAULT_URL=https://discord.com/api/webhooks/YOUR/WEBHOOK/URL
```

### Microsoft Teams (Power Automate Workflows)

Microsoft deprecated the old connector webhooks. Use Power Automate Workflows instead:

1. In Teams, go to your channel → `...` menu → Workflows
2. Select "Post to a channel when a webhook request is received"
3. Configure the workflow and select your channel
4. Copy the generated webhook URL

Or create a custom flow:
1. Go to Power Automate (make.powerautomate.com)
2. Create → Instant cloud flow
3. Use trigger: "When a Teams webhook request is received"
4. Add action: "Post card in a chat or channel"
5. Save and copy the HTTP POST URL from the trigger

```
WEBHOOK_DEFAULT_PLATFORM=teams
WEBHOOK_DEFAULT_URL=https://prod-XX.westus.logic.azure.com:443/workflows/...
```

### Mattermost

1. Go to your Mattermost server → Integrations → Incoming Webhooks
2. Add an incoming webhook and select a channel
3. Copy the webhook URL (format: `https://your-server.com/hooks/...`)

```
WEBHOOK_DEFAULT_PLATFORM=mattermost
WEBHOOK_DEFAULT_URL=https://mattermost.yourcompany.com/hooks/YOUR_HOOK_ID
```

### RocketChat

1. Go to Administration → Integrations → Incoming
2. Create a new incoming webhook
3. Configure the channel and options
4. Save and copy the webhook URL (format: `https://your-server.com/hooks/.../...`)

```
WEBHOOK_DEFAULT_PLATFORM=rocketchat
WEBHOOK_DEFAULT_URL=https://rocket.yourcompany.com/hooks/TOKEN/HOOK_ID
```

## Event Types

The webhook sends notifications for:
- **New Callback** - When a new agent checks in
- **Alert** - System alerts (throttled to prevent spam)
- **Feedback** - Operator feedback (bugs, feature requests, etc.)
- **Startup** - Service startup notifications
- **Custom** - Custom webhook messages

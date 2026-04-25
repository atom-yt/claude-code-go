# Gateway Configuration

This document describes how to configure interface providers for the gateway.

## Quick Start

1. Create or edit `~/.claude/settings.json` or `.claude/settings.json`
2. Add your provider configuration
3. Run `gateway start`

## Example Configuration

```json
{
  "model": "claude-sonnet-4-6",
  "provider": "anthropic",
  "apiKey": "sk-ant-xxx",
  "interfaces": {
    "feishu": {
      "enabled": true,
      "config": {
        "app_id": "cli_xxxxxxxxx",
        "app_secret": "xxxxxxxxxxxxxxxx",
        "mode": "dual",
        "websocket_url": "wss://open.feishu.cn/open-apis/bot/v4/ws",
        "webhook_port": 8080,
        "webhook_path": "/webhook",
        "webhook_secret": "your-secret-key",
        "max_sessions": 100,
        "session_timeout": "1h",
        "persist_sessions": true,
        "rate_limit_requests": 60,
        "rate_limit_burst": 10,
        "model": "claude-sonnet-4-6",
        "system_prompt": "You are a helpful AI assistant.",
        "max_history_size": 50,
        "enable_markdown": true,
        "enable_cards": true,
        "enable_images": true,
        "enable_reactions": false,
        "output_format": "auto",
        "truncate_length": 2000
      }
    }
  }
}
```

## Configuration Options

### Provider Configuration

| Option | Type | Description | Default |
|---------|------|-------------|----------|
| `enabled` | boolean | Whether this provider is enabled | `false` |
| `config` | object | Provider-specific configuration | - |

### Feishu Configuration

#### Credentials

| Option | Type | Description | Required |
|---------|------|-------------|-----------|
| `app_id` | string | Feishu app ID | Yes |
| `app_secret` | string | Feishu app secret | Yes |
| `encrypt_key` | string | Encryption key (if enabled) | No |
| `verification_token` | string | URL verification token | No |

#### Connection Mode

| Option | Type | Description | Default |
|---------|------|-------------|----------|
| `mode` | string | Connection mode: `websocket`, `webhook`, or `dual` | `dual` |
| `websocket_url` | string | WebSocket URL | `wss://open.feishu.cn/open-apis/bot/v4/ws` |
| `webhook_port` | int | Webhook server port | `8080` |
| `webhook_path` | string | Webhook URL path | `/webhook` |
| `webhook_url` | string | Public webhook URL (for Feishu) | - |
| `webhook_secret` | string | Secret for webhook signature | - |

#### Session Management

| Option | Type | Description | Default |
|---------|------|-------------|----------|
| `max_sessions` | int | Maximum concurrent sessions | `100` |
| `session_timeout` | string | Session timeout duration | `1h` |
| `persist_sessions` | boolean | Persist session history to disk | `true` |

#### Rate Limiting

| Option | Type | Description | Default |
|---------|------|-------------|----------|
| `rate_limit_requests` | int | Requests per second | `60` |
| `rate_limit_burst` | int | Burst capacity | `10` |

#### Agent Settings

| Option | Type | Description | Default |
|---------|------|-------------|----------|
| `model` | string | Model to use (overrides global) | - |
| `system_prompt` | string | System prompt for all sessions | - |
| `max_history_size` | int | Maximum messages per session history | `50` |

#### Features

| Option | Type | Description | Default |
|---------|------|-------------|----------|
| `enable_markdown` | boolean | Enable markdown formatting | `true` |
| `enable_cards` | boolean | Enable interactive cards | `true` |
| `enable_images` | boolean | Enable image/file support | `true` |
| `enable_reactions` | boolean | Enable reaction support | `false` |

#### Output Formatting

| Option | Type | Description | Default |
|---------|------|-------------|----------|
| `output_format` | string | Format: `auto`, `text`, `markdown`, `card` | `auto` |
| `truncate_length` | int | Maximum message length | `2000` |

## Commands

### List Providers

```bash
./gateway list
```

Shows all registered interface providers and their enabled status.

### Start Provider

```bash
# Start all enabled providers
./gateway start

# Start specific provider(s)
./gateway start feishu

# Start multiple providers
./gateway start feishu wechat
```

## Environment Variables

You can also configure providers via environment variables:

| Variable | Description |
|----------|-------------|
| `FEISHU_APP_ID` | Feishu app ID |
| `FEISHU_APP_SECRET` | Feishu app secret |
| `FEISHU_ENCRYPT_KEY` | Encryption key |
| `FEISHU_VERIFICATION_TOKEN` | Verification token |

Environment variables override settings.json values.

## Adding New Providers

To add a new interface provider:

1. Create a new package in `internal/interfaces/<provider>/`
2. Implement the `interfaces.Provider` interface
3. Register the provider in `init()`:

```go
package myprovider

import "github.com/atom-yt/claude-code-go/internal/interfaces"

type Provider struct{}

func (p *Provider) Name() string {
    return "myprovider"
}

func (p *Provider) Description() string {
    return "My Provider Description"
}

func (p *Provider) Start(ctx context.Context, config interfaces.ProviderConfig) error {
    // Implementation
    return nil
}

// ... other methods ...

func init() {
    interfaces.Register(&Provider{})
}
```

4. Import the package in `cmd/gateway/main.go` to trigger registration:

```go
import _ "github.com/atom-yt/claude-code-go/internal/interfaces/myprovider"
```
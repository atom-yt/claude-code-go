// Package feishu implements the Feishu/Lark interface provider.
//
// Feishu (飞书) is an enterprise collaboration platform with rich messaging
// capabilities including text, images, cards, and interactive elements.
//
// This provider supports:
// - WebSocket mode: Real-time message push from Feishu
// - Webhook mode: HTTP callback for event delivery
// - Dual mode: Both WebSocket and Webhook for redundancy
//
// Usage:
//
//	// Provider is automatically registered via init()
//	// To start the Feishu provider:
//
//	provider, _ := interfaces.Get("feishu")
//	config := interfaces.ProviderConfig{
//	    Model:        "claude-sonnet-4-6",
//	    ToolRegistry: tools.NewRegistry(),
//	    Settings: map[string]any{
//	        "app_id":    "cli_xxxxxxxxx",
//	        "app_secret": "xxxxxxxx",
//	        "mode":      "dual",
//	    },
//	}
//	provider.Start(ctx, config)
package feishu
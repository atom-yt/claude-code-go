// Package interfaces provides a registry and framework for implementing
// input/output interface providers (e.g., Feishu, WeChat, Telegram).
//
// An InterfaceProvider handles:
// - Receiving messages from an external platform
// - Formatting and sending responses to the platform
// - Managing platform-specific features (cards, reactions, etc.)
//
// To add a new interface provider:
// 1. Create a subpackage (e.g., internal/interfaces/wechat)
// 2. Implement the Provider interface
// 3. Register the provider in internal/interfaces/registry.go
//
// Example:
//
//	package wechat
//
//	import "github.com/example/claude-code-go/internal/interfaces"
//
//	func init() {
//	    interfaces.Register(&WeChatProvider{})
//	}
package interfaces
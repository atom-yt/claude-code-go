package api

import "context"

// Streamer is the interface implemented by all API clients.
// The agent depends on this interface, not on a concrete client.
type Streamer interface {
	StreamMessages(ctx context.Context, req MessagesRequest) <-chan APIEvent
}

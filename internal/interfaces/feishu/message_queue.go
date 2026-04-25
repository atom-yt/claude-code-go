package feishu

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"
)

// Priority represents message priority.
type Priority int

const (
	// PriorityLow is for background messages.
	PriorityLow Priority = iota

	// PriorityNormal is for regular messages.
	PriorityNormal

	// PriorityHigh is for important messages.
	PriorityHigh

	// PriorityUrgent is for urgent messages (e.g., errors).
	PriorityUrgent
)

// QueuedMessage represents a message waiting to be processed.
type QueuedMessage struct {
	Message    *IncomingMessage
	Timestamp  time.Time
	Priority   Priority
	Index      int // For heap implementation
}

// MessageQueue is a priority queue for processing messages.
type MessageQueue struct {
	mu         sync.RWMutex
	heap       *messageHeap
	queue      map[string]*QueuedMessage // For quick lookups
	rateLimiter *RateLimiter
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	// Channels
	inputCh    chan *IncomingMessage
	outputCh   chan *IncomingMessage
	errorCh    chan error

	// Statistics
	stats      QueueStats
}

// QueueStats contains queue statistics.
type QueueStats struct {
	TotalEnqueued int64
	TotalProcessed int64
	TotalFailed   int64
	CurrentSize   int
	MaxSize       int
}

// messageHeap implements heap.Interface for []*QueuedMessage.
type messageHeap []*QueuedMessage

func (h messageHeap) Len() int { return len(h) }

func (h messageHeap) Less(i, j int) bool {
	// Higher priority (lower number) comes first
	if h[i].Priority != h[j].Priority {
		return h[i].Priority < h[j].Priority
	}
	// Within same priority, older messages come first
	return h[i].Timestamp.Before(h[j].Timestamp)
}

func (h messageHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

func (h *messageHeap) Push(x any) {
	n := len(*h)
	item := x.(*QueuedMessage)
	item.Index = n
	*h = append(*h, item)
}

func (h *messageHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.Index = -1
	*h = old[0 : n-1]
	return item
}

// NewMessageQueue creates a new message queue.
func NewMessageQueue(ctx context.Context, rps, burst int) *MessageQueue {
	ctx, cancel := context.WithCancel(ctx)

	mq := &MessageQueue{
		heap:        &messageHeap{},
		queue:       make(map[string]*QueuedMessage),
		rateLimiter: NewRateLimiter(rps, burst),
		ctx:        ctx,
		cancel:      cancel,
		inputCh:     make(chan *IncomingMessage, 100),
		outputCh:    make(chan *IncomingMessage, 100),
		errorCh:     make(chan error, 10),
	}

	heap.Init(mq.heap)

	// Start processing goroutine
	mq.wg.Add(1)
	go mq.processLoop()

	return mq
}

// Enqueue adds a message to the queue.
func (mq *MessageQueue) Enqueue(msg *IncomingMessage, priority Priority) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	// Check if message already queued
	key := mq.messageKey(msg)
	if _, exists := mq.queue[key]; exists {
		return
	}

	queued := &QueuedMessage{
		Message:   msg,
		Timestamp: time.Now(),
		Priority:  priority,
	}

	// Add to heap and map
	heap.Push(mq.heap, queued)
	mq.queue[key] = queued

	// Update stats
	mq.stats.TotalEnqueued++
	mq.stats.CurrentSize++
	if mq.stats.CurrentSize > mq.stats.MaxSize {
		mq.stats.MaxSize = mq.stats.CurrentSize
	}

	// Signal input channel
	select {
	case mq.inputCh <- msg:
	default:
		// Channel full, message is still in heap
	}
}

// Dequeue removes and returns the next message.
func (mq *MessageQueue) Dequeue(ctx context.Context) (*IncomingMessage, error) {
	select {
	case msg := <-mq.outputCh:
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// WaitOnRateLimit waits until rate limiter allows a message.
func (mq *MessageQueue) WaitOnRateLimit(ctx context.Context, userID string) error {
	return mq.rateLimiter.WaitUser(ctx, userID)
}

// processLoop processes queued messages.
func (mq *MessageQueue) processLoop() {
	defer mq.wg.Done()

	for {
		select {
		case <-mq.ctx.Done():
			return

		case <-mq.inputCh:
			// Message added, process if rate limit allows
			mq.processNext()

		case <-time.After(100 * time.Millisecond):
			// Periodic check for messages
			mq.processNext()
		}
	}
}

// processNext processes the next message in the queue.
func (mq *MessageQueue) processNext() {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if mq.heap.Len() == 0 {
		return
	}

	// Check rate limit
	next := (*mq.heap)[0]
	if !mq.rateLimiter.AllowUser(next.Message.UserID) {
		// Rate limited, don't process now
		return
	}

	// Dequeue message
	queued := heap.Pop(mq.heap).(*QueuedMessage)
	key := mq.messageKey(queued.Message)
	delete(mq.queue, key)

	// Update stats
	mq.stats.CurrentSize--

	// Send to output channel
	select {
	case mq.outputCh <- queued.Message:
		mq.stats.TotalProcessed++
	default:
		mq.errorCh <- ErrQueueFull
		mq.stats.TotalFailed++
	}
}

// Remove removes a message from the queue.
func (mq *MessageQueue) Remove(msg *IncomingMessage) bool {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	key := mq.messageKey(msg)
	queued, exists := mq.queue[key]
	if !exists {
		return false
	}

	// Remove from heap
	heap.Remove(mq.heap, queued.Index)
	delete(mq.queue, key)

	// Update stats
	mq.stats.CurrentSize--

	return true
}

// Size returns the current queue size.
func (mq *MessageQueue) Size() int {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	return mq.heap.Len()
}

// Stats returns the queue statistics.
func (mq *MessageQueue) Stats() QueueStats {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	return mq.stats
}

// Errors returns a channel for error notifications.
func (mq *MessageQueue) Errors() <-chan error {
	return mq.errorCh
}

// Clear clears the queue.
func (mq *MessageQueue) Clear() {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.heap = &messageHeap{}
	mq.queue = make(map[string]*QueuedMessage)
	mq.stats.CurrentSize = 0
	heap.Init(mq.heap)
}

// Close closes the queue.
func (mq *MessageQueue) Close() {
	mq.cancel()
	mq.wg.Wait()

	close(mq.inputCh)
	close(mq.outputCh)
	close(mq.errorCh)
}

// messageKey generates a unique key for a message.
func (mq *MessageQueue) messageKey(msg *IncomingMessage) string {
	return msg.ChatID + ":" + msg.MessageID
}

// SetPriority updates the priority of a queued message.
func (mq *MessageQueue) SetPriority(msg *IncomingMessage, priority Priority) bool {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	key := mq.messageKey(msg)
	queued, exists := mq.queue[key]
	if !exists {
		return false
	}

	queued.Priority = priority
	heap.Fix(mq.heap, queued.Index)

	return true
}

// CleanupUser removes all messages from a specific user.
func (mq *MessageQueue) CleanupUser(userID string) int {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	var removed int

	for key, queued := range mq.queue {
		if queued.Message.UserID == userID {
			heap.Remove(mq.heap, queued.Index)
			delete(mq.queue, key)
			removed++
		}
	}

	mq.stats.CurrentSize -= removed
	mq.rateLimiter.ClearUser(userID)

	return removed
}

// Drain processes all queued messages synchronously.
func (mq *MessageQueue) Drain(ctx context.Context) int {
	drained := 0

	for mq.Size() > 0 {
		select {
		case <-mq.outputCh:
			drained++
		case <-ctx.Done():
			return drained
		}
	}

	return drained
}

// Queue errors
var (
	ErrQueueFull = fmt.Errorf("message queue is full")
)
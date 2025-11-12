package slack

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/slack-go/slack"
)

type Connector struct {
	client      SlackClientInterface
	config      *Config
	rateLimit   *RateLimitInfo
	rateLimitMu sync.RWMutex
	status      *SyncStatus
	statusMu    sync.RWMutex
}

type Config struct {
	Token        string        `json:"token"`
	Channels     []string      `json:"channels"`      // Channels to index
	SyncInterval time.Duration `json:"sync_interval"` // How often to sync
	MaxMessages  int           `json:"max_messages"`  // Max messages per channel
}

type Message struct {
	ID         string    `json:"id"`
	Channel    string    `json:"channel"`
	User       string    `json:"user"`
	Text       string    `json:"text"`
	Timestamp  string    `json:"timestamp"`
	ThreadTS   string    `json:"thread_ts,omitempty"`
	ReplyCount int       `json:"reply_count"`
	CreatedAt  time.Time `json:"created_at"`
}

type Channel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsPrivate bool   `json:"is_private"`
	Members   int    `json:"members"`
}

type Thread struct {
	ParentMessage Message   `json:"parent_message"`
	Replies       []Message `json:"replies"`
}

type RateLimitInfo struct {
	Remaining int       `json:"remaining"`
	Reset     time.Time `json:"reset"`
}

type SyncStatus struct {
	LastSync       time.Time      `json:"last_sync"`
	TotalMessages  int            `json:"total_messages"`
	TotalChannels  int            `json:"total_channels"`
	SyncInProgress bool           `json:"sync_in_progress"`
	Error          string         `json:"error,omitempty"`
	RateLimit      *RateLimitInfo `json:"rate_limit,omitempty"`
}

func NewConnector(config *Config) (*Connector, error) {
	if config.Token == "" {
		return nil, fmt.Errorf("Slack token is required")
	}

	if len(config.Channels) == 0 {
		return nil, fmt.Errorf("at least one Slack channel is required")
	}

	if config.SyncInterval == 0 {
		config.SyncInterval = 5 * time.Minute // Default sync interval
	}

	if config.MaxMessages == 0 {
		config.MaxMessages = 1000 // Default max messages per channel
	}

	slackClient := slack.New(config.Token)
	client := NewRealSlackClient(slackClient)

	connector := &Connector{
		client:    client,
		config:    config,
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	return connector, nil
}

// SyncMessages syncs messages from configured channels
func (sc *Connector) SyncMessages(ctx context.Context) ([]Message, error) {
	// Update sync status
	sc.statusMu.Lock()
	sc.status.SyncInProgress = true
	sc.statusMu.Unlock()

	var allMessages []Message
	totalChannels := len(sc.config.Channels)

	for _, channelID := range sc.config.Channels {
		messages, err := sc.getChannelHistory(ctx, channelID)
		if err != nil {
			log.Printf("Warning: Failed to sync channel %s: %v", channelID, err)
			sc.updateSyncStatus(0, 0, err)
			continue
		}

		allMessages = append(allMessages, messages...)
	}

	sc.updateSyncStatus(len(allMessages), totalChannels, nil)
	return allMessages, nil
}

// getChannelHistory retrieves message history for a specific channel
func (sc *Connector) getChannelHistory(ctx context.Context, channelID string) ([]Message, error) {
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     200, // Max per request
	}

	var allMessages []Message
	messageCount := 0

	for {
		if messageCount >= sc.config.MaxMessages {
			break
		}

		history, err := sc.client.GetConversationHistory(params)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation history: %w", err)
		}

		for _, msg := range history.Messages {
			// Skip bot messages and system messages
			if msg.BotID != "" || msg.SubType == "channel_join" || msg.SubType == "channel_leave" {
				continue
			}

			createdAt, _ := parseSlackTimestamp(msg.Timestamp)

			message := Message{
				ID:         msg.ClientMsgID,
				Channel:    channelID,
				User:       msg.User,
				Text:       msg.Text,
				Timestamp:  msg.Timestamp,
				ThreadTS:   msg.ThreadTimestamp,
				ReplyCount: msg.ReplyCount,
				CreatedAt:  createdAt,
			}

			allMessages = append(allMessages, message)
			messageCount++

			if messageCount >= sc.config.MaxMessages {
				break
			}
		}

		if !history.HasMore {
			break
		}

		// Get next cursor if available
		if history.ResponseMetaData.NextCursor != "" {
			params.Cursor = history.ResponseMetaData.NextCursor
		} else {
			break
		}
	}

	return allMessages, nil
}

// SearchMessages searches for messages across channels
func (sc *Connector) SearchMessages(ctx context.Context, query string) ([]Message, error) {
	params := slack.SearchParameters{
		Count: 100,
		Page:  1,
	}

	searchResult, err := sc.client.SearchMessages(query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	var messages []Message
	for _, match := range searchResult.Matches {
		createdAt, _ := parseSlackTimestamp(match.Timestamp)

		message := Message{
			ID:        match.Timestamp, // Use timestamp as ID for search results
			Channel:   match.Channel.ID,
			User:      match.Username,
			Text:      match.Text,
			Timestamp: match.Timestamp,
			CreatedAt: createdAt,
		}

		messages = append(messages, message)
	}

	return messages, nil
}

// GetThread retrieves all messages in a thread
func (sc *Connector) GetThread(ctx context.Context, channelID, threadTS string) (*Thread, error) {
	params := &slack.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: threadTS,
	}

	messages, _, _, err := sc.client.GetConversationReplies(params)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread replies: %w", err)
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages found in thread")
	}

	// First message is the parent
	parentMsg := messages[0]
	createdAt, _ := parseSlackTimestamp(parentMsg.Timestamp)

	thread := &Thread{
		ParentMessage: Message{
			ID:         parentMsg.ClientMsgID,
			Channel:    channelID,
			User:       parentMsg.User,
			Text:       parentMsg.Text,
			Timestamp:  parentMsg.Timestamp,
			ThreadTS:   parentMsg.ThreadTimestamp,
			ReplyCount: parentMsg.ReplyCount,
			CreatedAt:  createdAt,
		},
		Replies: []Message{},
	}

	// Remaining messages are replies
	for i := 1; i < len(messages); i++ {
		msg := messages[i]
		replyCreatedAt, _ := parseSlackTimestamp(msg.Timestamp)

		reply := Message{
			ID:        msg.ClientMsgID,
			Channel:   channelID,
			User:      msg.User,
			Text:      msg.Text,
			Timestamp: msg.Timestamp,
			ThreadTS:  msg.ThreadTimestamp,
			CreatedAt: replyCreatedAt,
		}

		thread.Replies = append(thread.Replies, reply)
	}

	return thread, nil
}

// ListChannels lists all accessible channels
func (sc *Connector) ListChannels(ctx context.Context) ([]Channel, error) {
	params := &slack.GetConversationsParameters{
		ExcludeArchived: true,
		Limit:           200,
		Types:           []string{"public_channel", "private_channel"},
	}

	var allChannels []Channel

	for {
		channels, nextCursor, err := sc.client.GetConversations(params)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversations: %w", err)
		}

		for _, ch := range channels {
			channel := Channel{
				ID:        ch.ID,
				Name:      ch.Name,
				IsPrivate: ch.IsPrivate,
				Members:   ch.NumMembers,
			}

			allChannels = append(allChannels, channel)
		}

		if nextCursor == "" {
			break
		}

		params.Cursor = nextCursor
	}

	return allChannels, nil
}

// GetType returns the connector type
func (sc *Connector) GetType() string {
	return "slack"
}

// GetRateLimit returns current rate limit information
func (sc *Connector) GetRateLimit() *RateLimitInfo {
	sc.rateLimitMu.RLock()
	defer sc.rateLimitMu.RUnlock()

	if sc.rateLimit != nil {
		return &RateLimitInfo{
			Remaining: sc.rateLimit.Remaining,
			Reset:     sc.rateLimit.Reset,
		}
	}
	return &RateLimitInfo{}
}

// GetSyncStatus returns current sync status
func (sc *Connector) GetSyncStatus() *SyncStatus {
	sc.statusMu.RLock()
	defer sc.statusMu.RUnlock()

	return &SyncStatus{
		LastSync:       sc.status.LastSync,
		TotalMessages:  sc.status.TotalMessages,
		TotalChannels:  sc.status.TotalChannels,
		SyncInProgress: sc.status.SyncInProgress,
		Error:          sc.status.Error,
		RateLimit:      sc.GetRateLimit(),
	}
}

// updateSyncStatus updates the sync status
func (sc *Connector) updateSyncStatus(totalMessages, totalChannels int, err error) {
	sc.statusMu.Lock()
	defer sc.statusMu.Unlock()

	sc.status.LastSync = time.Now()
	sc.status.TotalMessages = totalMessages
	sc.status.TotalChannels = totalChannels
	sc.status.SyncInProgress = false

	if err != nil {
		sc.status.Error = err.Error()
	} else {
		sc.status.Error = ""
	}
}

// parseSlackTimestamp converts Slack timestamp to time.Time
func parseSlackTimestamp(ts string) (time.Time, error) {
	// Slack timestamps are Unix timestamps with microsecond precision (e.g., "1234567890.123456")
	parts := strings.Split(ts, ".")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid timestamp format: %s", ts)
	}

	var unixSeconds int64
	var unixMicros int64

	fmt.Sscanf(parts[0], "%d", &unixSeconds)
	fmt.Sscanf(parts[1], "%d", &unixMicros)

	return time.Unix(unixSeconds, unixMicros*1000), nil
}

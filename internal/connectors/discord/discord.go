package discord

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Connector struct {
	client      DiscordClientInterface
	config      *Config
	rateLimit   *RateLimitInfo
	rateLimitMu sync.RWMutex
	status      *SyncStatus
	statusMu    sync.RWMutex
}

type Config struct {
	Token        string        `json:"token"`         // Bot token
	GuildID      string        `json:"guild_id"`      // Server/Guild ID
	Channels     []string      `json:"channels"`      // Channel IDs to index
	SyncInterval time.Duration `json:"sync_interval"` // How often to sync
	MaxMessages  int           `json:"max_messages"`  // Max messages per channel
}

type Message struct {
	ID        string     `json:"id"`
	ChannelID string     `json:"channel_id"`
	GuildID   string     `json:"guild_id"`
	Author    string     `json:"author"`
	Content   string     `json:"content"`
	Timestamp time.Time  `json:"timestamp"`
	EditedAt  *time.Time `json:"edited_at,omitempty"`
	Embeds    int        `json:"embeds"`
	Mentions  []string   `json:"mentions"`
}

type Channel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Topic    string `json:"topic"`
	Position int    `json:"position"`
	GuildID  string `json:"guild_id"`
}

type Guild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MemberCount int    `json:"member_count"`
	OwnerID     string `json:"owner_id"`
}

type Thread struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
	GuildID  string `json:"guild_id"`
	Archived bool   `json:"archived"`
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
		return nil, fmt.Errorf("Discord token is required")
	}

	if config.GuildID == "" {
		return nil, fmt.Errorf("Discord guild ID is required")
	}

	if len(config.Channels) == 0 {
		return nil, fmt.Errorf("at least one Discord channel is required")
	}

	if config.SyncInterval == 0 {
		config.SyncInterval = 5 * time.Minute // Default sync interval
	}

	if config.MaxMessages == 0 {
		config.MaxMessages = 1000 // Default max messages per channel
	}

	session, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	client := NewRealDiscordClient(session)

	connector := &Connector{
		client:    client,
		config:    config,
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	return connector, nil
}

// SyncMessages syncs messages from configured channels
func (dc *Connector) SyncMessages(ctx context.Context) ([]Message, error) {
	// Update sync status
	dc.statusMu.Lock()
	dc.status.SyncInProgress = true
	dc.statusMu.Unlock()

	var allMessages []Message
	totalChannels := len(dc.config.Channels)

	for _, channelID := range dc.config.Channels {
		messages, err := dc.getChannelHistory(ctx, channelID)
		if err != nil {
			log.Printf("Warning: Failed to sync channel %s: %v", channelID, err)
			dc.updateSyncStatus(0, 0, err)
			continue
		}

		allMessages = append(allMessages, messages...)
	}

	dc.updateSyncStatus(len(allMessages), totalChannels, nil)
	return allMessages, nil
}

// getChannelHistory retrieves message history for a specific channel
func (dc *Connector) getChannelHistory(ctx context.Context, channelID string) ([]Message, error) {
	var allMessages []Message
	beforeID := ""
	limit := 100 // Discord API max limit

	for {
		if len(allMessages) >= dc.config.MaxMessages {
			break
		}

		messages, err := dc.client.ChannelMessages(channelID, limit, beforeID, "", "")
		if err != nil {
			return nil, fmt.Errorf("failed to get channel messages: %w", err)
		}

		if len(messages) == 0 {
			break
		}

		for _, msg := range messages {
			// Skip bot messages if needed
			if msg.Author.Bot {
				continue
			}

			timestamp := msg.Timestamp

			var editedAt *time.Time
			if msg.EditedTimestamp != nil {
				editedAt = msg.EditedTimestamp
			}

			var mentions []string
			for _, mention := range msg.Mentions {
				mentions = append(mentions, mention.Username)
			}

			message := Message{
				ID:        msg.ID,
				ChannelID: msg.ChannelID,
				GuildID:   msg.GuildID,
				Author:    msg.Author.Username,
				Content:   msg.Content,
				Timestamp: timestamp,
				EditedAt:  editedAt,
				Embeds:    len(msg.Embeds),
				Mentions:  mentions,
			}

			allMessages = append(allMessages, message)
			beforeID = msg.ID

			if len(allMessages) >= dc.config.MaxMessages {
				break
			}
		}

		if len(messages) < limit {
			break
		}
	}

	return allMessages, nil
}

// SearchMessages searches for messages in a channel
func (dc *Connector) SearchMessages(ctx context.Context, channelID, query string) ([]Message, error) {
	// Note: Discord API doesn't have a native search endpoint for bots
	// We need to fetch messages and filter client-side
	messages, err := dc.getChannelHistory(ctx, channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	var results []Message
	for _, msg := range messages {
		if contains(msg.Content, query) {
			results = append(results, msg)
		}
	}

	return results, nil
}

// GetThreadMessages retrieves all messages from a thread
func (dc *Connector) GetThreadMessages(ctx context.Context, threadID string) ([]Message, error) {
	messages, err := dc.client.ChannelMessages(threadID, 100, "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get thread messages: %w", err)
	}

	var threadMessages []Message
	for _, msg := range messages {
		timestamp := msg.Timestamp

		var editedAt *time.Time
		if msg.EditedTimestamp != nil {
			editedAt = msg.EditedTimestamp
		}

		var mentions []string
		for _, mention := range msg.Mentions {
			mentions = append(mentions, mention.Username)
		}

		message := Message{
			ID:        msg.ID,
			ChannelID: msg.ChannelID,
			GuildID:   msg.GuildID,
			Author:    msg.Author.Username,
			Content:   msg.Content,
			Timestamp: timestamp,
			EditedAt:  editedAt,
			Embeds:    len(msg.Embeds),
			Mentions:  mentions,
		}

		threadMessages = append(threadMessages, message)
	}

	return threadMessages, nil
}

// ListChannels lists all channels in the guild
func (dc *Connector) ListChannels(ctx context.Context) ([]Channel, error) {
	discordChannels, err := dc.client.GuildChannels(dc.config.GuildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild channels: %w", err)
	}

	var channels []Channel
	for _, ch := range discordChannels {
		channelType := fmt.Sprintf("%d", ch.Type)

		channel := Channel{
			ID:       ch.ID,
			Name:     ch.Name,
			Type:     channelType,
			Topic:    ch.Topic,
			Position: ch.Position,
			GuildID:  ch.GuildID,
		}

		channels = append(channels, channel)
	}

	return channels, nil
}

// GetGuild retrieves guild information
func (dc *Connector) GetGuild(ctx context.Context) (*Guild, error) {
	guild, err := dc.client.Guild(dc.config.GuildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild: %w", err)
	}

	g := &Guild{
		ID:          guild.ID,
		Name:        guild.Name,
		Description: guild.Description,
		MemberCount: guild.MemberCount,
		OwnerID:     guild.OwnerID,
	}

	return g, nil
}

// ListThreads lists active threads in a channel
func (dc *Connector) ListThreads(ctx context.Context, channelID string) ([]Thread, error) {
	activeThreads, err := dc.client.ThreadsActive(channelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active threads: %w", err)
	}

	var threads []Thread
	for _, t := range activeThreads.Threads {
		thread := Thread{
			ID:       t.ID,
			Name:     t.Name,
			ParentID: t.ParentID,
			GuildID:  t.GuildID,
			Archived: false,
		}

		threads = append(threads, thread)
	}

	return threads, nil
}

// GetType returns the connector type
func (dc *Connector) GetType() string {
	return "discord"
}

// GetRateLimit returns current rate limit information
func (dc *Connector) GetRateLimit() *RateLimitInfo {
	dc.rateLimitMu.RLock()
	defer dc.rateLimitMu.RUnlock()

	if dc.rateLimit != nil {
		return &RateLimitInfo{
			Remaining: dc.rateLimit.Remaining,
			Reset:     dc.rateLimit.Reset,
		}
	}
	return &RateLimitInfo{}
}

// GetSyncStatus returns current sync status
func (dc *Connector) GetSyncStatus() *SyncStatus {
	dc.statusMu.RLock()
	defer dc.statusMu.RUnlock()

	return &SyncStatus{
		LastSync:       dc.status.LastSync,
		TotalMessages:  dc.status.TotalMessages,
		TotalChannels:  dc.status.TotalChannels,
		SyncInProgress: dc.status.SyncInProgress,
		Error:          dc.status.Error,
		RateLimit:      dc.GetRateLimit(),
	}
}

// updateSyncStatus updates the sync status
func (dc *Connector) updateSyncStatus(totalMessages, totalChannels int, err error) {
	dc.statusMu.Lock()
	defer dc.statusMu.Unlock()

	dc.status.LastSync = time.Now()
	dc.status.TotalMessages = totalMessages
	dc.status.TotalChannels = totalChannels
	dc.status.SyncInProgress = false

	if err != nil {
		dc.status.Error = err.Error()
	} else {
		dc.status.Error = ""
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr))))
}

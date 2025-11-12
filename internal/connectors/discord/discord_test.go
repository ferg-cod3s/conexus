package discord

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

// MockDiscordClient implements DiscordClientInterface for testing
type MockDiscordClient struct {
	ChannelMessagesFunc func(channelID string, limit int, beforeID, afterID, aroundID string) ([]*discordgo.Message, error)
	GuildChannelsFunc   func(guildID string) ([]*discordgo.Channel, error)
	GuildFunc           func(guildID string) (*discordgo.Guild, error)
	ThreadsActiveFunc   func(channelID string) (*discordgo.ThreadsList, error)
}

func (m *MockDiscordClient) ChannelMessages(channelID string, limit int, beforeID, afterID, aroundID string) ([]*discordgo.Message, error) {
	if m.ChannelMessagesFunc != nil {
		return m.ChannelMessagesFunc(channelID, limit, beforeID, afterID, aroundID)
	}
	return []*discordgo.Message{}, nil
}

func (m *MockDiscordClient) GuildChannels(guildID string) ([]*discordgo.Channel, error) {
	if m.GuildChannelsFunc != nil {
		return m.GuildChannelsFunc(guildID)
	}
	return []*discordgo.Channel{}, nil
}

func (m *MockDiscordClient) Guild(guildID string) (*discordgo.Guild, error) {
	if m.GuildFunc != nil {
		return m.GuildFunc(guildID)
	}
	return &discordgo.Guild{}, nil
}

func (m *MockDiscordClient) ThreadsActive(channelID string) (*discordgo.ThreadsList, error) {
	if m.ThreadsActiveFunc != nil {
		return m.ThreadsActiveFunc(channelID)
	}
	return &discordgo.ThreadsList{}, nil
}

func TestNewConnector(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name: "Valid config",
			config: &Config{
				Token:    "test-token",
				GuildID:  "123456789",
				Channels: []string{"987654321"},
			},
			expectErr: false,
		},
		{
			name: "Missing token",
			config: &Config{
				GuildID:  "123456789",
				Channels: []string{"987654321"},
			},
			expectErr: true,
		},
		{
			name: "Missing guild ID",
			config: &Config{
				Token:    "test-token",
				Channels: []string{"987654321"},
			},
			expectErr: true,
		},
		{
			name: "Missing channels",
			config: &Config{
				Token:   "test-token",
				GuildID: "123456789",
			},
			expectErr: true,
		},
		{
			name: "With custom settings",
			config: &Config{
				Token:        "test-token",
				GuildID:      "123456789",
				Channels:     []string{"987654321", "111222333"},
				SyncInterval: 10 * time.Minute,
				MaxMessages:  500,
			},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			connector, err := NewConnector(tc.config)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, connector)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, connector)
				assert.NotNil(t, connector.client)
				assert.NotNil(t, connector.config)
			}
		})
	}
}

func TestSyncMessages(t *testing.T) {
	mockClient := &MockDiscordClient{
		ChannelMessagesFunc: func(channelID string, limit int, beforeID, afterID, aroundID string) ([]*discordgo.Message, error) {
			return []*discordgo.Message{
				{
					ID:        "1",
					ChannelID: channelID,
					GuildID:   "123456789",
					Content:   "Test message 1",
					Timestamp: time.Now(),
					Author: &discordgo.User{
						ID:       "user1",
						Username: "testuser1",
						Bot:      false,
					},
				},
				{
					ID:        "2",
					ChannelID: channelID,
					GuildID:   "123456789",
					Content:   "Test message 2",
					Timestamp: time.Now(),
					Author: &discordgo.User{
						ID:       "user2",
						Username: "testuser2",
						Bot:      false,
					},
				},
			}, nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			Token:       "test-token",
			GuildID:     "123456789",
			Channels:    []string{"987654321"},
			MaxMessages: 1000,
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	messages, err := connector.SyncMessages(ctx)

	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "1", messages[0].ID)
	assert.Equal(t, "Test message 1", messages[0].Content)
	assert.Equal(t, "testuser1", messages[0].Author)
}

func TestSearchMessages(t *testing.T) {
	mockClient := &MockDiscordClient{
		ChannelMessagesFunc: func(channelID string, limit int, beforeID, afterID, aroundID string) ([]*discordgo.Message, error) {
			return []*discordgo.Message{
				{
					ID:        "1",
					ChannelID: channelID,
					GuildID:   "123456789",
					Content:   "This message contains the query word",
					Timestamp: time.Now(),
					Author: &discordgo.User{
						ID:       "user1",
						Username: "testuser",
						Bot:      false,
					},
				},
				{
					ID:        "2",
					ChannelID: channelID,
					GuildID:   "123456789",
					Content:   "This message does not contain it",
					Timestamp: time.Now(),
					Author: &discordgo.User{
						ID:       "user2",
						Username: "testuser2",
						Bot:      false,
					},
				},
			}, nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			Token:       "test-token",
			GuildID:     "123456789",
			Channels:    []string{"987654321"},
			MaxMessages: 1000,
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	messages, err := connector.SearchMessages(ctx, "987654321", "query")

	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "This message contains the query word", messages[0].Content)
}

func TestGetThreadMessages(t *testing.T) {
	mockClient := &MockDiscordClient{
		ChannelMessagesFunc: func(channelID string, limit int, beforeID, afterID, aroundID string) ([]*discordgo.Message, error) {
			return []*discordgo.Message{
				{
					ID:        "thread1",
					ChannelID: channelID,
					GuildID:   "123456789",
					Content:   "Thread message 1",
					Timestamp: time.Now(),
					Author: &discordgo.User{
						ID:       "user1",
						Username: "threaduser1",
						Bot:      false,
					},
				},
				{
					ID:        "thread2",
					ChannelID: channelID,
					GuildID:   "123456789",
					Content:   "Thread message 2",
					Timestamp: time.Now(),
					Author: &discordgo.User{
						ID:       "user2",
						Username: "threaduser2",
						Bot:      false,
					},
				},
			}, nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			Token:    "test-token",
			GuildID:  "123456789",
			Channels: []string{"987654321"},
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	messages, err := connector.GetThreadMessages(ctx, "thread123")

	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "Thread message 1", messages[0].Content)
	assert.Equal(t, "threaduser1", messages[0].Author)
}

func TestListChannels(t *testing.T) {
	mockClient := &MockDiscordClient{
		GuildChannelsFunc: func(guildID string) ([]*discordgo.Channel, error) {
			return []*discordgo.Channel{
				{
					ID:       "channel1",
					GuildID:  guildID,
					Name:     "general",
					Type:     discordgo.ChannelTypeGuildText,
					Topic:    "General discussion",
					Position: 0,
				},
				{
					ID:       "channel2",
					GuildID:  guildID,
					Name:     "dev-team",
					Type:     discordgo.ChannelTypeGuildText,
					Topic:    "Development team channel",
					Position: 1,
				},
			}, nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			Token:    "test-token",
			GuildID:  "123456789",
			Channels: []string{"channel1"},
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	channels, err := connector.ListChannels(ctx)

	assert.NoError(t, err)
	assert.Len(t, channels, 2)
	assert.Equal(t, "channel1", channels[0].ID)
	assert.Equal(t, "general", channels[0].Name)
	assert.Equal(t, "General discussion", channels[0].Topic)
	assert.Equal(t, "channel2", channels[1].ID)
	assert.Equal(t, "dev-team", channels[1].Name)
}

func TestGetGuild(t *testing.T) {
	mockClient := &MockDiscordClient{
		GuildFunc: func(guildID string) (*discordgo.Guild, error) {
			return &discordgo.Guild{
				ID:          guildID,
				Name:        "Test Server",
				Description: "A test Discord server",
				MemberCount: 150,
				OwnerID:     "owner123",
			}, nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			Token:    "test-token",
			GuildID:  "123456789",
			Channels: []string{"channel1"},
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	guild, err := connector.GetGuild(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, guild)
	assert.Equal(t, "123456789", guild.ID)
	assert.Equal(t, "Test Server", guild.Name)
	assert.Equal(t, "A test Discord server", guild.Description)
	assert.Equal(t, 150, guild.MemberCount)
	assert.Equal(t, "owner123", guild.OwnerID)
}

func TestListThreads(t *testing.T) {
	mockClient := &MockDiscordClient{
		ThreadsActiveFunc: func(channelID string) (*discordgo.ThreadsList, error) {
			return &discordgo.ThreadsList{
				Threads: []*discordgo.Channel{
					{
						ID:       "thread1",
						Name:     "Bug Discussion",
						ParentID: channelID,
						GuildID:  "123456789",
					},
					{
						ID:       "thread2",
						Name:     "Feature Request",
						ParentID: channelID,
						GuildID:  "123456789",
					},
				},
			}, nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			Token:    "test-token",
			GuildID:  "123456789",
			Channels: []string{"channel1"},
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	threads, err := connector.ListThreads(ctx, "channel1")

	assert.NoError(t, err)
	assert.Len(t, threads, 2)
	assert.Equal(t, "thread1", threads[0].ID)
	assert.Equal(t, "Bug Discussion", threads[0].Name)
	assert.Equal(t, "thread2", threads[1].ID)
	assert.Equal(t, "Feature Request", threads[1].Name)
}

func TestGetSyncStatus(t *testing.T) {
	connector := &Connector{
		config: &Config{
			Token:    "test-token",
			GuildID:  "123456789",
			Channels: []string{"channel1"},
		},
		rateLimit: &RateLimitInfo{
			Remaining: 100,
			Reset:     time.Now().Add(1 * time.Hour),
		},
		status: &SyncStatus{
			LastSync:       time.Now(),
			TotalMessages:  500,
			TotalChannels:  2,
			SyncInProgress: false,
		},
	}

	status := connector.GetSyncStatus()

	assert.NotNil(t, status)
	assert.Equal(t, 500, status.TotalMessages)
	assert.Equal(t, 2, status.TotalChannels)
	assert.False(t, status.SyncInProgress)
	assert.NotNil(t, status.RateLimit)
	assert.Equal(t, 100, status.RateLimit.Remaining)
}

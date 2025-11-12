package slack

import (
	"context"
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

// MockSlackClient implements SlackClientInterface for testing
type MockSlackClient struct {
	GetConversationHistoryFunc func(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error)
	SearchMessagesFunc         func(query string, params slack.SearchParameters) (*slack.SearchMessages, error)
	GetConversationRepliesFunc func(params *slack.GetConversationRepliesParameters) (msgs []slack.Message, hasMore bool, nextCursor string, err error)
	GetConversationsFunc       func(params *slack.GetConversationsParameters) (channels []slack.Channel, nextCursor string, err error)
}

func (m *MockSlackClient) GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
	if m.GetConversationHistoryFunc != nil {
		return m.GetConversationHistoryFunc(params)
	}
	return &slack.GetConversationHistoryResponse{}, nil
}

func (m *MockSlackClient) SearchMessages(query string, params slack.SearchParameters) (*slack.SearchMessages, error) {
	if m.SearchMessagesFunc != nil {
		return m.SearchMessagesFunc(query, params)
	}
	return &slack.SearchMessages{}, nil
}

func (m *MockSlackClient) GetConversationReplies(params *slack.GetConversationRepliesParameters) (msgs []slack.Message, hasMore bool, nextCursor string, err error) {
	if m.GetConversationRepliesFunc != nil {
		return m.GetConversationRepliesFunc(params)
	}
	return []slack.Message{}, false, "", nil
}

func (m *MockSlackClient) GetConversations(params *slack.GetConversationsParameters) (channels []slack.Channel, nextCursor string, err error) {
	if m.GetConversationsFunc != nil {
		return m.GetConversationsFunc(params)
	}
	return []slack.Channel{}, "", nil
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
				Token:    "xoxb-test-token",
				Channels: []string{"C1234567890"},
			},
			expectErr: false,
		},
		{
			name: "Missing token",
			config: &Config{
				Channels: []string{"C1234567890"},
			},
			expectErr: true,
		},
		{
			name: "Missing channels",
			config: &Config{
				Token: "xoxb-test-token",
			},
			expectErr: true,
		},
		{
			name: "With custom settings",
			config: &Config{
				Token:        "xoxb-test-token",
				Channels:     []string{"C1234567890", "C0987654321"},
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
	mockClient := &MockSlackClient{
		GetConversationHistoryFunc: func(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
			return &slack.GetConversationHistoryResponse{
				Messages: []slack.Message{
					{
						Msg: slack.Msg{
							ClientMsgID: "msg1",
							Timestamp:   "1234567890.123456",
							Text:        "Test message 1",
							User:        "U1234567890",
						},
					},
					{
						Msg: slack.Msg{
							ClientMsgID: "msg2",
							Timestamp:   "1234567891.123456",
							Text:        "Test message 2",
							User:        "U0987654321",
						},
					},
				},
				HasMore: false,
			}, nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			Token:       "xoxb-test-token",
			Channels:    []string{"C1234567890"},
			MaxMessages: 1000,
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	messages, err := connector.SyncMessages(ctx)

	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "msg1", messages[0].ID)
	assert.Equal(t, "Test message 1", messages[0].Text)
	assert.Equal(t, "U1234567890", messages[0].User)
}

func TestSearchMessages(t *testing.T) {
	mockClient := &MockSlackClient{
		SearchMessagesFunc: func(query string, params slack.SearchParameters) (*slack.SearchMessages, error) {
			return &slack.SearchMessages{
				Matches: []slack.SearchMessage{
					{
						Timestamp: "1234567890.123456",
						Text:      "Found message matching query",
						Username:  "testuser",
						Channel: slack.CtxChannel{
							ID: "C1234567890",
						},
					},
				},
			}, nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			Token:    "xoxb-test-token",
			Channels: []string{"C1234567890"},
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	messages, err := connector.SearchMessages(ctx, "test query")

	assert.NoError(t, err)
	assert.Len(t, messages, 1)
	assert.Equal(t, "Found message matching query", messages[0].Text)
	assert.Equal(t, "testuser", messages[0].User)
	assert.Equal(t, "C1234567890", messages[0].Channel)
}

func TestGetThread(t *testing.T) {
	mockClient := &MockSlackClient{
		GetConversationRepliesFunc: func(params *slack.GetConversationRepliesParameters) (msgs []slack.Message, hasMore bool, nextCursor string, err error) {
			return []slack.Message{
				{
					Msg: slack.Msg{
						ClientMsgID:     "parent",
						Timestamp:       "1234567890.123456",
						ThreadTimestamp: "1234567890.123456",
						Text:            "Parent message",
						User:            "U1234567890",
						ReplyCount:      2,
					},
				},
				{
					Msg: slack.Msg{
						ClientMsgID:     "reply1",
						Timestamp:       "1234567891.123456",
						ThreadTimestamp: "1234567890.123456",
						Text:            "Reply 1",
						User:            "U0987654321",
					},
				},
				{
					Msg: slack.Msg{
						ClientMsgID:     "reply2",
						Timestamp:       "1234567892.123456",
						ThreadTimestamp: "1234567890.123456",
						Text:            "Reply 2",
						User:            "U1111111111",
					},
				},
			}, false, "", nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			Token:    "xoxb-test-token",
			Channels: []string{"C1234567890"},
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	thread, err := connector.GetThread(ctx, "C1234567890", "1234567890.123456")

	assert.NoError(t, err)
	assert.NotNil(t, thread)
	assert.Equal(t, "Parent message", thread.ParentMessage.Text)
	assert.Len(t, thread.Replies, 2)
	assert.Equal(t, "Reply 1", thread.Replies[0].Text)
	assert.Equal(t, "Reply 2", thread.Replies[1].Text)
}

func TestListChannels(t *testing.T) {
	mockClient := &MockSlackClient{
		GetConversationsFunc: func(params *slack.GetConversationsParameters) (channels []slack.Channel, nextCursor string, err error) {
			return []slack.Channel{
				{
					GroupConversation: slack.GroupConversation{
						Conversation: slack.Conversation{
							ID: "C1234567890",
						},
						Name: "general",
					},
				},
				{
					GroupConversation: slack.GroupConversation{
						Conversation: slack.Conversation{
							ID: "C0987654321",
						},
						Name: "dev-team",
					},
				},
			}, "", nil
		},
	}

	connector := &Connector{
		client: mockClient,
		config: &Config{
			Token:    "xoxb-test-token",
			Channels: []string{"C1234567890"},
		},
		rateLimit: &RateLimitInfo{},
		status:    &SyncStatus{},
	}

	ctx := context.Background()
	channels, err := connector.ListChannels(ctx)

	assert.NoError(t, err)
	assert.Len(t, channels, 2)
	assert.Equal(t, "C1234567890", channels[0].ID)
	assert.Equal(t, "general", channels[0].Name)
}

func TestParseSlackTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp string
		expectErr bool
	}{
		{
			name:      "Valid timestamp",
			timestamp: "1234567890.123456",
			expectErr: false,
		},
		{
			name:      "Invalid format - no dot",
			timestamp: "1234567890",
			expectErr: true,
		},
		{
			name:      "Invalid format - multiple dots",
			timestamp: "1234567890.123.456",
			expectErr: true,
		},
		{
			name:      "Empty timestamp",
			timestamp: "",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseSlackTimestamp(tc.timestamp)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.False(t, result.IsZero())
			}
		})
	}
}

func TestGetSyncStatus(t *testing.T) {
	connector := &Connector{
		config: &Config{
			Token:    "xoxb-test-token",
			Channels: []string{"C1234567890"},
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

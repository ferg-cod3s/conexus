package slack

import (
	"github.com/slack-go/slack"
)

// SlackClientInterface defines the interface for Slack API operations
// This allows for easier testing with mock clients
type SlackClientInterface interface {
	GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error)
	SearchMessages(query string, params slack.SearchParameters) (*slack.SearchMessages, error)
	GetConversationReplies(params *slack.GetConversationRepliesParameters) (msgs []slack.Message, hasMore bool, nextCursor string, err error)
	GetConversations(params *slack.GetConversationsParameters) (channels []slack.Channel, nextCursor string, err error)
}

// RealSlackClient wraps the real Slack client to implement SlackClientInterface
type RealSlackClient struct {
	client *slack.Client
}

// NewRealSlackClient creates a new real Slack client wrapper
func NewRealSlackClient(client *slack.Client) *RealSlackClient {
	return &RealSlackClient{client: client}
}

func (r *RealSlackClient) GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error) {
	return r.client.GetConversationHistory(params)
}

func (r *RealSlackClient) SearchMessages(query string, params slack.SearchParameters) (*slack.SearchMessages, error) {
	return r.client.SearchMessages(query, params)
}

func (r *RealSlackClient) GetConversationReplies(params *slack.GetConversationRepliesParameters) (msgs []slack.Message, hasMore bool, nextCursor string, err error) {
	return r.client.GetConversationReplies(params)
}

func (r *RealSlackClient) GetConversations(params *slack.GetConversationsParameters) (channels []slack.Channel, nextCursor string, err error) {
	return r.client.GetConversations(params)
}

package discord

import (
	"github.com/bwmarrin/discordgo"
)

// DiscordClientInterface defines the interface for Discord API operations
// This allows for easier testing with mock clients
type DiscordClientInterface interface {
	ChannelMessages(channelID string, limit int, beforeID, afterID, aroundID string) ([]*discordgo.Message, error)
	GuildChannels(guildID string) ([]*discordgo.Channel, error)
	Guild(guildID string) (*discordgo.Guild, error)
	ThreadsActive(channelID string) (*discordgo.ThreadsList, error)
}

// RealDiscordClient wraps the real Discord session to implement DiscordClientInterface
type RealDiscordClient struct {
	session *discordgo.Session
}

// NewRealDiscordClient creates a new real Discord client wrapper
func NewRealDiscordClient(session *discordgo.Session) *RealDiscordClient {
	return &RealDiscordClient{session: session}
}

func (r *RealDiscordClient) ChannelMessages(channelID string, limit int, beforeID, afterID, aroundID string) ([]*discordgo.Message, error) {
	return r.session.ChannelMessages(channelID, limit, beforeID, afterID, aroundID)
}

func (r *RealDiscordClient) GuildChannels(guildID string) ([]*discordgo.Channel, error) {
	return r.session.GuildChannels(guildID)
}

func (r *RealDiscordClient) Guild(guildID string) (*discordgo.Guild, error) {
	return r.session.Guild(guildID)
}

func (r *RealDiscordClient) ThreadsActive(channelID string) (*discordgo.ThreadsList, error) {
	return r.session.ThreadsActive(channelID)
}

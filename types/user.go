package types

// User represents a Discord and Forum user
type User struct {
	DiscordID string `json:"discord_id" bson:"discord_id"` // discord user ID
	ForumID   string `json:"forum_id" bson:"forum_id"`     // IPB forum user ID
}

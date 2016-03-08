package bot

import "github.com/wcharczuk/go-slack"

// Giffy is the bot that talks to slack realtime to post gifs.
type Giffy struct {
	id    string
	token string

	organizationName string
	configuration    map[string]string
	state            map[string]interface{}

	client *slack.Client
}

// ID returns the identifier for the bot.
func (g Giffy) ID() string {
	return g.id
}

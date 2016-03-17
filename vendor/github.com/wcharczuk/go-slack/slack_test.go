package slack

import (
	"os"
	"testing"

	"github.com/blendlabs/go-assert"
)

func getSlackToken(a *assert.Assertions) string {
	token := os.Getenv("SLACK_TOKEN")
	//a.NotEmpty(token, "`SLACK_TOKEN` environment variable must be set.")
	if len(token) == 0 {
		return UUIDv4().ToShortString()
	}
	return token
}

func TestClientAuthTest(t *testing.T) {
	a := assert.New(t)
	defer ClearMockedResponses()

	MockResponseFromFile("POST", "https://slack.com/api/auth.test", 200, "testdata/auth.test.json")

	c := NewClient(getSlackToken(a))
	result, resultErr := c.AuthTest()
	a.Nil(resultErr)
	a.True(result.OK)
	a.Empty(result.Error)
}

func TestClientChannelsHistory(t *testing.T) {
	a := assert.New(t)
	defer ClearMockedResponses()
	MockResponseFromFile("POST", "https://slack.com/api/channels.history", 200, "testdata/channels.history.json")

	c := NewClient(getSlackToken(a))
	history, historyErr := c.ChannelsHistory("CTESTCHANEL", nil, nil, -1, true)
	a.Nil(historyErr)
	a.NotEmpty(history.Messages)
}

func TestClientChannelsInfo(t *testing.T) {
	a := assert.New(t)
	defer ClearMockedResponses()
	MockResponseFromFile("POST", "https://slack.com/api/channels.info", 200, "testdata/channels.info.json")

	c := NewClient(getSlackToken(a))
	info, infoErr := c.ChannelsInfo("CTESTCHANEL")
	a.Nil(infoErr)
	a.NotEmpty(info.ID)
	a.NotNil(info.Latest)
	a.NotNil(info.Purpose)
	a.NotNil(info.Topic)
}

go-slack
========

[![Build Status](https://travis-ci.org/wcharczuk/go-slack.svg?branch=master)](https://travis-ci.org/wcharczuk/go-slack) [![GoDoc](https://godoc.org/github.com/wcharczuk/go-slack?status.svg)](http://godoc.org/github.com/wcharczuk/go-slack)

This is a very basic real time api client for slack. It abstracts away the details of the websocket connection and uses goroutines and "listeners" to handle incoming messages. 

##Example

```go
import "github.com/wcharczuk/go-slack"
...

client := slack.Connect(os.Getenv("SLACK_TOKEN"))
client.Listen(slack.EventHello, func(m *slack.Message, c *slack.Client) {
	fmt.Println("connected")
})
client.Listen(slack.EventMessage, func(m *slack.Message, c *slack.Client) {
	fmt.Prinln("message received!")
})
session, err := client.Start() //session has the current users list and channel list
if err != nil {
	fmt.Printf("%v\n", err)
}
```

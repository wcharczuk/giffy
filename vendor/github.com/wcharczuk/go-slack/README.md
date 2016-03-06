go-slack
========

This is a very basic real time api client for slack. It abstracts away the details of the websocket connection and uses goroutines and "listeners" to handle incoming messages. 

##Example

```go
import "github.com/wcharczuk/go-slack"
...

client := slack.Connect(TOKEN())
client.Listen(slack.EVENT_HELLO, func(m *slack.Message, c *slack.Client) {
	fmt.Println("connected")
})
client.Listen(slack.EVENT_MESSAGE, func(m *slack.Message, c *slack.Client) {
	fmt.Prinln("message received!")
})
session, err := client.Start() //session has the current users list and channel list
if err != nil {
	fmt.Printf("%v\n", err)
}
```
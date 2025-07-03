package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	messages := []int{}
	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		message := int(body["message"].(float64))
		messages = append(messages, message)

		responseBody := map[string]any{}
		responseBody["type"] = "broadcast_ok"

		return n.Reply(msg, responseBody)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		responseBody := map[string]any{}
		responseBody["type"] = "read_ok"
		responseBody["messages"] = messages

		return n.Reply(msg, responseBody)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		responseBody := map[string]any{}
		responseBody["type"] = "topology_ok"

		return n.Reply(msg, responseBody)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

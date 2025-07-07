package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type BroadcastBody struct {
	maelstrom.MessageBody
	Message int `json:"message"`
}

type broadcastRequest struct {
	messageID   int
	message     int
	destination string
}

func main() {
	n := maelstrom.NewNode()
	var messages []int
	var neighbors []string
	var retryRequests []broadcastRequest

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body BroadcastBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		if !slices.Contains(messages, body.Message) {
			messages = append(messages, body.Message)

			for _, node := range neighbors {
				// Don't send back broadcast
				if node == msg.Src {
					continue
				}

				request := broadcastRequest{
					destination: node,
					message:     body.Message,
				}
				retryRequests = append(retryRequests, request)

				go func() {
					responseReceived := false
					for {
						err := n.RPC(
							node,
							map[string]any{
								"type":    "broadcast",
								"message": body.Message,
							},
							func(msg maelstrom.Message) error {
								var body BroadcastBody
								if err := json.Unmarshal(msg.Body, &body); err != nil {
									return err
								}

								responseReceived = true
								retryRequests = slices.DeleteFunc(retryRequests, func(req broadcastRequest) bool {
									return req.message == body.Message && req.destination == msg.Src
								})

								return nil
							},
						)
						if err != nil {
							fmt.Println(fmt.Errorf("failed to broadcast message to node %s: %v", node, err))
						}

						if responseReceived {
							return
						}

						time.Sleep(1 * time.Second)
					}
				}()
			}
		}

		if body.MsgID != 0 {
			return n.Reply(msg, map[string]any{
				"type": "broadcast_ok",
			})
		}

		return nil
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
		l := log.New(os.Stderr, n.ID(), 0)
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		topology := body["topology"].(map[string]any)
		topologyNeighbours := topology[n.ID()].([]any)
		for _, neighbor := range topologyNeighbours {
			neighbors = append(neighbors, neighbor.(string))
		}

		l.Printf("Topology received. Neghbors are: %v", neighbors)

		responseBody := map[string]any{}
		responseBody["type"] = "topology_ok"

		return n.Reply(msg, responseBody)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

package main //nolint:cyclop

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type BroadcastBody struct {
	maelstrom.MessageBody
	Message int `json:"message"`
}

type ToplogyBody struct{}

type ToplogyRequest struct {
	Topology map[string][]string `json:"topology"`
	maelstrom.MessageBody
}

// nolint: gocognit,funlen
func main() {
	n := maelstrom.NewNode()
	var messages []int
	var neighbors []string

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body BroadcastBody
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// We already have message. Return
		if slices.Contains(messages, body.Message) {
			if body.MsgID != 0 {
				return n.Reply(msg, map[string]any{
					"type": "broadcast_ok",
				})
			}

			return nil
		}

		messages = append(messages, body.Message)
		for _, node := range neighbors {
			// Don't send back broadcast
			if node == msg.Src {
				continue
			}

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
							responseReceived = true
							return nil
						},
					)
					if err != nil {
						fmt.Println(fmt.Errorf("failed to broadcast message to node %s: %w", node, err))
					}

					if responseReceived {
						return
					}

					time.Sleep(1 * time.Second)
				}
			}()
		}

		if body.MsgID != 0 {
			return n.Reply(msg, map[string]any{
				"type": "broadcast_ok",
			})
		}

		return nil
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": messages,
		})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body ToplogyRequest
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		neighbors = body.Topology[n.ID()]
		return n.Reply(msg, map[string]any{
			"type": "topology_ok",
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

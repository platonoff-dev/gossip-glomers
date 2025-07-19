package main

import (
	"encoding/json"
	"gossip-glomers/internal/worker"
	"log"
	"os"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type ReplicateMessage struct {
	maelstrom.MessageBody
	Value map[string]int `json:"value"`
}

type AddMessage struct {
	maelstrom.MessageBody
	Delta int `json:"delta"`
}

func main() {
	l := log.New(os.Stderr, "TASK: ", log.Default().Flags())
	n := maelstrom.NewNode()

	gcounter := newCounter()

	w := worker.Worker{
		Tasks: []worker.Task{
			{
				Dt: 5 * time.Second,
				F: func() error {
					for _, neighbor := range n.NodeIDs() {
						err := n.Send(
							neighbor,
							map[string]any{
								"type":  "replicate",
								"value": gcounter.serialize(),
							},
						)
						if err != nil {
							l.Println("Send failed: ", err)
						}
					}

					return nil
				},
			},
		},
	}

	n.Handle("add", func(msg maelstrom.Message) error {
		var body AddMessage
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		gcounter.add(n.ID(), body.Delta)

		return n.Reply(msg, map[string]any{
			"type": "add_ok",
		})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type":  "read_ok",
			"value": gcounter.read(),
		})
	})

	n.Handle("replicate", func(msg maelstrom.Message) error {
		var body ReplicateMessage
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		gcounter.merge(deserialize(body.Value))
		return nil
	})

	w.Start()
	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}

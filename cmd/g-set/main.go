package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"gossip-glomers/internal/worker"
)

type ReplicateMessage struct {
	maelstrom.MessageBody
	Value []int `json:"value"`
}

type AddMessage struct {
	maelstrom.MessageBody
	Element int `json:"element"`
}

func main() {
	l := log.New(os.Stderr, "TASK: ", log.Default().Flags())
	n := maelstrom.NewNode()

	gset := newSet()

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
								"value": gset.serialize(),
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

		gset.add(body.Element)

		return n.Reply(msg, map[string]any{
			"type": "add_ok",
		})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{
			"type":  "read_ok",
			"value": gset.serialize(),
		})
	})

	n.Handle("replicate", func(msg maelstrom.Message) error {
		var body ReplicateMessage
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		gset.union(deserialize(body.Value))
		return nil
	})

	w.Start()

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

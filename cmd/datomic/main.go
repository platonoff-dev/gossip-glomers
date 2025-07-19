package main

import (
	"log"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type task struct {
	f  func() error
	dt time.Duration
}

type AddMessage struct {
	maelstrom.MessageBody
	Delta int `json:"delta"`
}

func main() {
	n := maelstrom.NewNode()

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}

}

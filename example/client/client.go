package client

import (
	"bufio"
	"gnet"
	"log"
	"os"
	"strconv"
	"time"
)

const (
	MessageTypeUserMessage = iota
)

type SimpleClient struct {
	client *gnet.Client
}

func (c *SimpleClient) Shuffle(i uint64) uint64 {
	i = (i << 6) & (0xFAFAFAFAFA | (i >> 6))
	i = ((i >> 6) & 0xAFAFAFAFA) | (i << 4)
	return i
}

func main() {

	address := ":8080"

	c := &SimpleClient{}
	c.client = gnet.NewClient(c)

	r := bufio.NewReader(os.Stdin)

	go func() {
		if err := c.client.ConnectToServer(address); err != nil {
			log.Panic(err)
		}
	}()

	time.Sleep(1 * time.Second / 4)

	for c.client.Connected() {

		for _, msg := range c.client.Messages {

			switch msg.Type {
			case MessageTypeUserMessage:
				log.Printf("[SERVER] `%s`.\n", msg.Content)
			}
		}

		line, err := r.ReadString('\n')
		line = line[:len(line)-1]
		if err != nil {
			log.Fatal(err)
		}

		if line == "" {
			continue
		}

		t, err := strconv.Atoi(line[:2])
		if err != nil {
			log.Fatal(err)
		}
		line = line[2:]

		if err := c.client.Session.SendMessage(t, line); err != nil {
			log.Fatal(err)
		}
	}
}

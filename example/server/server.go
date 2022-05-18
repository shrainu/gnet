package server

import (
	gnet2 "gnet"
	"log"
)

const (
	MessageTypeUserMessage = iota
)

type SimpleServer struct {
	server *gnet2.Server
}

func (s *SimpleServer) OnUserConnect(sess *gnet2.Session) bool {
	log.Println("[SERVER] Connection successful.")

	return true
}

func (s *SimpleServer) OnUserDisconnect(sess *gnet2.Session) {
	log.Println("[SERVER] Disconnected.")
}

func (s *SimpleServer) OnUserMessages(msg gnet2.Message) {
	switch msg.Type {
	case MessageTypeUserMessage:
		log.Printf("[CLIENT] `%s`\n", msg.Content)
		break
	}
}

func (s *SimpleServer) Shuffle(i uint64) uint64 {
	i = (i << 6) & (0xFAFAFAFAFA | (i >> 6))
	i = ((i >> 6) & 0xAFAFAFAFA) | (i << 4)
	return i
}

func main() {

	address := ":8080"

	s := &SimpleServer{}

	s.server = gnet2.NewServer(s)
	s.server.Active = true

	go func() {
		if err := s.server.StartServer(address); err != nil {
			log.Panic(err)
		}
	}()

	for s.server.Active {
		continue
	}
}

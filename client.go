package gnet

import (
	"encoding/binary"
	"log"
	"net"
	"strings"
)

type ClientInterface interface {
	Shuffle(uint64) uint64
}

type Client struct {
	Interface ClientInterface
	Session   *Session
	Channel   chan Message
}

func NewClient(ci ClientInterface) *Client {
	return &Client{
		Interface: ci,
		Session:   nil,
		Channel:   nil,
	}
}

func authenticateServer(ci ClientInterface, s *Session, m Message) error {
	i := binary.BigEndian.Uint64([]byte(m.Content))
	key := ci.Shuffle(i)

	buff := make([]byte, 8)
	binary.BigEndian.PutUint64(buff, key)

	err := s.SendMessage(0, string(buff))
	return err
}

func (c *Client) Connected() bool {
	return c.Session.Active
}

func (c *Client) ConnectToServer(address string) error {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	c.Session = &Session{
		Conn:   conn,
		Active: true,
	}

	// Log
	log.Println("[CLIENT] Connected.")

	// Buffer
	buffer := make([]byte, 4096)

	// Message Queue
	var msgQueue []Message = nil

	// Authentication
	auth := false

	for {
		n, err := c.Session.Conn.Read(buffer)
		if err != nil {
			c.Session.Close()
			break
		}

		// Validate message length
		if n <= 2 {
			log.Println("[SERVER] Zero bytes. Closing the connection")
			c.Session.Close()
		}

		// Clear the queues
		msgQueue = nil

		msg := string(buffer[:n])

		for {
			index := strings.Index(msg, "\r\n")
			if index == -1 {
				break
			}

			cpyBuff := make([]byte, len(msg[:index]))
			copy(cpyBuff, msg[:index])

			cpy := string(cpyBuff)

			msgType, err := getMessageType(cpy)
			if err != nil {
				log.Println(err)
			} else {
				cpy = cpy[2:]

				msgQueue = append(msgQueue, Message{
					Sess:    c.Session,
					Type:    msgType,
					Content: cpy,
				})
			}

			if index+2 < len(msg) {
				msg = msg[index+2:]
			} else {
				break
			}
		}

		for _, msg := range msgQueue {
			if !auth {
				err := authenticateServer(c.Interface, c.Session, msg)
				if err != nil {
					log.Fatal(err)
				}
				auth = true
				continue
			}
			c.Channel <- msg
		}
	}

	return nil
}

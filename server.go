package gnet

import (
	"encoding/binary"
	"log"
	"math/rand"
	"net"
	"strings"
)

type Server struct {
	Interface ServerInterface
	Sessions  []*Session
	Listener  net.Listener
	Active    bool
	AuthKey   uint64
}

func NewServer(si ServerInterface) *Server {
	return &Server{
		Active:    false,
		Listener:  nil,
		Sessions:  nil,
		Interface: si,
	}
}

func authenticateUser(si ServerInterface, m Message, key uint64) bool {
	i := binary.BigEndian.Uint64([]byte(m.Content))
	if i == si.Shuffle(key) {
		return true
	}
	return false
}

func (s *Server) serverHandleConnection(si ServerInterface, sess *Session, key uint64) {

	// Send on user connect event, accept connection according to result
	if !si.OnUserConnect(sess) {
		si.OnUserDisconnect(sess)
		s.CloseSession(sess)
		return
	}

	// Send the auth key
	keyBuffer := make([]byte, 8)
	binary.BigEndian.PutUint64(keyBuffer, key)
	if err := sess.SendMessage(0, string(keyBuffer)); err != nil {
		si.OnUserDisconnect(sess)
		s.CloseSession(sess)
		return
	}

	// Message buffer
	buffer := make([]byte, 4096)

	// Message Queue
	var msgQueue []Message = nil

	// Authentication
	auth := false

	for {
		n, err := sess.Conn.Read(buffer)
		if err != nil {
			sess.Active = false
			si.OnUserDisconnect(sess)
			break
		}

		// Validate message length
		if n <= 2 {
			log.Println("[SERVER] Zero bytes. Closing the connection")
			si.OnUserDisconnect(sess)
			s.CloseSession(sess)
		}

		// Clear the queue
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
					Sess:    sess,
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
				auth = authenticateUser(si, msg, key)
				if auth {
					log.Println("[CLIENT] Authenticated.")
					_ = sess.SendMessage(0, "[SERVER] Authenticated.")
				}
				continue
			}

			si.OnUserMessages(msg)
		}

		if !auth {
			log.Println("[CLIENT] Failed to authenticate")
			si.OnUserDisconnect(sess)
			s.CloseSession(sess)
			break
		}
	}
}

func (s *Server) StartServer(address string) error {

	// Start listening to given address
	var err error
	s.Listener, err = net.Listen("tcp", address)

	if err != nil {
		return err
	}

	// Generate key
	s.AuthKey = rand.Uint64()

	// Log
	log.Println("[SERVER] Started.")

	for {
		if !s.Active {
			break
		}

		conn, err := s.Listener.Accept()
		if err != nil {
			log.Panic(err)
		}

		sess := &Session{
			Conn:   conn,
			Active: true,
		}

		s.Sessions = append(s.Sessions, sess)

		go s.serverHandleConnection(s.Interface, sess, s.AuthKey)
	}

	return nil
}

func (s *Server) CloseSession(sess *Session) bool {
	s.Interface.OnUserDisconnect(sess)
	sess.Close()

	index := -1
	for i, v := range s.Sessions {
		if v == sess {
			index = i
		}
	}

	if index == -1 {
		return false
	}

	s.Sessions[index] = s.Sessions[len(s.Sessions)-1]
	s.Sessions = s.Sessions[:len(s.Sessions)-1]
	return true
}

func (s *Server) CloseServer() {
	s.Active = false

	for _, user := range s.Sessions {
		s.Interface.OnUserDisconnect(user)
		user.Close()
	}
	s.Sessions = nil

	err := s.Listener.Close()
	if err != nil {
		log.Println("[SERVER] Error during termination:\n", err)
	}
}

func (s *Server) SendMessage(sess *Session, t int, content string) {
	err := sess.SendMessage(t, content)
	if err != nil {
		s.Interface.OnUserDisconnect(sess)
	}
}

type ServerInterface interface {
	Shuffle(uint64) uint64

	OnUserConnect(sess *Session) bool
	OnUserDisconnect(sess *Session)
	OnUserMessages(msg Message)
}

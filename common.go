package gnet

import (
	"encoding/binary"
	"net"
)

type Session struct {
	Conn   net.Conn
	Active bool
}

func (s *Session) Close() {
	s.Active = false
	_ = s.Conn.Close()
}

func (s *Session) Write(msg string) error {
	buff := []byte(msg + "\r\n")
	_, err := s.Conn.Write(buff)
	return err
}

func (s *Session) SendMessage(t int, content string) error {
	buff := make([]byte, 2)
	binary.BigEndian.PutUint16(buff, uint16(t))
	err := s.Write(string(buff) + content)
	return err
}

package gnet

import (
	"encoding/binary"
	"fmt"
)

type Message struct {
	Sess    *Session
	Type    int
	Content string
}

func getMessageType(msg string) (int, error) {
	if len(msg) < 2 {
		return -1, fmt.Errorf("invalid message length")
	}

	t := int(binary.BigEndian.Uint16([]byte(msg[:2])))

	return t, nil
}

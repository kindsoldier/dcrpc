/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

type Packet struct {
	header     []byte
	rcpPayload []byte
}

func NewEmptyPacket() *Packet {
	packet := &Packet{
		header:     make([]byte, 0),
		rcpPayload: make([]byte, 0),
	}
	return packet
}

/*
 *
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 *
 */

package dsrpc

import (
    "encoding/json"
)

type Packet struct {
    header      []byte
    rcpPayload  []byte
}

func NewPacket() *Packet {
    return &Packet{}
}

func (this *Packet) JSON() []byte {
    jBytes, _ := json.Marshal(this)
    return jBytes
}

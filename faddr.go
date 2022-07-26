/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

type FAddr struct {
    network string
    address string
}

func NewFAddr() *FAddr {
    var addr FAddr
    addr.network  = "tcp"
    addr.address = "127.0.0.1:5000"
    return &addr
}

func (addr *FAddr) Network() string {
    return addr.network
}

func (addr *FAddr) String() string {
    return addr.address
}

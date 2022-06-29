/*
 *
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 *
 */

package dsrpc

import (
    "io"
    "net"
    "time"
)

type Context struct {
    start       time.Time
    remoteHost  string
    sockReader  io.Reader
    sockWriter  io.Writer

    reqHeader   *Header
    reqRPC      *Request
    reqPacket   *Packet

    resPacket   *Packet
    resHeader   *Header
    resRPC      *Response

    binReader   io.Reader
    binWriter   io.Writer
}


func NewContext() *Context {
    context := &Context{}
    context.start = time.Now()
    return context
}

func CreateContext(conn net.Conn) *Context {
    context := &Context{}
    context.start = time.Now()
    context.sockReader = conn
    context.sockWriter = conn

    context.reqPacket = NewPacket()
    context.resPacket = NewPacket()

    context.reqHeader = NewHeader()
    context.reqRPC   = NewRequest()

    context.resHeader = NewHeader()
    context.resRPC = NewResponse()
    context.resRPC = NewResponse()

    return context
}

func (context *Context) Request() *Request  {
    return context.reqRPC
}


func (context *Context) SetAuthIdent(ident []byte)  {
    context.reqRPC.Auth.Ident = ident
}

func (context *Context) SetAuthSalt(salt []byte)  {
    context.reqRPC.Auth.Salt = salt
}

func (context *Context) SetAuthHash(hash []byte)  {
    context.reqRPC.Auth.Hash = hash
}

func (context *Context) AuthIdent() []byte {
    return context.reqRPC.Auth.Ident
}

func (context *Context) AuthSalt() []byte {
    return context.reqRPC.Auth.Salt
}

func (context *Context) AuthHash() []byte {
    return context.reqRPC.Auth.Hash
}

func (context *Context) Auth() *Auth {
    return context.reqRPC.Auth
}

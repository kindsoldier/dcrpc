/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
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

func (context *Context) RemoteHost() string {
    return context.remoteHost
}

func (context *Context) Start() time.Time {
    return context.start
}

func (context *Context) Method() string {
    var method string
    if context.reqRPC != nil {
        method = context.reqRPC.Method
    }
    return method
}

func (context *Context) ReqRpcSize() int64 {
    var size int64
    if context.reqHeader != nil {
        size = context.reqHeader.rpcSize
    }
    return size
}


func (context *Context) ReqBinSize() int64 {
    var size int64
    if context.reqHeader != nil {
        size = context.reqHeader.binSize
    }
    return size
}

func (context *Context) ResBinSize() int64 {
    var size int64
    if context.resHeader != nil {
        size = context.resHeader.binSize
    }
    return size
}

func (context *Context) ResRpcSize() int64 {
    var size int64
    if context.resHeader != nil {
        size = context.resHeader.rpcSize
    }
    return size
}

func (context *Context) ReqSize() int64 {
    var size int64
    if context.reqHeader != nil {
        size += context.reqHeader.binSize
        size += context.reqHeader.rpcSize
    }
    return size
}

func (context *Context) ResSize() int64 {
    var size int64
    if context.resHeader != nil {
        size += context.resHeader.binSize
        size += context.resHeader.rpcSize
    }
    return size
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

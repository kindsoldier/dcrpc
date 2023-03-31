/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
	"io"
	"net"
	"time"
)

type Content struct {
	start      time.Time
	remoteHost string

	sockReader io.Reader
	sockWriter io.Writer

	reqPacket *Packet
	reqHeader *Header
	reqBlock  *Request

	resPacket *Packet
	resHeader *Header
	resBlock  *Response

	binReader io.Reader
	binWriter io.Writer
}

func CreateContent(conn net.Conn) *Content {
	context := &Content{
		start:      time.Now(),
		sockReader: conn,
		sockWriter: conn,

		reqPacket: NewEmptyPacket(),
		reqHeader: NewEmptyHeader(),
		reqBlock:  NewEmptyRequest(),

		resPacket: NewEmptyPacket(),
		resHeader: NewEmptyHeader(),
		resBlock:  NewEmptyResponse(),
	}
	return context
}

func (context *Content) Request() *Request {
	return context.reqBlock
}

func (context *Content) RemoteHost() string {
	return context.remoteHost
}

func (context *Content) Start() time.Time {
	return context.start
}

func (context *Content) Method() string {
	var method string
	if context.reqBlock != nil {
		method = context.reqBlock.Method
	}
	return method
}

func (context *Content) ReqRpcSize() int64 {
	var size int64
	if context.reqHeader != nil {
		size = context.reqHeader.rpcSize
	}
	return size
}

func (context *Content) ReqBinSize() int64 {
	var size int64
	if context.reqHeader != nil {
		size = context.reqHeader.binSize
	}
	return size
}

func (context *Content) ResBinSize() int64 {
	var size int64
	if context.resHeader != nil {
		size = context.resHeader.binSize
	}
	return size
}

func (context *Content) ResRpcSize() int64 {
	var size int64
	if context.resHeader != nil {
		size = context.resHeader.rpcSize
	}
	return size
}

func (context *Content) ReqSize() int64 {
	var size int64
	if context.reqHeader != nil {
		size += context.reqHeader.binSize
		size += context.reqHeader.rpcSize
	}
	return size
}

func (context *Content) ResSize() int64 {
	var size int64
	if context.resHeader != nil {
		size += context.resHeader.binSize
		size += context.resHeader.rpcSize
	}
	return size
}

func (context *Content) SetAuthIdent(ident []byte) {
	context.reqBlock.Auth.Ident = ident
}

func (context *Content) SetAuthSalt(salt []byte) {
	context.reqBlock.Auth.Salt = salt
}

func (context *Content) SetAuthHash(hash []byte) {
	context.reqBlock.Auth.Hash = hash
}

func (context *Content) AuthIdent() []byte {
	return context.reqBlock.Auth.Ident
}

func (context *Content) AuthSalt() []byte {
	return context.reqBlock.Auth.Salt
}

func (context *Content) AuthHash() []byte {
	return context.reqBlock.Auth.Hash
}

func (context *Content) Auth() *Auth {
	return context.reqBlock.Auth
}

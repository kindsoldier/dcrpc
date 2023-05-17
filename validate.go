/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
	"context"
	"io"
	"net"
)

func LocalExec(method string, param, result any, auth *Auth, handler HandlerFunc) error {
	var err error

	cliConn, srvConn := NewFConn()

	content := CreateContent(cliConn)
	content.reqBlock.Method = method

	if param != nil {
		content.reqBlock.Params = param
	}
	if auth != nil {
		content.reqBlock.Auth = auth
	}
	if result != nil {
		content.resBlock.Result = result
	}

	err = content.createRequest()
	if err != nil {
		return err
	}
	err = content.writeRequest()
	if err != nil {
		return err
	}
	err = LocalService(srvConn, handler)
	if err != nil {
		return err
	}
	err = content.readResponse()
	if err != nil {
		return err
	}
	err = content.bindResponse()
	if err != nil {
		return err
	}

	return err
}

func LocalPut(ctx context.Context, method string, reader io.Reader, size int64, param, result any, auth *Auth, handler HandlerFunc) error {

	var err error

	cliConn, srvConn := NewFConn()

	content := CreateContent(cliConn)
	content.reqBlock.Method = method

	if param != nil {
		content.reqBlock.Params = param
	}
	if auth != nil {
		content.reqBlock.Auth = auth
	}
	if result != nil {
		content.resBlock.Result = result
	}

	content.binReader = reader
	content.binWriter = cliConn

	content.reqHeader.binSize = size

	err = content.createRequest()
	if err != nil {
		return err
	}
	err = content.writeRequest()
	if err != nil {
		return err
	}
	err = content.uploadBin(ctx)
	if err != nil {
		return err
	}
	err = LocalService(srvConn, handler)
	if err != nil {
		return err
	}
	err = content.readResponse()
	if err != nil {
		return err
	}
	err = content.bindResponse()
	if err != nil {
		return err
	}
	return err
}

func LocalGet(ctx context.Context, method string, writer io.Writer, param, result any, auth *Auth, handler HandlerFunc) error {
	var err error

	cliConn, srvConn := NewFConn()

	content := CreateContent(cliConn)
	content.reqBlock.Method = method

	if param != nil {
		content.reqBlock.Params = param
	}
	if auth != nil {
		content.reqBlock.Auth = auth
	}
	if result != nil {
		content.resBlock.Result = result
	}

	content.binReader = cliConn
	content.binWriter = writer

	err = content.createRequest()
	if err != nil {
		return err
	}
	err = content.writeRequest()
	if err != nil {
		return err
	}

	err = LocalService(srvConn, handler)
	if err != nil {
		return err
	}
	err = content.readResponse()
	if err != nil {
		return err
	}
	err = content.downloadBin(ctx)
	if err != nil {
		return err
	}
	err = content.bindResponse()
	if err != nil {
		return err
	}
	return err
}

func LocalService(conn net.Conn, handler HandlerFunc) error {
	var err error
	content := CreateContent(conn)

	remoteAddr := conn.RemoteAddr().String()
	remoteHost, _, _ := net.SplitHostPort(remoteAddr)
	content.remoteHost = remoteHost

	content.binReader = conn
	content.binWriter = io.Discard

	err = content.ReadRequest()
	if err != nil {
		return err
	}
	err = content.BindMethod()
	if err != nil {
		return err
	}
	return handler(content)
}

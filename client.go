/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	encoder "encoding/json"
)

func Put(ctx context.Context, address string, method string, reader io.Reader, binSize int64, param, result any, auth *Auth) error {
	var err error

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		err = fmt.Errorf("unable to resolve adddress: %s", err)
		return err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	return ConnPut(ctx, conn, method, reader, binSize, param, result, auth)
}

func ConnPut(ctx context.Context, conn net.Conn, method string, reader io.Reader, binSize int64, param, result any, auth *Auth) error {
	var err error
	content := CreateContent(conn)

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
	content.binWriter = conn

	content.reqHeader.binSize = binSize

	err = content.createRequest()
	if err != nil {
		return err
	}
	err = content.writeRequest()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	wg.Add(1)
	go content.readResponseAsync(&wg, errChan)

	wg.Add(1)
	go content.uploadBinAsync(ctx, &wg)

	wg.Wait()
	err = <-errChan
	if err != nil {
		return err
	}
	err = content.bindResponse()
	if err != nil {
		return err
	}
	return err
}

func Get(ctx context.Context, address string, method string, writer io.Writer, param, result any, auth *Auth) error {
	var err error

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		err = fmt.Errorf("unable to resolve adddress: %s", err)
		return err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	return ConnGet(ctx, conn, method, writer, param, result, auth)
}

func ConnGet(ctx context.Context, conn net.Conn, method string, writer io.Writer, param, result any, auth *Auth) error {
	var err error

	content := CreateContent(conn)
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

	content.binReader = conn
	content.binWriter = writer

	err = content.createRequest()
	if err != nil {
		return err
	}
	err = content.writeRequest()
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

func Exec(ctx context.Context, address, method string, param any, result any, auth *Auth) error {
	var err error

	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		err = fmt.Errorf("unable to resolve adddress: %s", err)
		return err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	err = ConnExec(ctx, conn, method, param, result, auth)
	if err != nil {
		return err
	}
	return err
}

func ConnExec(ctx context.Context, conn net.Conn, method string, param any, result any, auth *Auth) error {
	var err error

	content := CreateContent(conn)
	content.reqBlock.Method = method

	if param != nil {
		content.reqBlock.Params = param
	}
	if result != nil {
		content.resBlock.Result = result
	}
	if auth != nil {
		content.reqBlock.Auth = auth
	}

	err = content.createRequest()
	if err != nil {
		return err
	}
	err = content.writeRequest()
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

func (content *Content) createRequest() error {
	var err error

	content.reqPacket.rcpPayload, err = content.reqBlock.Pack()
	if err != nil {
		return err
	}
	rpcSize := int64(len(content.reqPacket.rcpPayload))
	content.reqHeader.rpcSize = rpcSize

	content.reqPacket.header, err = content.reqHeader.Pack()
	if err != nil {
		return err
	}
	return err
}

func (content *Content) writeRequest() error {
	var err error
	_, err = content.sockWriter.Write(content.reqPacket.header)
	if err != nil {
		return err
	}
	_, err = content.sockWriter.Write(content.reqPacket.rcpPayload)
	if err != nil {
		return err
	}
	return err
}

func (content *Content) uploadBin(ctx context.Context) error {
	var err error
	_, err = CopyBytes(ctx, content.binReader, content.binWriter, content.reqHeader.binSize)
	return err
}

func (content *Content) readResponse() error {
	var err error

	content.resPacket.header, err = ReadBytes(content.sockReader, headerSize)
	if err != nil {
		return err
	}
	content.resHeader, err = UnpackHeader(content.resPacket.header)
	if err != nil {
		return err
	}
	rpcSize := content.resHeader.rpcSize
	content.resPacket.rcpPayload, err = ReadBytes(content.sockReader, rpcSize)
	if err != nil {
		return err
	}
	return err
}

func (content *Content) uploadBinAsync(ctx context.Context, wg *sync.WaitGroup) {
	exitFunc := func() {
		wg.Done()
	}
	defer exitFunc()
	_, _ = CopyBytes(ctx, content.binReader, content.binWriter, content.reqHeader.binSize)
	return
}

func (content *Content) readResponseAsync(wg *sync.WaitGroup, errChan chan error) {
	var err error
	exitFunc := func() {
		errChan <- err
		wg.Done()
	}
	defer exitFunc()
	content.resPacket.header, err = ReadBytes(content.sockReader, headerSize)
	if err != nil {
		err = err
		return
	}
	content.resHeader, err = UnpackHeader(content.resPacket.header)
	if err != nil {
		err = err
		return
	}
	rpcSize := content.resHeader.rpcSize
	content.resPacket.rcpPayload, err = ReadBytes(content.sockReader, rpcSize)
	if err != nil {
		err = err
		return
	}
	return
}

func (content *Content) downloadBin(ctx context.Context) error {
	var err error
	_, err = CopyBytes(ctx, content.binReader, content.binWriter, content.resHeader.binSize)
	return err
}

func (content *Content) bindResponse() error {
	var err error

	err = encoder.Unmarshal(content.resPacket.rcpPayload, content.resBlock)
	if err != nil {
		return err
	}
	if len(content.resBlock.Error) > 0 {
		err = errors.New(content.resBlock.Error)
		return err
	}
	return err
}

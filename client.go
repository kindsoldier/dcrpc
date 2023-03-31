/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	encoder "github.com/vmihailenco/msgpack/v5"
)

func Put(address string, method string, reader io.Reader, binSize int64, param, result any, auth *Auth) error {
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

	return ConnPut(conn, method, reader, binSize, param, result, auth)
}

func ConnPut(conn net.Conn, method string, reader io.Reader, binSize int64, param, result any, auth *Auth) error {
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

	err = content.CreateRequest()
	if err != nil {
		return err
	}
	err = content.WriteRequest()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	wg.Add(1)
	go content.ReadResponseAsync(&wg, errChan)

	wg.Add(1)
	go content.UploadBinAsync(&wg)

	wg.Wait()
	err = <-errChan
	if err != nil {
		return err
	}
	err = content.BindResponse()
	if err != nil {
		return err
	}
	return err
}

func Get(address string, method string, writer io.Writer, param, result any, auth *Auth) error {
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

	return ConnGet(conn, method, writer, param, result, auth)
}

func ConnGet(conn net.Conn, method string, writer io.Writer, param, result any, auth *Auth) error {
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

	err = content.CreateRequest()
	if err != nil {
		return err
	}
	err = content.WriteRequest()
	if err != nil {
		return err
	}
	err = content.ReadResponse()
	if err != nil {
		return err
	}
	err = content.DownloadBin()
	if err != nil {
		return err
	}
	err = content.BindResponse()
	if err != nil {
		return err
	}
	return err
}

func Exec(address, method string, param any, result any, auth *Auth) error {
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

	err = ConnExec(conn, method, param, result, auth)
	if err != nil {
		return err
	}
	return err
}

func ConnExec(conn net.Conn, method string, param any, result any, auth *Auth) error {
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

	err = content.CreateRequest()
	if err != nil {
		return err
	}
	err = content.WriteRequest()
	if err != nil {
		return err
	}
	err = content.ReadResponse()
	if err != nil {
		return err
	}
	err = content.BindResponse()
	if err != nil {
		return err
	}
	return err
}

func (content *Content) CreateRequest() error {
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

func (content *Content) WriteRequest() error {
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

func (content *Content) UploadBin() error {
	var err error
	_, err = CopyBytes(content.binReader, content.binWriter, content.reqHeader.binSize)
	return err
}

func (content *Content) ReadResponse() error {
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

func (content *Content) UploadBinAsync(wg *sync.WaitGroup) {
	exitFunc := func() {
		wg.Done()
	}
	defer exitFunc()
	_, _ = CopyBytes(content.binReader, content.binWriter, content.reqHeader.binSize)
	return
}

func (content *Content) ReadResponseAsync(wg *sync.WaitGroup, errChan chan error) {
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

func (content *Content) DownloadBin() error {
	var err error
	_, err = CopyBytes(content.binReader, content.binWriter, content.resHeader.binSize)
	return err
}

func (content *Content) BindResponse() error {
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

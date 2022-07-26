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


func Put(address string, method string, reader io.Reader, size int64, param, result any, auth *Auth) error {
    var err error

    addr, err := net.ResolveTCPAddr("tcp", address)
    if err != nil {
        err = fmt.Errorf("unable to resolve adddress: %s", err)
        return Err(err)
    }
    conn, err := net.DialTCP("tcp", nil, addr)
    if err != nil {
        return Err(err)
    }
    defer conn.Close()

    //err = conn.SetKeepAlive(true)
    //if err != nil {
    //    err = fmt.Errorf("unable to set keepalive: %s", err)
    //    return Err(err)
    //}

    //err = conn.SetKeepAlivePeriod(10 * time.Second)
    //if err != nil {
    //    err = fmt.Errorf("unable to set keepalive period: %s", err)
    //    return Err(err)
    //}

    return ConnPut(conn, method, reader, size, param, result, auth)
}


func ConnPut(conn net.Conn, method string, reader io.Reader, size int64, param, result any, auth *Auth) error {
    var err error
    context := CreateContext(conn)
    context.reqRPC.Method = method
    context.reqRPC.Params = param
    context.reqRPC.Auth = auth
    context.resRPC.Result = result

    context.binReader = reader
    context.binWriter = conn

    context.reqHeader.binSize = size

    if context.reqRPC.Params == nil {
        context.reqRPC.Params = NewEmpty()
    }

    err = context.CreateRequest()
    if err != nil {
        return Err(err)
    }
    err = context.WriteRequest()
    if err != nil {
        return Err(err)
    }

    var wg sync.WaitGroup
    errChan := make(chan error, 1)

    wg.Add(1)
    go context.ReadResponseAsync(&wg, errChan)

    wg.Add(1)
    go context.UploadBinAsync(&wg)

    wg.Wait()
    err = <- errChan
    if err != nil {
        return Err(err)
    }
    err = context.BindResponse()
    if err != nil {
        return Err(err)
    }
    return Err(err)
}

func Get(address string, method string, writer io.Writer, param, result any, auth *Auth) error {
    var err error

    addr, err := net.ResolveTCPAddr("tcp", address)
    if err != nil {
        err = fmt.Errorf("unable to resolve adddress: %s", err)
        return Err(err)
    }
    conn, err := net.DialTCP("tcp", nil, addr)
    if err != nil {
        return Err(err)
    }
    defer conn.Close()

    //err = conn.SetKeepAlive(true)
    //if err != nil {
    //    err = fmt.Errorf("unable to set keepalive: %s", err)
    //    return Err(err)
    //}

    //err = conn.SetKeepAlivePeriod(10 * time.Second)
    //if err != nil {
    //    err = fmt.Errorf("unable to set keepalive period: %s", err)
    //    return Err(err)
    //}

    return ConnGet(conn, method, writer, param, result, auth)
}

func ConnGet(conn net.Conn, method string, writer io.Writer, param, result any, auth *Auth) error {
    var err error

    context := CreateContext(conn)
    context.reqRPC.Method = method
    context.reqRPC.Params = param
    context.reqRPC.Auth = auth
    context.resRPC.Result = result

    context.binReader = conn
    context.binWriter = writer

    if context.reqRPC.Params == nil {
        context.reqRPC.Params = NewEmpty()
    }
    err = context.CreateRequest()
    if err != nil {
        return Err(err)
    }
    err = context.WriteRequest()
    if err != nil {
        return Err(err)
    }
    err = context.ReadResponse()
    if err != nil {
        return Err(err)
    }
    err = context.DownloadBin()
    if err != nil {
        return Err(err)
    }
    err = context.BindResponse()
    if err != nil {
        return Err(err)
    }
    return Err(err)
}

func Exec(address, method string, param any, result any, auth *Auth) error {
    var err error

    addr, err := net.ResolveTCPAddr("tcp", address)
    if err != nil {
        err = fmt.Errorf("unable to resolve adddress: %s", err)
        return Err(err)
    }
    conn, err := net.DialTCP("tcp", nil, addr)
    if err != nil {
        return Err(err)
    }
    defer conn.Close()

    //err = conn.SetKeepAlive(true)
    //if err != nil {
    //    err = fmt.Errorf("unable to set keepalive: %s", err)
    //    return Err(err)
    //}
    //err = conn.SetKeepAlivePeriod(10 * time.Second)
    //if err != nil {
    //    err = fmt.Errorf("unable to set keepalive period: %s", err)
    //    return Err(err)
    //}

    err = ConnExec(conn, method, param, result, auth)
    if err != nil {
        return Err(err)
    }
    return Err(err)
}


func ConnExec(conn net.Conn, method string, param any, result any, auth *Auth) error {
    var err error

    context := CreateContext(conn)
    context.reqRPC.Method = method
    context.reqRPC.Params = param
    context.reqRPC.Auth = auth
    context.resRPC.Result = result

    if context.reqRPC.Params == nil {
        context.reqRPC.Params = NewEmpty()
    }

    err = context.CreateRequest()
    if err != nil {
        return Err(err)
    }
    err = context.WriteRequest()
    if err != nil {
        return Err(err)
    }
    err = context.ReadResponse()
    if err != nil {
        return Err(err)
    }
    err = context.BindResponse()
    if err != nil {
        return Err(err)
    }
    return Err(err)
}


func (context *Context) CreateRequest() error {
    var err error

    context.reqPacket.rcpPayload, err = context.reqRPC.Pack()
    if err != nil {
        return Err(err)
    }
    rpcSize := int64(len(context.reqPacket.rcpPayload))
    context.reqHeader.rpcSize = rpcSize

    context.reqPacket.header, err = context.reqHeader.Pack()
    if err != nil {
        return Err(err)
    }
    return Err(err)
}

func (context *Context) WriteRequest() error {
    var err error
    _, err = context.sockWriter.Write(context.reqPacket.header)
    if err != nil {
        return Err(err)
    }
    _, err = context.sockWriter.Write(context.reqPacket.rcpPayload)
    if err != nil {
        return Err(err)
    }
    return Err(err)
}

func (context *Context) UploadBin() error {
    var err error
    _, err = CopyBytes(context.binReader, context.binWriter, context.reqHeader.binSize)
    return Err(err)
}

func (context *Context) ReadResponse() error {
    var err error

    context.resPacket.header, err = ReadBytes(context.sockReader, headerSize)
    if err != nil {
        return Err(err)
    }
    context.resHeader, err = UnpackHeader(context.resPacket.header)
    if err != nil {
        return Err(err)
    }
    rpcSize := context.resHeader.rpcSize
    context.resPacket.rcpPayload, err = ReadBytes(context.sockReader, rpcSize)
    if err != nil {
        return Err(err)
    }
    return Err(err)
}

func (context *Context) UploadBinAsync(wg *sync.WaitGroup) {
    exitFunc := func() {
        wg.Done()
    }
    defer exitFunc()
    _, _ = CopyBytes(context.binReader, context.binWriter, context.reqHeader.binSize)
    return
}

func (context *Context) ReadResponseAsync(wg *sync.WaitGroup, errChan chan error) {
    var err error
    exitFunc := func() {
        errChan <- err
        wg.Done()
    }
    defer exitFunc()
    context.resPacket.header, err = ReadBytes(context.sockReader, headerSize)
    if err != nil {
        err = Err(err)
        return
    }
    context.resHeader, err = UnpackHeader(context.resPacket.header)
    if err != nil {
        err = Err(err)
        return
    }
    rpcSize := context.resHeader.rpcSize
    context.resPacket.rcpPayload, err = ReadBytes(context.sockReader, rpcSize)
    if err != nil {
        err = Err(err)
        return
    }
    return
}

func (context *Context) DownloadBin() error {
    var err error
    _, err = CopyBytes(context.binReader, context.binWriter, context.resHeader.binSize)
    return Err(err)
}

func (context *Context) BindResponse() error {
    var err error

    err = encoder.Unmarshal(context.resPacket.rcpPayload, context.resRPC)
    if err != nil {
        return Err(err)
    }
    if len(context.resRPC.Error) > 0 {
        err = errors.New(context.resRPC.Error)
        return Err(err)
    }
    return Err(err)
}

/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc


import (
    "io"
    "net"
)

func LocalExec(method string, param any, result any, auth *Auth, handler HandlerFunc) error {
    var err error

    cliConn, srvConn := NewFConn()

    context := CreateContext(cliConn)
    context.reqRPC.Method = method
    context.reqRPC.Params = param
    context.reqRPC.Auth = auth
    context.resRPC.Result = result

    if context.reqRPC.Params == nil {
        context.reqRPC.Params = NewEmpty()
    }
    err = context.CreateRequest()
    if err != nil {
        return err
    }
    err = context.WriteRequest()
    if err != nil {
        return err
    }
    err = LocalService(srvConn, handler)
    if err != nil {
        return err
    }
    err = context.ReadResponse()
    if err != nil {
        return err
    }
    err = context.BindResponse()
    if err != nil {
        return err
    }

    return err
}

func LocalPut(method string, reader io.Reader, size int64, param, result any, auth *Auth, handler HandlerFunc) error {

    var err error

    cliConn, srvConn := NewFConn()

    context := CreateContext(cliConn)
    context.reqRPC.Method = method
    context.reqRPC.Params = param
    context.reqRPC.Auth = auth
    context.resRPC.Result = result

    context.binReader = reader
    context.binWriter = cliConn

    context.reqHeader.binSize = size

    if context.reqRPC.Params == nil {
        context.reqRPC.Params = NewEmpty()
    }
    err = context.CreateRequest()
    if err != nil {
        return err
    }
    err = context.WriteRequest()
    if err != nil {
        return err
    }
    err = context.UploadBin()
    if err != nil {
        return err
    }
    err = LocalService(srvConn, handler)
    if err != nil {
        return err
    }
    err = context.ReadResponse()
    if err != nil {
        return err
    }
    err = context.BindResponse()
    if err != nil {
        return err
    }
    return err
}


func LocalGet(method string, writer io.Writer, param, result any, auth *Auth, handler HandlerFunc) error {
    var err error

    cliConn, srvConn := NewFConn()

    context := CreateContext(cliConn)
    context.reqRPC.Method = method
    context.reqRPC.Params = param
    context.reqRPC.Auth = auth
    context.resRPC.Result = result

    context.binReader = cliConn
    context.binWriter = writer

    if context.reqRPC.Params == nil {
        context.reqRPC.Params = NewEmpty()
    }
    err = context.CreateRequest()
    if err != nil {
        return err
    }
    err = context.WriteRequest()
    if err != nil {
        return err
    }

    err = LocalService(srvConn, handler)
    if err != nil {
        return err
    }
    err = context.ReadResponse()
    if err != nil {
        return err
    }
    err = context.DownloadBin()
    if err != nil {
        return err
    }
    err = context.BindResponse()
    if err != nil {
        return err
    }
    return err
}

func LocalService(conn net.Conn, handler HandlerFunc) error {
    var err error
    context := CreateContext(conn)

    remoteAddr := conn.RemoteAddr().String()
    remoteHost, _, _ := net.SplitHostPort(remoteAddr)
    context.remoteHost = remoteHost

    context.binReader = conn
    context.binWriter = io.Discard

    err = context.ReadRequest()
    if err != nil {
        return err
    }
    err = context.BindMethod()
    if err != nil {
        return err
    }
    return handler(context)
}

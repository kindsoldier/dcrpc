/*
 *
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 *
 */

package dsrpc

import (
    "bytes"
    "io"
    "net"
    "time"
)

type FConn struct {
    reader io.Reader
    writer io.Writer
}

func NewFConn() (*FConn, *FConn){
    c2sBuffer := bytes.NewBuffer(make([]byte, 0))
    s2cBuffer := bytes.NewBuffer(make([]byte, 0))

    var client FConn
    client.writer = c2sBuffer
    client.reader = s2cBuffer

    var server FConn
    server.writer = s2cBuffer
    server.reader = c2sBuffer

    return &client, &server
}

func (conn FConn) SetDeadline(t time.Time) error {
    var err error
    return err
}
func (conn FConn) SetReadDeadline(t time.Time) error  {
    var err error
    return err
}
func (conn FConn) SetWriteDeadline(t time.Time) error {
    var err error
    return err
}

func (conn FConn) LocalAddr() net.Addr {
    return NewFAddr()
}

func (conn FConn) RemoteAddr() net.Addr {
    return NewFAddr()
}

func (conn FConn) Write(data []byte) (int, error) {
    return conn.writer.Write(data)
}

func (conn FConn) Read(data []byte) (int, error) {
    return conn.reader.Read(data)
}

func (conn FConn) Close() error {
    var err error
    return err
}

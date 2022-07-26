/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
    "bytes"
    "encoding/json"
    "errors"
    "io"
    "math/rand"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
)

func TestLocalExec(t *testing.T) {
    var err error
    params := NewHelloParams()
    params.Message = "hello server!"
    result := NewHelloResult()

    auth := CreateAuth([]byte("qwert"), []byte("12345"))

    err = LocalExec(HelloMethod, params, result, auth, helloHandler)
    require.NoError(t, err)
    resultJSON, _ := json.Marshal(result)
    logDebug("method result:", string(resultJSON))
}


func TestLocalSave(t *testing.T) {
    var err error

    params := NewSaveParams()
    params.Message = "save data!"
    result := NewHelloResult()
    auth := CreateAuth([]byte("qwert"), []byte("12345"))

    var binSize int64 = 16
    rand.Seed(time.Now().UnixNano())
    binBytes := make([]byte, binSize)
    rand.Read(binBytes)

    reader := bytes.NewReader(binBytes)

    err = LocalPut(SaveMethod, reader, binSize, params, result, auth, saveHandler)
    require.NoError(t, err)

    resultJSON, _ := json.Marshal(result)
    logDebug("method result:", string(resultJSON))
}


func TestLocalLoad(t *testing.T) {
    var err error

    params := NewLoadParams()
    params.Message = "load data!"
    result := NewHelloResult()
    auth := CreateAuth([]byte("qwert"), []byte("12345"))

    binBytes := make([]byte, 0)
    writer := bytes.NewBuffer(binBytes)

    err = LocalGet(LoadMethod, writer, params, result, auth, loadHandler)
    require.NoError(t, err)

    resultJSON, _ := json.Marshal(result)
    logDebug("method result:", string(resultJSON))
    logDebug("bin size:", len(writer.Bytes()))
}


func TestNetExec(t *testing.T) {
    go testServ(false)
    time.Sleep(10 * time.Millisecond)
    err := clientHello()

    require.NoError(t, err)
}

func TestNetSave(t *testing.T) {
    go testServ(false)
    time.Sleep(10 * time.Millisecond)
    err := clientSave()
    require.NoError(t, err)
}

func TestNetLoad(t *testing.T) {
    go testServ(false)
    time.Sleep(10 * time.Millisecond)
    err := clientLoad()
    require.NoError(t, err)
}

func BenchmarkNetPut(b *testing.B) {
    go testServ(true)
    time.Sleep(10 * time.Millisecond)
    clientSave()

    pBench := func(pb *testing.PB) {
        for pb.Next() {
            clientSave()
        }
    }
    b.SetParallelism(10)
    b.RunParallel(pBench)
}

func clientHello() error {
    var err error

    params := NewHelloParams()
    params.Message = "hello server!"
    result := NewHelloResult()
    auth := CreateAuth([]byte("qwert"), []byte("12345"))

    var binSize int64 = 16
    rand.Seed(time.Now().UnixNano())
    binBytes := make([]byte, binSize)
    rand.Read(binBytes)

    err = Exec("127.0.0.1:8081", HelloMethod, params, result, auth)
    if err != nil {
        logError("method err:", err)
        return err
    }
    resultJSON, _ := json.Marshal(result)
    logDebug("method result:", string(resultJSON))
    return err
}


func clientSave() error {
    var err error

    params := NewSaveParams()
    params.Message = "save data!"
    result := NewHelloResult()
    auth := CreateAuth([]byte("qwert"), []byte("12345"))

    var binSize int64 = 16
    rand.Seed(time.Now().UnixNano())
    binBytes := make([]byte, binSize)
    rand.Read(binBytes)

    reader := bytes.NewReader(binBytes)

    err = Put("127.0.0.1:8081", SaveMethod, reader, binSize, params, result, auth)
    if err != nil {
        logError("method err:", err)
        return err
    }
    resultJSON, _ := json.Marshal(result)
    logDebug("method result:", string(resultJSON))
    return err
}


func clientLoad() error {
    var err error

    params := NewLoadParams()
    params.Message = "load data!"
    result := NewHelloResult()
    auth := CreateAuth([]byte("qwert"), []byte("12345"))


    binBytes := make([]byte, 0)
    writer := bytes.NewBuffer(binBytes)

    err = Get("127.0.0.1:8081", LoadMethod, writer, params, result, auth)
    if err != nil {
        logError("method err:", err)
        return err
    }
    resultJSON, _ := json.Marshal(result)
    logDebug("method result:", string(resultJSON))
    logDebug("bin size:", len(writer.Bytes()))
    return err
}


var testServRun bool = false

func testServ(quiet bool) error {
    var err error

    if testServRun {
        return err
    }
    testServRun = true

    if quiet {
        SetAccessWriter(io.Discard)
        SetMessageWriter(io.Discard)
    }
    serv := NewService()
    serv.Handler(HelloMethod, helloHandler)
    serv.Handler(SaveMethod, saveHandler)
    serv.Handler(LoadMethod, loadHandler)

    serv.PreMiddleware(LogRequest)
    serv.PreMiddleware(auth)

    serv.PostMiddleware(LogResponse)
    serv.PostMiddleware(LogAccess)

    err = serv.Listen(":8081")
    if err != nil {
        return err
    }
    return err
}

func auth(context *Context) error {
    var err error
    reqIdent := context.AuthIdent()
    reqSalt := context.AuthSalt()
    reqHash := context.AuthHash()

    ident := reqIdent
    pass := []byte("12345")

    auth := context.Auth()
    logDebug("auth ", string(auth.JSON()))

    ok := CheckHash(ident, pass, reqSalt, reqHash)
    logDebug("auth ok:", ok)
    if !ok {
        err = errors.New("auth ident or pass missmatch")
        context.SendError(err)
        return err
    }
    return err
}

func helloHandler(context *Context) error {
    var err error
    params := NewHelloParams()

    err = context.BindParams(params)
    if err != nil {
        return err
    }

    err = context.ReadBin(io.Discard)
    if err != nil {
        context.SendError(err)
        return err
    }

    result := NewHelloResult()
    result.Message = "hello, client!"

    err = context.SendResult(result, 0)
    if err != nil {
        return err
    }
    return err
}

func saveHandler(context *Context) error {
    var err error
    params := NewSaveParams()

    err = context.BindParams(params)
    if err != nil {
        return err
    }

    bufferBytes := make([]byte, 0, 1024)
    binWriter := bytes.NewBuffer(bufferBytes)

    err = context.ReadBin(binWriter)
    if err != nil {
        context.SendError(err)
        return err
    }

    result := NewSaveResult()
    result.Message = "saved successfully!"

    err = context.SendResult(result, 0)
    if err != nil {
        return err
    }
    return err
}

func loadHandler(context *Context) error {
    var err error
    params := NewSaveParams()

    err = context.BindParams(params)
    if err != nil {
        return err
    }

    err = context.ReadBin(io.Discard)
    if err != nil {
        context.SendError(err)
        return err
    }

    var binSize int64 = 1024
    rand.Seed(time.Now().UnixNano())
    binBytes := make([]byte, binSize)
    rand.Read(binBytes)

    binReader := bytes.NewReader(binBytes)

    result := NewSaveResult()
    result.Message = "load successfully!"

    err = context.SendResult(result, binSize)
    if err != nil {
        return err
    }
    binWriter := context.BinWriter()
    _, err = CopyBytes(binReader, binWriter, binSize)
    if err != nil {
        return err
    }

    return err
}


const HelloMethod string = "hello"

type HelloParams struct {
    Message string      `json:"message" msgpack:"message"`
}

func NewHelloParams() *HelloParams {
    return &HelloParams{}
}

type HelloResult struct {
    Message string      `json:"message" msgpack:"message"`
}

func NewHelloResult() *HelloResult {
    return &HelloResult{}
}


const SaveMethod string = "save"
type SaveParams HelloParams
type SaveResult HelloResult

func NewSaveParams() *SaveParams {
    return &SaveParams{}
}
func NewSaveResult() *SaveResult {
    return &SaveResult{}
}



const LoadMethod string = "load"
type LoadParams HelloParams
type LoadResult HelloResult

func NewLoadParams() *LoadParams {
    return &LoadParams{}
}
func NewLoadResult() *LoadResult {
    return &LoadResult{}
}

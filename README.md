# dsrpc, Data RPC

DSRPC is easy and simple RPC framework over TCP socket.

### Purpose

A very easy and open RPC framework with data streaming.

### You can

- Use own post and pre-execution middleware
- Hash-based authentication in middleware
- Test remote function without network

Socket encryption is not used at this time since framefork
is oriented to transfer large amounts of data.

Style of the framework is similar of GIN framework.

## Exec method example

### Server

```
package main

import (
    "log"
    "github.com/kindsoldier/dsrpc"
    "netsrv/api"
)

func main() {
    err := server()
    if err != nil {
        log.Println(err)
    }
}

func server() error {
    var err error

    serv := dsrpc.NewService()

    cont := NewController()
    serv.Handler(api.HelloMethod, cont.HelloHandler)

    serv.PreMiddleware(dsrpc.LogRequest)
    serv.PostMiddleware(dsrpc.LogResponse)
    serv.PostMiddleware(dsrpc.LogAccess)

    err = serv.Listen(":8081")
    if err != nil {
        return err
    }
    return err
}


type Controller struct {
}

func NewController() *Controller {
    return &Controller{}
}

func (cont *Controller) HelloHandler(context *dsrpc.Context) error {
    var err error
    params := api.NewHelloParams()
    err = context.BindParams(params)
    if err != nil {
        return err
    }

    log.Println("hello message:", params.Message)

    result := api.NewHelloResult()
    result.Message = "hello!"

    err = context.SendResult(result, 0)
    if err != nil {
        return err
    }

    return err
}

```

### Client

```
package main

import (
    "fmt"
    "github.com/kindsoldier/dsrpc"
    "netsrv/api"
)

func main() {

    err := exec()
    if err != nil {
        fmt.Println("exec err:", err)
    }
}

func exec() error {
    var err error

    params := api.NewHelloParams()
    params.Message = "hello, server!"

    result := api.NewHelloResult()

    err = dsrpc.Exec("127.0.0.1:8081", api.HelloMethod, params, result, nil)
    if err != nil {
        return err
    }

    fmt.Println("result:", result.Message)
    return err
}


```

### Common api

```
package api

const HelloMethod string = "hello"

type HelloParams struct {
    Message string      `json:"message"`
}

func NewHelloParams() *HelloParams {
    return &HelloParams{}
}

type HelloResult struct {
    Message string      `json:"message"`
}

func NewHelloResult() *HelloResult {
    return &HelloResult{}
}

```

### Authentication and authorization

#### Client side

```

func clientHello() error {
    var err error

    params := NewHelloParams()
    params.Message = "hello server!"
    result := NewHelloResult()

    auth := dsrpc.CreateAuth([]byte("login"), []byte("password"))

    err = dsrpc.Exec("127.0.0.1:8081", HelloMethod, params, result, auth)
    if err != nil {
        log.Println("method err:", err)
        return err
    }

    //...
}


```

#### Server side

```

func authMiddleware(context *dsrpc.Context) error {
    var err error
    reqIdent := context.AuthIdent()
    reqSalt := context.AuthSalt()
    reqHash := context.AuthHash()

    if reqIdent != "login" {
        err = errors.New("auth ident or pass mismatch")
        context.SendError(err)
        return err
    }

    ident := reqIdent
    pass := []byte("password")

    ok := dsrpc.CheckHash(ident, pass, reqSalt, reqHash)
    log.Println("auth is ok:", ok)
    if !ok {
        err = errors.New("auth ident or pass mismatch")
        context.SendError(err)
        return err
    }
    return err
}

func sampleServ(quiet bool) error {
    var err error

    if quiet {
        dsrpc.SetAccessWriter(io.Discard)
        dsrpc.SetMessageWriter(io.Discard)
    }
    serv := NewService()

    serv.PreMiddleware(authMiddleware)
    serv.PreMiddleware(dsrpc.LogRequest)

    serv.Handler(HelloMethod, helloHandler)
    serv.Handler(SaveMethod, saveHandler)
    serv.Handler(LoadMethod, loadHandler)

    serv.PostMiddleware(dsrpc.LogResponse)
    serv.PostMiddleware(dsrpc.LogAccess)

    err = serv.Listen(":8081")
    if err != nil {
        return err
    }
    return err
}

```

### Put method

#### Client side sample

```
    var binSize int64 = 16
    rand.Seed(time.Now().UnixNano())
    binBytes := make([]byte, binSize)
    rand.Read(binBytes)

    reader := bytes.NewReader(binBytes)

    err = dsrpc.Put("127.0.0.1:8081", SaveMethod, reader, binSize, params, result, auth)

```
#### Server side

```
func saveHandler(context *dsrpc.Context) error {
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

```

### Get method

#### Client side

```
    params := NewLoadParams()
    params.Message = "load data!"
    result := NewHelloResult()
    auth := CreateAuth([]byte("qwert"), []byte("12345"))

    binBytes := make([]byte, 0)
    writer := bytes.NewBuffer(binBytes)

    err = dsrpc.Get("127.0.0.1:8081", LoadMethod, writer, params, result, auth)
    if err != nil {
        return err
    }

    //...

```

#### Server side

```

func getHandler(context *dsrpc.Context) error {
    var err error
    params := NewSaveParams()

    err = context.BindParams(params)
    if err != nil {
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
    _, err = dsrpc.CopyBytes(binReader, binWriter, binSize)
    if err != nil {
        return err
    }

    return err
}

```

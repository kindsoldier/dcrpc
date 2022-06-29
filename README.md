# dsrpc, Data RPC

DSRPC is easy and simple RPC framework over TCP socket.

### Purpose

A very easy and open RPC framework with data streaming. 


### You can 

- Use post and pre-execution middleware
- Hash-based authentication in middleware
- Test call remote function without service organization

Socket encryption is not used at this time since framefork 
is oriented to transfer large amounts of data

Style of the framework is similar to that of GIN framework.

## Example

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
    Message string      `msgpack:"message" json:"message"`
}

func NewHelloParams() *HelloParams {
    return &HelloParams{}
}

type HelloResult struct {
    Message string      `msgpack:"message" json:"message"`
}

func NewHelloResult() *HelloResult {
    return &HelloResult{}
}

```

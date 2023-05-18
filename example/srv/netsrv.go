/*
 *
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 *
 */

package main

import (
    "log"
    "netsrv/api"

    "github.com/kindsoldier/dsrpc"
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
    serv.Handle(api.HelloMethod, cont.HelloHandler)

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

func (cont *Controller) HelloHandler(content *dsrpc.Content) error {
    var err error
    params := api.HelloParams{}
    err = content.BindParams(&params)
    if err != nil {
        return err
    }
    log.Println("hello message:", params.Message)

    result := api.HelloResult{
        Message: "hello, client!",
    }
    err = content.SendResult(&result, 0)
    if err != nil {
        return err
    }

    return err
}

/*
 *
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 *
 */

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

    params := api.HelloParams{
        Message: "hello, server!",
    }

    result := api.HelloResult{}

    err = dsrpc.Exec("127.0.0.1:8081", api.HelloMethod, &params, &result, nil)
    if err != nil {
        return err
    }

    fmt.Println("result:", result.Message)
    return err
}

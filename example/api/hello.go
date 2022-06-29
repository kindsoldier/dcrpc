/*
 *
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 *
 */

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

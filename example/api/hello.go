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

type HelloResult struct {
    Message string      `msgpack:"message" json:"message"`
}


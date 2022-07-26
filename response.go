/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
    "encoding/json"
    encoder "github.com/vmihailenco/msgpack/v5"
)


type Response struct {
    Error   string      `json:"error"   msgpack:"error"`
    Result  any         `json:"result"  msgpack:"result"`
}

func NewResponse() *Response {
    return &Response{}
}

func (resp *Response) JSON() []byte {
    jBytes, _ := json.Marshal(resp)
    return jBytes
}

func (resp *Response) Pack() ([]byte, error) {
    rBytes, err := encoder.Marshal(resp)
    return rBytes, Err(err)
}

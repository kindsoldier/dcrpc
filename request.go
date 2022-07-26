/*
 *
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 *
 */

package dsrpc

import (
    "encoding/json"
    encoder "github.com/vmihailenco/msgpack/v5"
)

type Request struct {
    Method  string      `json:"method"            msgpack:"method"`
    Params  any         `json:"params,omitempty"  msgpack:"params"`
    Auth    *Auth       `json:"auth,omitempty"    msgpack:"auth"`
}

func NewRequest() *Request {
    req := &Request{}
    req.Auth = &Auth{}
    return req
}

func (req *Request) Pack() ([]byte, error) {
    rBytes, err := encoder.Marshal(req)
    return rBytes, Err(err)
}

func (req *Request) JSON() []byte {
    jBytes, _ := json.Marshal(req)
    return jBytes
}

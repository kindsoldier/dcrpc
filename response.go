/*
 *
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 *
 */

package dsrpc

import (
    "encoding/json"
)


type Response struct {
    Error   string      `json:"error,omitempty" msgpack:"error,omitempty"`
    Result  any         `json:"result,omitemty" msgpack:"result,omitemty"`
}

func NewResponse() *Response {
    return &Response{}
}

func (this *Response) JSON() []byte {
    jBytes, _ := json.Marshal(this)
    return jBytes
}

func (this *Response) Pack() ([]byte, error) {
    rBytes, err := json.Marshal(this)
    return rBytes, err
}

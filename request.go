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

type EmptyParams struct{}

func NewEmptyParams() *EmptyParams {
	return &EmptyParams{}
}

type Request struct {
	Method string `json:"method"            msgpack:"method"`
	Params any    `json:"params,omitempty"  msgpack:"params"`
	Auth   *Auth  `json:"auth,omitempty"    msgpack:"auth"`
}

func NewEmptyRequest() *Request {
	req := &Request{}
	req.Auth = &Auth{}
	req.Params = NewEmptyParams()
	return req
}

func (req *Request) Pack() ([]byte, error) {
	rBytes, err := encoder.Marshal(req)
	return rBytes, err
}

func (req *Request) ToJson() []byte {
	jBytes, _ := json.Marshal(req)
	return jBytes
}

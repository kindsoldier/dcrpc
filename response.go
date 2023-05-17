/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
	"encoding/json"

	encoder "encoding/json"
)

type EmptyResult struct{}

func NewEmptyResult() *EmptyResult {
	return &EmptyResult{}
}

type Response struct {
	Error  string `json:"error"   msgpack:"error"`
	Result any    `json:"result"  msgpack:"result"`
}

func NewEmptyResponse() *Response {
	return &Response{
		Result: NewEmptyResult(),
	}
}

func (resp *Response) ToJson() []byte {
	jBytes, _ := json.Marshal(resp)
	return jBytes
}

func (resp *Response) Pack() ([]byte, error) {
	rBytes, err := encoder.Marshal(resp)
	return rBytes, err
}

/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
	"time"
)

func LogRequest(content *Content) error {
	var err error
	logDebug("request:", string(content.reqBlock.ToJson()))
	return err
}

func LogResponse(content *Content) error {
	var err error
	logDebug("response:", string(content.resBlock.ToJson()))
	return err
}

func LogAccess(content *Content) error {
	var err error
	execTime := time.Now().Sub(content.start)
	login := string(content.AuthIdent())
	logAccess(content.remoteHost, login, content.reqBlock.Method, execTime)
	return err
}

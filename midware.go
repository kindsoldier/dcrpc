/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
    "time"
)

func LogRequest(context *Context) error {
    var err error
    logDebug("request:", string(context.reqRPC.JSON()))
    return Err(err)
}

func LogResponse(context *Context) error {
    var err error
    logDebug("response:", string(context.resRPC.JSON()))
    return Err(err)
}

func LogAccess(context *Context) error {
    var err error
    execTime := time.Now().Sub(context.start)
    login := string(context.AuthIdent())
    logAccess(context.remoteHost, login, context.reqRPC.Method, execTime)
    return Err(err)
}

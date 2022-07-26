/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
    "fmt"
    "runtime"
    "io"
)

var develMode bool = false
var debugMode bool = false


func SetDevelMode(mode bool) {
    develMode = mode
}
func SetDebugMode(mode bool) {
    debugMode = mode
}

func Err(err error) error {
    switch err {
        case io.EOF:
            return err
    }
    if err != nil {
        switch {
            case develMode == true:
                pc, filename, line, _ := runtime.Caller(1)
                funcName := runtime.FuncForPC(pc).Name()
                err = fmt.Errorf(" %s:%d:%s:%s", filename, line, funcName, err.Error())
            case debugMode == true:
                pc, _, line, _ := runtime.Caller(1)
                funcName := runtime.FuncForPC(pc).Name()
                err = fmt.Errorf(" %s:%d:%s ", funcName, line, err.Error())
            default:
        }
    }
    return err
}

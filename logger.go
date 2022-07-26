/*
 *
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 *
 */

package dsrpc

import (
    "fmt"
    "io"
    "os"
    "time"
)

var messageWriter io.Writer = os.Stdout
var accessWriter io.Writer = os.Stdout

func logDebug(messages ...any) {
    stamp := time.Now().Format(time.RFC3339)
    fmt.Fprintln(messageWriter, stamp, "debug", messages)
}

func logInfo(messages ...any) {
    stamp := time.Now().Format(time.RFC3339)
    fmt.Fprintln(messageWriter, stamp, "info", messages)
}

func logError(messages ...any) {
    stamp := time.Now().Format(time.RFC3339)
    fmt.Fprintln(messageWriter, stamp, "error", messages)
}

func logAccess(messages ...any) {
    stamp := time.Now().Format(time.RFC3339)
    fmt.Fprintln(accessWriter, stamp, "access", messages)
}

func SetAccessWriter(writer io.Writer) {
    accessWriter = writer
}

func SetMessageWriter(writer io.Writer) {
    messageWriter = writer
}

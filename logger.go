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

var (
	messageWriter io.Writer = os.Stdout
	accessWriter  io.Writer = os.Stdout
	logTimestamp  bool      = false
)

func getLogStamp() string {
	var stamp string
	if logTimestamp {
		stamp = time.Now().Format(time.RFC3339)
	}
	return stamp
}

func logDebug(messages ...any) {
	stamp := getLogStamp()
	fmt.Fprintln(messageWriter, stamp, "debug", messages)
}

func logInfo(messages ...any) {
	stamp := getLogStamp()
	fmt.Fprintln(messageWriter, stamp, "info", messages)
}

func logError(messages ...any) {
	stamp := getLogStamp()
	fmt.Fprintln(messageWriter, stamp, "error", messages)
}

func logAccess(messages ...any) {
	stamp := getLogStamp()
	fmt.Fprintln(accessWriter, stamp, "access", messages)
}

func SetAccessWriter(writer io.Writer) {
	accessWriter = writer
}

func SetMessageWriter(writer io.Writer) {
	messageWriter = writer
}

func EnableLogTimestamp(enable bool) {
	logTimestamp = enable
}


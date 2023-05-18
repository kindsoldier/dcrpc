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
)

var (
	messageWriter io.Writer = os.Stdout
	accessWriter  io.Writer = os.Stdout
)

func logDebug(messages ...any) {
	fmt.Fprintln(messageWriter, "debug:", messages)
}

func logInfo(messages ...any) {
	fmt.Fprintln(messageWriter, "info:", messages)
}

func logError(messages ...any) {
	fmt.Fprintln(messageWriter, "error:", messages)
}

func logAccess(messages ...any) {
	fmt.Fprintln(accessWriter, "access:", messages)
}

func SetAccessWriter(writer io.Writer) {
	accessWriter = writer
}

func SetMessageWriter(writer io.Writer) {
	messageWriter = writer
}

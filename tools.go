/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
	"context"
	"errors"
	"fmt"
	"io"
)

func ReadBytes(reader io.Reader, size int64) ([]byte, error) {
	buffer := make([]byte, size)
	read, err := io.ReadFull(reader, buffer)
	return buffer[0:read], err
}

func CopyBytes(ctx context.Context, reader io.Reader, writer io.Writer, dataSize int64) (int64, error) {
	var err error
	var bSize int64 = 1024 * 16
	var total int64 = 0
	var remains int64 = dataSize
	buffer := make([]byte, bSize)

	for {
		select {
		case <-ctx.Done():
			return total, errors.New("break by context")
		default:
		}

		if reader == nil {
			return total, errors.New("reader is nil")
		}
		if writer == nil {
			return total, errors.New("writer is nil")
		}
		if remains == 0 {
			return total, err
		}
		if remains < bSize {
			bSize = remains
		}
		received, err := reader.Read(buffer[0:bSize])
		if err != nil {
			err = fmt.Errorf("read error: %v", err)
			return total, err
		}
		recorded, err := writer.Write(buffer[0:received])
		if err != nil {
			err = fmt.Errorf("write error: %v", err)
			return total, err
		}
		if recorded != received {
			err = errors.New("size mismatch")
			return total, err
		}
		total += int64(recorded)
		remains -= int64(recorded)
	}
	return total, err
}

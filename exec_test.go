/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const HelloMethod string = "hello"

type HelloParams struct {
	Message string `json:"message" msgpack:"message"`
}

type HelloResult struct {
	Message string `json:"message" msgpack:"message"`
}

const SaveMethod string = "save"

type SaveParams HelloParams
type SaveResult HelloResult

const LoadMethod string = "load"

type LoadParams HelloParams
type LoadResult HelloResult

func TestLocalExec(t *testing.T) {
	var err error
	params := HelloParams{}
	params.Message = "hello server!"
	result := HelloResult{}

	auth := CreateAuth([]byte("qwert"), []byte("12345"))

	err = LocalExec(HelloMethod, &params, &result, auth, helloHandler)
	require.NoError(t, err)

	resultJson, _ := json.Marshal(result)
	logDebug("method result:", string(resultJson))
}

func TestLocalSave(t *testing.T) {
	var err error

	params := SaveParams{}
	params.Message = "save data!"
	result := SaveResult{}
	auth := CreateAuth([]byte("qwert"), []byte("12345"))

	var binSize int64 = 16
	rand.Seed(time.Now().UnixNano())
	binBytes := make([]byte, binSize)
	rand.Read(binBytes)

	reader := bytes.NewReader(binBytes)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	err = LocalPut(ctx, SaveMethod, reader, binSize, &params, &result, auth, saveHandler)
	require.NoError(t, err)

	resultJson, _ := json.Marshal(result)
	logDebug("method result:", string(resultJson))
}

func TestLocalLoad(t *testing.T) {
	var err error

	params := LoadParams{}
	params.Message = "load data!"
	result := LoadResult{}
	auth := CreateAuth([]byte("qwert"), []byte("12345"))

	binBytes := make([]byte, 0)
	writer := bytes.NewBuffer(binBytes)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	err = LocalGet(ctx, LoadMethod, writer, &params, &result, auth, loadHandler)
	require.NoError(t, err)

	resultJson, _ := json.Marshal(result)
	logDebug("method result:", string(resultJson))
	logDebug("bin size:", len(writer.Bytes()))
}

func TestNetExec(t *testing.T) {
	go testServ(false)
	time.Sleep(10 * time.Millisecond)
	err := clientHello()

	require.NoError(t, err)
}

func TestNetSave(t *testing.T) {
	go testServ(false)
	time.Sleep(10 * time.Millisecond)
	err := clientSave()
	require.NoError(t, err)
}

func TestNetLoad(t *testing.T) {
	go testServ(false)
	time.Sleep(10 * time.Millisecond)
	err := clientLoad()
	require.NoError(t, err)
}

func BenchmarkNetPut(b *testing.B) {
	go testServ(true)
	time.Sleep(10 * time.Millisecond)
	clientSave()

	pBench := func(pb *testing.PB) {
		for pb.Next() {
			clientSave()
		}
	}
	b.SetParallelism(2000)
	b.RunParallel(pBench)
}

func clientHello() error {
	var err error

	params := HelloParams{}
	params.Message = "hello server!"
	result := HelloResult{}
	auth := CreateAuth([]byte("qwert"), []byte("12345"))

	var binSize int64 = 16
	rand.Seed(time.Now().UnixNano())
	binBytes := make([]byte, binSize)
	rand.Read(binBytes)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	err = Exec(ctx, "127.0.0.1:8081", HelloMethod, &params, &result, auth)
	if err != nil {
		logError("method err:", err)
		return err
	}
	resultJson, _ := json.Marshal(result)
	logDebug("method result:", string(resultJson))
	return err
}

func clientSave() error {
	var err error

	params := SaveParams{}
	params.Message = "save data!"
	result := SaveResult{}
	auth := CreateAuth([]byte("qwert"), []byte("12345"))

	var binSize int64 = 16
	rand.Seed(time.Now().UnixNano())
	binBytes := make([]byte, binSize)
	rand.Read(binBytes)

	reader := bytes.NewReader(binBytes)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	err = Put(ctx, "127.0.0.1:8081", SaveMethod, reader, binSize, &params, &result, auth)
	if err != nil {
		logError("method err:", err)
		return err
	}
	resultJson, _ := json.Marshal(result)
	logDebug("method result:", string(resultJson))
	return err
}

func clientLoad() error {
	var err error

	params := LoadParams{}
	params.Message = "load data!"
	result := LoadResult{}
	auth := CreateAuth([]byte("qwert"), []byte("12345"))

	binBytes := make([]byte, 0)
	writer := bytes.NewBuffer(binBytes)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	err = Get(ctx, "127.0.0.1:8081", LoadMethod, writer, &params, &result, auth)
	if err != nil {
		logError("method err:", err)
		return err
	}
	resultJson, _ := json.Marshal(result)
	logDebug("method result:", string(resultJson))
	logDebug("bin size:", len(writer.Bytes()))
	return err
}

var testServRun bool = false

func testServ(quiet bool) error {
	var err error

	if testServRun {
		return err
	}
	testServRun = true

	if quiet {
		SetAccessWriter(io.Discard)
		SetMessageWriter(io.Discard)
	}
	serv := NewService()
	serv.Handle(HelloMethod, helloHandler)
	serv.Handle(SaveMethod, saveHandler)
	serv.Handle(LoadMethod, loadHandler)

	serv.PreMiddleware(LogRequest)
	serv.PreMiddleware(auth)

	serv.PostMiddleware(LogResponse)
	serv.PostMiddleware(LogAccess)

	err = serv.Listen(":8081")
	if err != nil {
		return err
	}
	return err
}

func auth(content *Content) error {
	var err error
	reqIdent := content.AuthIdent()
	reqSalt := content.AuthSalt()
	reqHash := content.AuthHash()

	ident := reqIdent
	pass := []byte("12345")

	auth := content.Auth()
	logDebug("auth ", string(auth.Json()))

	ok := CheckHash(ident, pass, reqSalt, reqHash)
	logDebug("auth ok:", ok)
	if !ok {
		err = errors.New("auth ident or pass missmatch")
		content.SendError(err)
		return err
	}
	return err
}

func helloHandler(content *Content) error {
	var err error
	params := HelloParams{}

	err = content.BindParams(&params)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	err = content.ReadBin(ctx, io.Discard)
	if err != nil {
		content.SendError(err)
		return err
	}

	result := HelloResult{}
	result.Message = "hello, client!"

	err = content.SendResult(result, 0)
	if err != nil {
		return err
	}
	return err
}

func saveHandler(content *Content) error {
	var err error
	params := SaveParams{}

	err = content.BindParams(&params)
	if err != nil {
		return err
	}

	bufferBytes := make([]byte, 0, 1024)
	binWriter := bytes.NewBuffer(bufferBytes)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	err = content.ReadBin(ctx, binWriter)
	if err != nil {
		content.SendError(err)
		return err
	}

	result := SaveResult{}
	result.Message = "saved successfully!"

	err = content.SendResult(result, 0)
	if err != nil {
		return err
	}
	return err
}

func loadHandler(content *Content) error {
	var err error
	params := SaveParams{}

	err = content.BindParams(&params)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	err = content.ReadBin(ctx, io.Discard)
	if err != nil {
		content.SendError(err)
		return err
	}

	var binSize int64 = 1024
	rand.Seed(time.Now().UnixNano())
	binBytes := make([]byte, binSize)
	rand.Read(binBytes)

	binReader := bytes.NewReader(binBytes)

	result := SaveResult{}
	result.Message = "load successfully!"

	err = content.SendResult(result, binSize)
	if err != nil {
		return err
	}

	binWriter := content.BinWriter()
	_, err = CopyBytes(ctx, binReader, binWriter, binSize)
	if err != nil {
		return err
	}

	return err
}

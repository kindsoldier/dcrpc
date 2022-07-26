/*
 *
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 *
 */

package dsrpc

import (
    "errors"
    "encoding/binary"
    "encoding/json"
    "bytes"
)

const headerSize    int64   = 16 * 2
const sizeOfInt64   int     = 8
const magicCodeA    int64   = 0xEE00ABBA
const magicCodeB    int64   = 0xEE44ABBA

type Header struct {
    magicCodeA  int64   `json:"magicCodeA"`
    rpcSize     int64   `json:"rpcSize"`
    binSize     int64   `json:"binSize"`
    magicCodeB  int64   `json:"magicCodeB"`
}


func NewHeader() *Header {
    return &Header{
        magicCodeA: magicCodeA,
        magicCodeB: magicCodeB,
    }
}

func (hdr *Header) JSON() []byte {
    jBytes, _ := json.Marshal(hdr)
    return jBytes
}


func (hdr *Header) Pack() ([]byte, error) {
    var err error
    headerBytes := make([]byte, 0, headerSize)
    headerBuffer := bytes.NewBuffer(headerBytes)

    magicCodeABytes := encoderI64(hdr.magicCodeA)
    headerBuffer.Write(magicCodeABytes)

    rpcSizeBytes := encoderI64(hdr.rpcSize)
    headerBuffer.Write(rpcSizeBytes)

    binSizeBytes := encoderI64(hdr.binSize)
    headerBuffer.Write(binSizeBytes)

    magicCodeBBytes := encoderI64(hdr.magicCodeB)
    headerBuffer.Write(magicCodeBBytes)

    return headerBuffer.Bytes(), Err(err)
}

func UnpackHeader(headerBytes []byte) (*Header, error) {
    var err error
    header := NewHeader()
    headerReader := bytes.NewReader(headerBytes)

    magicCodeABytes := make([]byte, sizeOfInt64)
    headerReader.Read(magicCodeABytes)
    header.magicCodeA = decoderI64(magicCodeABytes)

    rpcSizeBytes := make([]byte, sizeOfInt64)
    headerReader.Read(rpcSizeBytes)
    header.rpcSize = decoderI64(rpcSizeBytes)

    binSizeBytes := make([]byte, sizeOfInt64)
    headerReader.Read(binSizeBytes)
    header.binSize = decoderI64(binSizeBytes)

    magicCodeBBytes := make([]byte, sizeOfInt64)
    headerReader.Read(magicCodeBBytes)
    header.magicCodeB = decoderI64(magicCodeBBytes)

    if header.magicCodeA != magicCodeA || header.magicCodeB != magicCodeB {
        err = errors.New("wrong protocol magic code")
        return header, Err(err)
    }
    return header, Err(err)
}

func encoderI64(i int64) []byte {
    buffer := make([]byte, sizeOfInt64)
    binary.BigEndian.PutUint64(buffer, uint64(i))
    return buffer
}

func decoderI64(b []byte) int64 {
    return int64(binary.BigEndian.Uint64(b))
}

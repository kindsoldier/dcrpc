/*
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 */

package dsrpc

import (
    "encoding/json"
    "bytes"
    "math/rand"
    "time"
    "crypto/sha256"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

type Auth struct {
    Ident   []byte      `msgpack:"ident"    json:"ident"`
    Salt    []byte      `msgpack:"salt"     json:"salt"`
    Hash    []byte      `msgpack:"hash"     json:"hash"`
}

func NewAuth() *Auth {
    return &Auth{}
}

func (this *Auth) JSON() []byte {
    jBytes, _ := json.Marshal(this)
    return jBytes
}

func CreateAuth(ident, pass []byte) *Auth {
    salt := CreateSalt()
    hash := CreateHash(ident, pass, salt)
    auth := &Auth{}
    auth.Ident = ident
    auth.Salt = salt
    auth.Hash = hash
    return auth
}

func CreateSalt() []byte {
    const saltSize = 16
    randBytes := make([]byte, saltSize)
    rand.Read(randBytes)
    return randBytes
}

func CreateHash(ident, pass, salt []byte) []byte {
    vec := make([]byte, 0, len(ident) + len(salt) + len(pass))
    vec = append(vec, ident...)
    vec = append(vec, salt...)
    vec = append(vec, pass...)
    hasher := sha256.New()
    hash := hasher.Sum(vec)
    return hash
}

func CheckHash(ident, pass, reqSalt, reqHash []byte) bool {
    localHash := CreateHash(ident, pass, reqSalt)
    return bytes.Equal(reqHash, localHash)
}

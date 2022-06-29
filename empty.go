/*
 *
 * Copyright 2022 Oleg Borodin  <borodin@unix7.org>
 *
 */

package dsrpc

type Empty struct {}

func NewEmpty() *Empty {
    return &Empty{}
}

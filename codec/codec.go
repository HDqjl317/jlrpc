/*
 * @Author: jiale_quan jiale_quan@ustc.edu
 * @Date: 2023-04-01 15:53:14
 * @LastEditTime: 2023-04-01 15:58:16
 * @Description:
 * Copyright © jiale_quan, All Rights Reserved
 */
package codec

import "io"

type Header struct {
	ServiceMethod string
	Seq           uint64
	Error         string
}

type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

const (
	GobType Type = "application/gob"
	Json    Type = "application/json"
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}

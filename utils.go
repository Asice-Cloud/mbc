package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

func IntToBytes(value int64) []byte {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, value); err != nil {
		log.Panicln(err)
	}

	return buf.Bytes()
}

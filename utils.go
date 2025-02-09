package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"reflect"
)

func IntToBytes(value int64) []byte {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, value); err != nil {
		log.Panicln(err)
	}

	return buf.Bytes()
}

func Reverse[T any](ts []T) []T {
	for i, j := 0, len(ts)-1; i < j; i, j = i+1, j-1 {
		ts[i], ts[j] = ts[j], ts[i]
	}
	return ts
}

// try-catch-finally implementation(((
type ErrorHolder struct {
	err error
}

func throw(err error) {
	panic(ErrorHolder{err: err})
}

type FinallyHandler interface {
	finally(handlers ...func())
}

type CatchHandler interface {
	catch(e error, handler func(err error)) CatchHandler
	catch_all(handler func(err error)) FinallyHandler
	FinallyHandler
}

type catchHandler struct {
	err      error
	hasCatch bool
}

func try(f func()) CatchHandler {
	t := &catchHandler{}
	defer func() {
		if r := recover(); r != nil {
			if eh, ok := r.(ErrorHolder); ok {
				t.err = eh.err
			} else if err, ok := r.(error); ok {
				t.err = err
			} else {
				t.err = fmt.Errorf("unknown panic: %v", r)
			}
			t.hasCatch = false
		}
	}()
	f()
	return t
}

func (t *catchHandler) catch(e error, handler func(err error)) CatchHandler {
	if t.err != nil && !t.hasCatch && reflect.TypeOf(t.err) == reflect.TypeOf(e) {
		handler(t.err)
		t.hasCatch = true
	}
	return t
}

func (t *catchHandler) catch_all(handler func(err error)) FinallyHandler {
	if t.err != nil && !t.hasCatch {
		handler(t.err)
	}
	return t
}

func (t *catchHandler) finally(handlers ...func()) {
	for _, handler := range handlers {
		handler()
	}
}

package main

import "testing"

func Test_al(t *testing.T) {
	t.Log("algo_test.go")
	t.Log(B58Encode([]byte("hello")))
	t.Log(B58Decode([]byte("hello")))
}

func Bench_al(b *testing.B) {
	for i := 0; i < b.N; i++ {
		B58Encode([]byte("hello"))
	}
	for i := 0; i < b.N; i++ {
		B58Decode([]byte("hello"))
	}
}

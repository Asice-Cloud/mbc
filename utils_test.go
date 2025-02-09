package main

import (
	"github.com/boltdb/bolt"
	"os"
	"testing"
)

func Test_utils(t *testing.T) {
	try(
		func() {
			db, err := bolt.Open("test.db", 0600, nil)
			if err != nil {
				throw(err)
			}
			defer db.Close()

			db.Update(
				func(tx *bolt.Tx) error {
					bucket := tx.Bucket([]byte("test"))
					if bucket == nil {
						bucket, err := tx.CreateBucket([]byte("test"))
						if err != nil {
							throw(err)
						}
						bucket.Put([]byte("key"), []byte("value"))
					}
					return nil
				},
			)

			db.View(
				func(tx *bolt.Tx) error {
					tx.Bucket([]byte("test")).ForEach(
						func(k, v []byte) error {
							if string(k) != "key" || string(v) != "value" {
								t.Error("key or value is not correct")
							}
							return nil
						})
					return nil
				},
			)
		},
	).catch_all(func(err error) {
		t.Error(err)
	}).finally(
		func() {
			err := os.Remove("test.db")
			if err != nil {
				t.Error(err)
			}
		},
	)
}

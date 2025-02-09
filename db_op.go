package main

import "os"

func db_exists(db_fd string) bool {
	if _, err := os.Stat(db_fd); os.IsNotExist(err) {
		return false
	}
	return true
}

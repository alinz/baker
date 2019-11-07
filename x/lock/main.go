package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var mu sync.RWMutex

	go func() {
		for {
			mu.Lock()
			fmt.Println("1", time.Now())
			mu.Unlock()
		}
	}()

	go func() {
		for {
			mu.Lock()
			fmt.Println("2", time.Now())
			mu.Unlock()
		}
	}()

	time.Sleep(time.Second * 2)
}

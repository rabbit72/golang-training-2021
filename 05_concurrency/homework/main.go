package main

import (
	"fmt"
	"github.com/rabbit72/golang-training-2021/05_concurrency/homework/cache"
	"time"
)

func main() {
	ticker := time.NewTicker(1 * time.Second)
	myCache := cache.NewSimpleCache(500 * time.Millisecond)
	defer myCache.Stop()
	myCache.Set("key1", "value1", 3*time.Second)
	myCache.Set("key2", "value3", 6*time.Second)
	timeOut := time.After(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			fmt.Println(myCache)
		case <-timeOut:
			myCache.Stop()
		}
	}
}

package main

import (
	"fmt"
	"lycCache/myCache"
	"time"
)

// Test 模拟缓存数据结构体
type Test struct {
	data string
	time time.Time
}

func main() {
	fmt.Println("hello")
	mycache := myCache.NewCache[string, Test](func(t Test) time.Time {
		return t.time
	})
	mycache.Set("aaa", Test{data: "heelo world", time: time.Now()}, 10*time.Second)
	data, flag := mycache.Get("aaa")
	fmt.Println("111111111", data, flag)
	time.Sleep(15 * time.Second)
	mycache.Get("aaa")
	fmt.Println("2222222222", data, flag)
}

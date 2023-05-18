package main

import (
	"context"
	"fmt"
	"log"

	"hurpc/gredis"
)

func main() {
	ctx := context.Background()

	rdb := gredis.NewClient("tcp", ":6379")

	// Set string value
	resp, err := rdb.StrSet(ctx, "key1", "hello world ajaaa!").Result()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(resp)

	// Get string value
	key1val, err := rdb.StrGet(ctx, "key1").Result()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(key1val)

}

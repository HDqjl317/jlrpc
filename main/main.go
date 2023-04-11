/*
 * @Author: jiale_quan jiale_quan@ustc.edu
 * @Date: 2023-04-11 13:56:24
 * @LastEditTime: 2023-04-11 15:57:59
 * @Description:
 * Copyright © jiale_quan, All Rights Reserved
 */
package main

import (
	"fmt"
	"jlrpc"
	"log"
	"net"
	"sync"
	"time"
)

func startServer(addr chan string) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	jlrpc.Accept(l)
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)

	client, _ := jlrpc.Dial("tcp", <-addr)
	defer func() {
		_ = client.Close()
	}()

	time.Sleep(time.Second)
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := fmt.Sprintf("jlrpc req %d", i)
			var reply string
			if err := client.Call("Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Println("reply:", reply)
		}(i)
	}
	wg.Wait()
}

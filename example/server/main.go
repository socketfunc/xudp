package main

import (
	"fmt"
	"log"

	"github.com/socketfunc/xudp"
)

func main() {
	ln, err := xudp.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	for {
		sess, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("Session", sess)
		go func() {
			for {
				buf, err := sess.Receive()
				if err != nil {
					fmt.Printf("%v\n", err)
					continue
				}
				fmt.Println(buf)
			}
		}()
	}
}

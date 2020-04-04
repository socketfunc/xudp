package main

import (
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/socketfunc/xudp"
	helloworld "github.com/socketfunc/xudp/example/proto"
)

func main() {
	sess, err := xudp.Dial("udp4", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	singer := &helloworld.Singer{
		SingerId:   100,
		FirstName:  "AAA",
		LastName:   "BBB",
		SingerInfo: []byte{0, 0, 0, 1},
	}
	singers := make([]*helloworld.Singer, 0, 100)
	for i := 0; i < 100; i++ {
		singers = append(singers, singer)
	}
	req := &helloworld.BulkInsertSingersRequest{
		Singers: singers,
	}
	buf, err := proto.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}
	if err := sess.Send(buf); err != nil {
		log.Fatal(err)
	}

	sess.Keepalive()

	t := time.NewTicker(3 * time.Second)
	defer t.Stop()
	for {
		select {
		case now := <-t.C:
			sess.Ping()
			fmt.Println(now)
		}
	}
}

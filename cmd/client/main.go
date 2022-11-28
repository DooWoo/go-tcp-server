package main

import (
	"fmt"
	"github.com/DooWoo/go-tcp-server/frame"
	"github.com/DooWoo/go-tcp-server/packet"
	"github.com/lucasepe/codename"
	"net"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	var num int = 5
	wg.Add(num)
	for i:=0; i< num; i++ {
		go func(i int) {
			defer wg.Done()
			startClient(i)
		}(i+1)
	}
	wg.Wait()
}

func startClient(i int) {
	quit := make(chan struct{})
	done := make(chan struct{})
	conn, err := net.Dial("tcp", ":8881")
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}
	defer conn.Close()

	rng, err := codename.DefaultRNG()
	if err != nil {
		panic(err)
	}
	frameCodec := frame.NewMyFrameCodec()
	var counter int
	go func() {
		for {
			select {
			case <-quit:
				done <- struct{}{}
				return
			default:
			}

			conn.SetReadDeadline(time.Now().Add(time.Second * 10))
			ackFramePayload, err := frameCodec.Decode(conn)
			if err != nil {
				if e, ok := err.(net.Error); ok {
					if e.Timeout() {
						continue
					}
					panic(err)
				}
			}
			p, err := packet.Decode(ackFramePayload)
			submitAck, ok := p.(*packet.SubmitAck)
			if !ok {
				panic("not submitAck")
			}
			fmt.Printf("[client %d]: the result of submit ack[%s] is %d\n", i, submitAck.ID, submitAck.Result)
		}
	}()

	for {
		counter++
		id := fmt.Sprintf("%08d", counter)
		payload := codename.Generate(rng, 4)
		s:= &packet.Submit{
			ID:      id,
			Payload: []byte(payload),
		}
		framePayload, err := packet.Encode(s)
		if err != nil {
			panic(err)
		}
		fmt.Printf("[client %d]: send submit id = %s, payload=%s, frame length = %d\n",
			i, s.ID, s.Payload, len(framePayload)+4)
		err = frameCodec.Encode(conn, framePayload)
		if err != nil {
			panic(err)
		}

		time.Sleep(1*time.Second)
		if counter >= 10 {
			quit <- struct{}{}
			<- done
			return
		}
	}
}

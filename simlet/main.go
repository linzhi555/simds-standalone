package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"simds-standalone/config"
	"simds-standalone/core"
	"simds-standalone/simlet/svc"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var NodeState *core.Node
var Lock sync.Mutex
var input core.Vec[core.Message]

type server struct {
}

func (s *server) SendMessage(ctx context.Context, msg *svc.Message) (*svc.Response, error) {
	log.Println(msg.From)
	input.InQueue(core.Message{})
	return &svc.Response{OK: true, ErrMsg: "asdf"}, nil
}

func serving() {

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", 8888))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	svc.RegisterSimletServerServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func clienting() {
	start := time.Now()
	conn, err := grpc.Dial("simds-node0-svc:8888", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	log.Println(time.Since(start))
	c := svc.NewSimletServerClient(conn)
	log.Println(time.Since(start))

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SendMessage(ctx, &svc.Message{From: config.Val.NodeName})
	if err != nil {
		log.Fatalf("could not get result: %v", err)
	}
	log.Println(time.Since(start))
	log.Printf("result %s", r)

	r2, err := c.SendMessage(ctx, &svc.Message{From: config.Val.NodeName})
	if err != nil {
		log.Fatalf("could not get result: %v", err)
	}
	log.Println(time.Since(start))
	log.Printf("result %s", r2)

}

func main() {
	log.Println("simlet started as", config.Val.NodeName)

	if config.Val.NodeName == "simds-node0" {
		serving()
	} else {
		start := time.Now()
		clienting()
		log.Println(time.Since(start))
	}

}

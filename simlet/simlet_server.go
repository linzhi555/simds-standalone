package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"simds-standalone/cluster/base"
	"simds-standalone/common"
	"simds-standalone/simlet/svc"
	"simds-standalone/tracing/rules"
)

type simletCli struct {
	addr     string
	cliAlive bool
	cli      svc.SimletServerClient
}

type SimletServer struct {
	routerTable *common.ConcurrentMap[string, simletCli]
	actorIns    *common.ConcurrentMap[string, chan base.Message]
	actorOut    chan base.Message
}

func NewServer() *SimletServer {
	return &SimletServer{
		routerTable: common.NewConcurrentMap[string, simletCli](),
		actorIns:    common.NewConcurrentMap[string, chan base.Message](),
		actorOut:    make(chan base.Message, 100000),
	}
}

// NewServer with initil routerTable
func NewServerWithRouterTable(table *svc.RouterTable) *SimletServer {
	server := NewServer()
	server._updateRouterTable(table)
	return server
}

func (s *SimletServer) _updateRouterTable(table *svc.RouterTable) {
	for _, newPair := range table.Columns {
		old, ok := s.routerTable.Load(newPair.ActorAddr)
		if ok && old.addr == newPair.SimletAddr {
			continue
		}
		s.routerTable.Store(newPair.ActorAddr, simletCli{addr: newPair.SimletAddr, cliAlive: false})
		log.Println("routerTable updated", newPair.ActorAddr, newPair.SimletAddr)
	}
}

// Register a new Actor
// just add in/out pair channel to manage
// Make the new actor'output be sended on time
// make the new actor get message on time when somebody send message to it
func (s *SimletServer) RegisterNewActor(actor *ActorOs) {
	actor.input = make(chan base.Message, 10000)
	s.actorIns.Store(actor.name, actor.input)
	actor.output = s.actorOut
}

// Run input server
// make the message's from other simlet be redirected to the actor' input channel
// update the routerTable when receive UpdateRouterTable request
func (s *SimletServer) RunInputServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", 8888))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	svc.RegisterSimletServerServer(grpcServer, s)
	log.Printf("input server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	panic("server finished")
}

// Run output thread
// make the message in the actor's shared output channel be send parallely.
func (s *SimletServer) RunOutputThread() {
	for {
		msg := <-s.actorOut
		if ch, ok := s.actorIns.Load(msg.To); ok {
			ch <- msg
			rules.CheckRulesThenExec(rules.RecvRules, time.Now(), &msg)
		} else {
			go s.doRouting(msg) // different simlet' communication
		}

	}
}

func (s *SimletServer) doRouting(m base.Message) {
	client, ok := s.routerTable.Load(m.To)
	if !ok {
		log.Println("can not find the target router for", m.To, "messge Type is", m.Head)
		return
	}

	// create the conn when cache miss
	if !client.cliAlive {
		start := time.Now()
		conn, err := grpc.Dial(client.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Println("establish new conn fail", m.To, client.addr, err)
			return
		} else {
			log.Println("establish", client.addr, "rpc connection spend", time.Since(start))
			cli := svc.NewSimletServerClient(conn)
			client.cli = cli
			client.cliAlive = true
			s.routerTable.Store(m.To, client)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := client.cli.SendMessage(ctx, &svc.Message{Id: m.Id, From: m.From, To: m.To, Content: m.Head, Body: base.ToJson(m.Body)})
	if err != nil {
		log.Println("could not get result: ", err, m)
	}
}

func (s *SimletServer) SendMessage(ctx context.Context, msg *svc.Message) (*svc.Response, error) {
	body := base.FromJson(msg.Content, msg.Body)

	ch, ok := s.actorIns.Load(msg.To)

	if !ok {
		e := "the actor is not here"
		return &svc.Response{OK: false, ErrMsg: e}, errors.New(e)
	}

	newMsg := base.Message{
		Id:   msg.Id,
		From: msg.From,
		To:   msg.To,
		Head: msg.Content,
		Body: body,
	}

	ch <- newMsg

	rules.CheckRulesThenExec(rules.RecvRules, time.Now(), &newMsg)
	//_logMsg("recv", &newMsg)

	return &svc.Response{OK: true, ErrMsg: "null"}, nil
}

func (s *SimletServer) UpdateRouterTable(ctx context.Context, table *svc.RouterTable) (*svc.Response, error) {

	s._updateRouterTable(table)
	return &svc.Response{OK: true, ErrMsg: "null"}, nil
}

// 提供抽象的os接口给actor模型使用
type ActorOs struct {
	name   string
	input  chan base.Message
	output chan base.Message
}

func NewActorOs(name string) *ActorOs {
	return &ActorOs{
		name: name,
	}
}

// GetTime 提供模拟时间
func (actor *ActorOs) GetTime() time.Time {
	return time.Now()
}

func (o *ActorOs) Send(msg base.Message) {
	msg.Id = string(common.GenerateUID())
	o.output <- msg
	rules.CheckRulesThenExec(rules.SendRules, time.Now(), &msg)
}

func (o *ActorOs) SetInterval(callback func(), t time.Duration) {
	go func() {
		for {
			time.Sleep(t)
			callback()
		}
	}()
}

func (o *ActorOs) SetTimeOut(callback func(), t time.Duration) {
	go func() {
		time.Sleep(t)
		callback()
	}()
}

func (o *ActorOs) RunCmd(callback func(err error), cmdstr string) {
	go func() {
		cmd := exec.Command("bash", "-c", cmdstr)
		err := cmd.Run()
		callback(err)
	}()
}

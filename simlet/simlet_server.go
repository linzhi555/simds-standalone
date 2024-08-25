package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"path"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"simds-standalone/cluster/base"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/simlet/svc"
)

const NETWORK_EVENT_LOG_NAME = "network_event.log"

type simletCli struct {
	addr     string
	cliAlive bool
	cli      svc.SimletServerClient
}

type initFlag struct {
	inited bool
	sync.RWMutex
}

func (flag *initFlag) setTrue() {
	flag.Lock()
	defer flag.Unlock()
	flag.inited = true
}

func (flag *initFlag) isTrue() bool {
	flag.RLock()
	defer flag.RUnlock()
	return flag.inited
}

type SimletServer struct {
	inited      initFlag
	routerTable *common.ConcurrentMap[string, simletCli]
	actorIns    *common.ConcurrentMap[string, chan base.Message]
	actorOut    chan base.Message
}

func NewServer() *SimletServer {
	return &SimletServer{
		inited:      initFlag{inited: false},
		routerTable: common.NewConcurrentMap[string, simletCli](),
		actorIns:    common.NewConcurrentMap[string, chan base.Message](),
		actorOut:    make(chan base.Message, 100000),
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
		message := <-s.actorOut
		if ch, ok := s.actorIns.Load(message.To); ok {
			ch <- message
		} else {
			go s.doRouting(message) // different simlet' communication
		}

	}
}

func (s *SimletServer) doRouting(m base.Message) {
	client, ok := s.routerTable.Load(m.To)
	if !ok {
		log.Println("can not find the target router for", m.To, "messge Type is", m.Content)
		return
	}

	// create the conn when cache miss
	if client.cliAlive == false {
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
	resp, err := client.cli.SendMessage(ctx, &svc.Message{From: m.From, To: m.To, Content: m.Content, Body: base.ToJson(m.Body)})
	if err != nil {
		log.Println("could not get result: ", err, m, resp.ErrMsg)
	}
}

func (s *SimletServer) SendMessage(ctx context.Context, msg *svc.Message) (*svc.Response, error) {
	body := base.FromJson(msg.Content, msg.Body)

	ch, ok := s.actorIns.Load(msg.To)

	if !ok {
		e := "the actor is not here"
		return &svc.Response{OK: false, ErrMsg: e}, errors.New(e)
	}

	ch <- base.Message{
		From:    msg.From,
		To:      msg.To,
		Content: msg.Content,
		Body:    body,
	}

	bodystring := fmt.Sprint(msg.Body)
	if len(bodystring) > 100 {
		bodystring = bodystring[0:97] + "..."
	}
	_logInfo(NETWORK_EVENT_LOG_NAME, msg.Content, msg.From, msg.To, bodystring)

	log.Println(msg.Content, msg.From, msg.To, body)

	return &svc.Response{OK: true, ErrMsg: "null"}, nil
}

func (s *SimletServer) UpdateRouterTable(ctx context.Context, table *svc.RouterTable) (*svc.Response, error) {
	for _, newPair := range table.Columns {
		old, ok := s.routerTable.Load(newPair.ActorAddr)
		if ok && old.addr == newPair.SimletAddr {
			continue
		}
		s.routerTable.Store(newPair.ActorAddr, simletCli{addr: newPair.SimletAddr, cliAlive: false})
		log.Println("routerTable updated", newPair.ActorAddr, newPair.SimletAddr)
	}
	s.inited.setTrue()
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

func (o *ActorOs) Run(f func()) {
	go f()
}

func (o *ActorOs) Send(m base.Message) error {
	o.output <- m
	return nil
}

func (o *ActorOs) LogInfo(out string, items ...string) {
	_logInfo(out, items...)
}

func _logInfo(out string, items ...string) {
	timestr := time.Now().Format(time.RFC3339Nano)
	s := ""
	if out == "stdout" {
		s += fmt.Sprint(timestr)
		for _, item := range items {
			s += ","
			s += fmt.Sprint(item)
		}
		fmt.Println(s)
	} else {
		line := append([]string{timestr}, items...)
		err := common.AppendLineCsvFile(path.Join(config.Val.OutputDir, out), line)
		if err != nil {
			panic(err)
		}
	}
}

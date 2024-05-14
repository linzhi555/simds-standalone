package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"path"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/core"
	"simds-standalone/simlet/svc"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var NodeState core.Node
var NodeInput *Input = &Input{}

type Input struct {
	sync.Mutex
	buffer core.Vec[core.Message]
}

func (input *Input) Recv() (core.Message, error) {
	input.Lock()
	defer input.Unlock()
	return input.buffer.Dequeue()
}

func (input *Input) InQueue(m core.Message) {
	input.Lock()
	defer input.Unlock()
	input.buffer.InQueue(m)
}

func (input *Input) HasMessage() bool {
	input.Lock()
	defer input.Unlock()
	return !input.buffer.Empty()
}

type server struct {
}

const NETWORK_EVENT_LOG_NAME = "network_event.log"

func (s *server) SendMessage(ctx context.Context, msg *svc.Message) (*svc.Response, error) {
	NodeInput.InQueue(core.Message{
		From:    msg.From,
		To:      msg.To,
		Content: msg.Content,
		Body:    core.FromJson(msg.Content, msg.Body),
	})
	_logInfo(NETWORK_EVENT_LOG_NAME, msg.Content, msg.From, msg.To)
	return &svc.Response{OK: true, ErrMsg: "null"}, nil
}

func InputServing() {

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
	panic("server finished")
}

// func clienting() {
// 	start := time.Now()
// 	conn, err := grpc.Dial("simds-node0-svc:8888", grpc.WithTransportCredentials(insecure.NewCredentials()))
// 	if err != nil {
// 		log.Fatalf("did not connect: %v", err)
// 	}
// 	defer conn.Close()
// 	log.Println(time.Since(start))
// 	c := svc.NewSimletServerClient(conn)
// 	log.Println(time.Since(start))

// 	// Contact the server and print out its response.
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()

// 	r, err := c.SendMessage(ctx, &svc.Message{From: config.Val.NodeName})
// 	if err != nil {
// 		log.Fatalf("could not get result: %v", err)
// 	}
// 	log.Println(time.Since(start))
// 	log.Printf("result %s", r)

// 	r2, err := c.SendMessage(ctx, &svc.Message{From: config.Val.NodeName})
// 	if err != nil {
// 		log.Fatalf("could not get result: %v", err)
// 	}
// 	log.Println(time.Since(start))
// 	log.Printf("result %s", r2)

// }

type NodeOs struct {
	clients map[string]svc.SimletServerClient
}

func NewNodeOs() *NodeOs {
	return &NodeOs{clients: make(map[string]svc.SimletServerClient)}
}

// GetTime 提供模拟时间
func (Node *NodeOs) GetTime() time.Time {
	return time.Now()
}

func (o *NodeOs) Run(f func()) {
	go f()
}

func (o *NodeOs) HasMessage() bool {
	return NodeInput.HasMessage()
}

func (o *NodeOs) Recv() (core.Message, error) {
	return NodeInput.Recv()
}

func (o *NodeOs) Send(m core.Message) error {
	if _, ok := o.clients[m.To]; !ok {
		start := time.Now()
		conn, err := grpc.Dial(fmt.Sprintf("simds-%s-svc:8888", m.To), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		log.Println("establish rpc connection spend", time.Since(start))
		c := svc.NewSimletServerClient(conn)
		o.clients[m.To] = c
	}
	c := o.clients[m.To]
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := c.SendMessage(ctx, &svc.Message{From: m.From, To: m.To, Content: m.Content, Body: core.ToJson(m.Body)})
	if err != nil {
		log.Println("could not get result: ", err, m)
	}
	return nil
}

func (o *NodeOs) LogInfo(out string, items ...string) {
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

func testGrpc() {
	start := time.Now()
	conn, err := grpc.Dial("127.0.0.1:8888", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	log.Println("establish rpc connection spend", time.Since(start))
	c := svc.NewSimletServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.SendMessage(ctx, &svc.Message{From: "asdfds", To: "sdf", Content: "TaskSubmit", Body: core.ToJson(core.TaskInfo{})})
	if err != nil {
		panic(err)
	}
}

func main() {
	log.Println("simlet started as", config.Val.NodeName)

	// Init log file
	common.AppendLineCsvFile(NETWORK_EVENT_LOG_NAME, []string{"time", "type", "from", "to"})
	common.AppendLineCsvFile(core.TASKS_EVENT_LOG_NAME, []string{"time", "taskid", "type", "nodeip", "cpu", "ram"})
	go InputServing()

	// core.InitLogs()
	config.LogConfig(path.Join(config.Val.OutputDir, "config.log"))

	// Initialize self as a specified node of cluster
	clusterBuilder, ok := core.ClusterMarket[config.Val.Cluster]
	if !ok {
		keys := make([]string, 0, len(core.ClusterMarket))
		for k := range core.ClusterMarket {
			keys = append(keys, k)
		}
		log.Panicln("wrong type of cluster,registed cluster is", keys)
	}

	var cluster core.Cluster = clusterBuilder()
	for _, n := range cluster.Nodes {
		if n.GetHostName() == config.Val.NodeName {
			NodeState = n
			break
		}
	}
	NodeState.SetOsApi(NewNodeOs())

	// Init input channel
	time.Sleep(time.Second * 3)
	// run node
	for {
		timer1 := time.NewTimer(time.Millisecond)
		NodeState.Update()
		<-timer1.C
	}

}

package core

import (
	"fmt"
	"math/rand"
	"time"
)

const ALL_NODE_NUM = 100

// 创造一个Raft集群
func BuildRaftCluster() Cluster {
	var cluster = createCluster()

	var nodes []Node
	for i := 0; i < int(ALL_NODE_NUM); i++ {
		nodeName := fmt.Sprintf("node%d", i)
		nodes = append(nodes, Node{
			nodeName,
			[]NodeComponent{
				NewRaftManager(nodeName),
			},
		})
	}
	cluster.Nodes = nodes
	cluster.RegisterFunc(CRaftManager, RaftSetup, RaftUpdate)

	return cluster
}

func RaftSetup(c interface{}) {
	raft := c.(*RaftManager)
	raft.Role = Follower
	raft.IsBroken = false
	raft.AllNodeNum = ALL_NODE_NUM
	raft.Term = 0
	raft.LastHeartBeat = raft.Os.GetTime()
	raft.LeaderTimeout = time.Millisecond * time.Duration(1000+rand.Intn(3000))
	LogInfo(raft.Os, raft.Os.Net().GetAddr(), "started", raft.LeaderTimeout)

}

func RaftUpdate(c interface{}) {
	raft := c.(*RaftManager)

	//编程性故障模拟
	if raft.Role == Leader {
		if raft.Os.GetTime().Sub(raft.LeaderTime) > 2*time.Second {
			raft.IsBroken = true
		}
	}

	if raft.IsBroken {
		return
	}

	switch raft.Role {
	case Leader:
		raftLeaderHandle(raft)
	case Follower:
		raftFollowerHandle(raft)
	case Candidate:
		raftCandidateHandle(raft)
	default:
		panic("role error" + raft.Role)
	}

}

func raftLeaderHandle(raft *RaftManager) {
	SinceLastHeartBeat := raft.Os.GetTime().Sub(raft.LastHeartBeat)
	// HeartBeat for 0.5s
	if SinceLastHeartBeat > time.Millisecond*500 {
		for i := 0; i < raft.AllNodeNum; i++ {
			distNode := fmt.Sprintf("node%d:RaftManager", i)
			if distNode == raft.Host {
				continue
			}
			err := raft.Os.Net().Send(Message{
				From:    raft.Os.Net().GetAddr(),
				To:      distNode,
				Content: "HeartBeat",
				Body:    raftNodeInfo{term: raft.Term},
			})
			if err != nil {
				panic(err)
			}
		}
		raft.LastHeartBeat = raft.Os.GetTime()
	}
}

type raftNodeInfo struct {
	term int
}

func (raftNodeInfo) MessageBody() {}

func raftFollowerHandle(raft *RaftManager) {
	SinceLastHeartBeat := raft.Os.GetTime().Sub(raft.LastHeartBeat)

	for !raft.Os.Net().Empty() {
		newMessage, err := raft.Os.Net().Recv()
		if err != nil {
			panic(err)
		}

		if newMessage.Content == "HeartBeat" {
			term := newMessage.Body.(raftNodeInfo).term
			if term >= raft.Term {
				raft.Term = term
				raft.LastHeartBeat = raft.Os.GetTime()
			}
		}

		if newMessage.Content == "PullVotesRequest" {
			term := newMessage.Body.(raftNodeInfo).term
			peer := newMessage.From
			reply := ""
			if term <= raft.Term {
				reply = "VoteNo"
			} else {
				reply = "VoteYes"
				raft.LastHeartBeat = raft.Os.GetTime()
				raft.Term = term
			}
			err := raft.Os.Net().Send(Message{
				From:    raft.Os.Net().GetAddr(),
				To:      peer,
				Content: reply,
				Body:    raftNodeInfo{term: raft.Term},
			})
			if err != nil {
				panic(err)
			}

		}
	}

	if SinceLastHeartBeat > raft.LeaderTimeout {
		raft.Term += 1
		raft.Role = Candidate
		LogInfo(raft.Os, "Now ,I am the new candidate")
		for i := 0; i < raft.AllNodeNum; i++ {
			distNode := fmt.Sprintf("node%d:RaftManager", i)
			if distNode == raft.Host {
				continue
			}
			err := raft.Os.Net().Send(Message{
				From:    raft.Os.Net().GetAddr(),
				To:      distNode,
				Content: "PullVotesRequest",
				Body:    raftNodeInfo{term: raft.Term},
			})
			if err != nil {
				panic(err)
			}
		}
		raft.ReceiveYES = 1
		raft.ReceiveNO = 0
	}
}

func raftCandidateHandle(raft *RaftManager) {
	for !raft.Os.Net().Empty() {
		newMessage, err := raft.Os.Net().Recv()
		if err != nil {
			panic(err)
		}

		if newMessage.Content == "VoteYes" {
			term := newMessage.Body.(raftNodeInfo).term
			if term == raft.Term {
				raft.Term = term
				raft.LastHeartBeat = raft.Os.GetTime()
				raft.ReceiveYES++
			}
		}
		if newMessage.Content == "VoteNo" {
			term := newMessage.Body.(raftNodeInfo).term
			if term == raft.Term {
				raft.Term = term
				raft.LastHeartBeat = raft.Os.GetTime()
				raft.ReceiveNO++
			}
		}

		if newMessage.Content == "HeartBeat" {
			term := newMessage.Body.(raftNodeInfo).term
			if term > raft.Term {
				raft.Term = term
				raft.Role = Follower
				raft.LastHeartBeat = raft.Os.GetTime()
			}
		}

	}

	if raft.ReceiveYES > raft.AllNodeNum/2 {
		LogInfo(raft.Os, "Now ,I am the new leader")
		raft.LeaderTime = raft.Os.GetTime()
		raft.Role = Leader
	}
}

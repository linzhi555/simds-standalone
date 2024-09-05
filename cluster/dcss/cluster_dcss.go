package dcss

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"simds-standalone/cluster/base"
	"simds-standalone/cluster/lib"
	"simds-standalone/config"
)

func BuildDcssCluster() base.Cluster {

	var cluster base.Cluster
	taskgen0 := lib.NewTaskGen("simds-taskgen0")

	for i := 0; i < int(config.Val.NodeNum); i++ {
		actorName := fmt.Sprintf("simds-node%d", i)
		actor := NewDcssNode(actorName)
		actor.setup()
		cluster.Join(base.NewNode(actor))
		taskgen0.Receivers = append(taskgen0.Receivers, actorName)
	}

	cluster.Join(base.NewNode(taskgen0))
	return cluster
}
func (node *DcssNode) setup() {
	// init local node info
	node.LocalNode = &lib.NodeInfo{Addr: node.GetHostName(), Cpu: config.Val.NodeCpu, Memory: config.Val.NodeMemory, CpuAllocted: 0, MemoryAllocted: 0}
	node.TaskMap = make(map[string]*lib.TaskInfo)
	node.RunningTask = make(map[string]*lib.TaskInfo)

	// init neibors
	neiborNum := int(config.Val.DcssNeibor)
	allNodeNum := int(config.Val.NodeNum)
	neiborRandom := int(config.Val.DcssNeiborRandomP * (float32(config.Val.DcssNeibor)))

	var neibors []string = make([]string, 0, neiborNum)

	selfIndex, err := strconv.Atoi(strings.TrimLeft(node.Host, "simds-node"))
	if err != nil {
		panic(err)
	}

	for _, neiborIndex := range getNeigbor(allNodeNum, selfIndex, neiborNum, neiborRandom) {
		newNeibor := fmt.Sprintf("simds-node%d", neiborIndex)
		neibors = append(neibors, newNeibor)
	}

	for _, n := range neibors {
		nodeInfo := &lib.NodeInfo{Addr: n, Cpu: config.Val.NodeCpu, Memory: config.Val.NodeMemory, CpuAllocted: 0, MemoryAllocted: 0}
		node.Neighbors[n] = nodeInfo.Clone()
	}

}

// 创建邻域的算法，输入 一个节点的编号（selfIndex ） 返回其领域的一系列编号
func getNeigbor(allNodes int, selfIndex int, neiborNum int, neiborRandom int) []int {
	if neiborNum <= 2 {
		panic("neigborNum can not smaller than 2 ")
	}

	if allNodes < neiborNum*2 {
		panic("allNodes num can not smaller than neigborNum*2")
	}

	if selfIndex < 0 || selfIndex > allNodes-1 {
		panic("wrong index")
	}

	if neiborRandom > neiborNum {
		panic("to many random neigbor")
	}

	neinodes := make([]int, 0)
	res := make(map[int]struct{})
	// 规则邻居的编号
	leftNei := 0
	rightNei := 0
	for q := 0; q < neiborNum-neiborRandom; q++ {
		if q%2 == 0 {
			//加左邻居
			nextLeft := (selfIndex - leftNei - 1 + allNodes) % allNodes
			res[nextLeft] = struct{}{}
			leftNei++
		} else {
			nextRight := (selfIndex + rightNei + 1 + allNodes) % allNodes
			res[nextRight] = struct{}{}
			rightNei++
		}
	}
	// 随机邻居的编号
	for len(res) < neiborNum {
		index := rand.Intn(allNodes)
		if index == selfIndex {
			continue
		}
		res[index] = struct{}{}
	}
	for k := range res {
		neinodes = append(neinodes, k)
	}
	return neinodes
}

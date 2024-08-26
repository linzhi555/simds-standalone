package dcss

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"simds-standalone/cluster/base"
	"simds-standalone/config"
)

func BuildDcssCluster() base.Cluster {

	var nodes []base.Node
	taskgen0 := base.NewTaskGen("simds-taskgen0")

	for i := 0; i < int(config.Val.NodeNum); i++ {
		nodeName := fmt.Sprintf("simds-node%d", i)
		newNode := DcssNode{
			BasicNode: base.BasicNode{
				Host: nodeName,
			},
			Neighbors: make(map[string]*base.NodeInfo),
			LocalNode: &base.NodeInfo{nodeName, config.Val.NodeCpu, config.Val.NodeMemory, 0, 0},
		}
		newNode.setup()
		nodes = append(nodes, &newNode)
		taskgen0.Receivers = append(taskgen0.Receivers, newNode.GetHostName())
	}
	nodes = append(nodes, taskgen0)
	return base.Cluster{Nodes: nodes}

}
func (node *DcssNode) setup() {
	// init local node info
	node.LocalNode = &base.NodeInfo{node.GetHostName(), config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}
	node.TaskMap = make(map[string]*base.TaskInfo)
	node.RunningTask = make(map[string]*base.TaskInfo)

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
		nodeInfo := &base.NodeInfo{n, config.Val.NodeCpu, config.Val.NodeMemory, 0, 0}
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

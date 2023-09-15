package main

import (
	"fmt"
	"strings"
)

var gloabT int
type node struct{
	Host string
	Cpu int
	Ram int
	CpuUsed int
	RamUsed int
	Neibor []string
	MessageCh string 
}

var allnodes map[string]*node =make(map[string]*node)
var timeEvent map[int] []func() = make(map[int][]func())

func (n *node)Update(){
	if strings.Contains(n.MessageCh,"newtask"){
		fmt.Println(n.Host,"forward the task to ",fmt.Sprint(n.Neibor[0]))
		sendtask(n.Neibor[0],"newtask",10)
		timeEvent[gloabT+1] =append(timeEvent[gloabT+1], func() {
			n.MessageCh=" "
		})


	}
}

func sendtask(host string,m string,delay int){
	timeEvent[gloabT+delay] = append(timeEvent[gloabT+delay], 
	func() {
		allnodes[host].MessageCh=fmt.Sprint(gloabT,"_")+m
	},
	)
}

func debugNodesMessges(){
	fmt.Println("VVVVVVVVVVV")
	for host,v := range allnodes{
		fmt.Println(host,v.Neibor,v.MessageCh)

	}
	fmt.Println("^^^^^^^^^^")

}






func NewNode(host string,cpu,ram int)*node{
	return &node {
		Host: host,
		Cpu:cpu,
		Ram: ram,
		CpuUsed:0,
		RamUsed:0,
		Neibor: make([]string,0),
		MessageCh: "",
	}
}

func main(){

	var allnum int = 3
	for i:=0;i<allnum;i++{
		nnode := NewNode("worker"+fmt.Sprint(i),10,10)
		nnode.Neibor = append(nnode.Neibor, "worker"+fmt.Sprint((i+1)%allnum))
		allnodes[nnode.Host]=nnode
	}

	for ;gloabT<100;gloabT++{
		fmt.Println("time: ",gloabT)
		debugNodesMessges()
		if timeEvent[gloabT] !=nil{
			for _,f := range timeEvent[gloabT]{
				f()
			}
		}
		for _,n := range allnodes{
			n.Update()
		}

		if gloabT==2{
			allnodes["worker0"].MessageCh="newtask"
		
		}


	}





}

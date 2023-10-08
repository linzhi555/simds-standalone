package main

type Node struct {
	Name       string
	Components []NodeComponent
}

type Cluster struct {
	Nodes   []Node
	Setups  map[ComponentName]func(interface{})
	Updates map[ComponentName]func(interface{})
}

func createCluster() Cluster {
	return Cluster{
		Nodes:   make([]Node, 0),
		Setups:  make(map[ComponentName]func(interface{})),
		Updates: make(map[ComponentName]func(interface{})),
	}

}

func (cluster *Cluster) RegisterFunc(c ComponentName, setup, update func(interface{})) {

	cluster.Setups[c] = setup
	cluster.Updates[c] = update

}

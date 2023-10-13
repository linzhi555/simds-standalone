package main

import "fmt"

func BuildShareStateCluster()Cluster{
	var  cluster Cluster = createCluster()
	
	message := Message{
		From:"asdf",
		To:"asdf",
		Content:"asdf",
		Body:Vec[TaskInfo]{TaskInfo{},
		},
	}
	fmt.Println(message)


	return cluster
}


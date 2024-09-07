package engine

import (
	"fmt"
	"log"
	"simds-standalone/cluster/base"
)

type ActorDebugInfo struct {
	Name       string `json:"name"`
	Node       string `json:"node"`
	IsBusy     string `json:"isBusy"`
	Progress   string `json:"progress"`
	Msg        string `json:"msg"`
	Difficulty string `json:"difficulty"`
	Data       string `json:"data"`
}

func (engine *Engine) DebugNodes() []ActorDebugInfo {
	var res []ActorDebugInfo
	for _, node := range engine.nodes() {
		for _, actor := range node.actors {
			res = append(res, ActorDebugInfo{
				Name:     actor.model.GetAddress(),
				Node:     actor.model.GetAddress(),
				IsBusy:   fmt.Sprint(actor.hide.IsBusy),
				Progress: fmt.Sprint(actor.hide.Progress.toFloat()),

				Msg: func() string {
					if actor.hide.LastMsg == nil {
						return "null"
					}
					return actor.hide.LastMsg.Head
				}(),

				Difficulty: fmt.Sprint(actor.hide.Difficulty),
			})
		}
	}
	return res
}

type MessageDebugInfo struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Head     string `json:"head"`
	Body     string `json:"body"`
	LeftTime string `json:"leftTime"`
}

type NetDebugInfo struct {
	Waitings []MessageDebugInfo `json:"waittings"`
	Sended   []MessageDebugInfo `json:"sended"`
}

func (engine *Engine) DebugNet() NetDebugInfo {
	var res NetDebugInfo

	addMsg := func(target *[]MessageDebugInfo, msg *base.Message) {
		*target = append(*target, MessageDebugInfo{
			From:     msg.From,
			To:       msg.To,
			Head:     msg.Head,
			Body:     "",
			LeftTime: fmt.Sprint(msg.LeftTime),
		})
	}

	for i := range engine.Network.Waittings {
		addMsg(&res.Waitings, &engine.Network.Waittings[i])
	}

	for _, outs := range engine.Network.Outs {
		for i := range *outs {
			addMsg(&res.Sended, &(*outs)[i])
			fmt.Println((*outs)[i].Head)
		}
	}

	log.Println(len(res.Sended), len(res.Waitings))

	return res
}

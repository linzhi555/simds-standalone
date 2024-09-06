// in this file, we define what message we tracing in simulation
// the actual log behavior is implement by the two kind engine
package rules

import (
	"fmt"
	"path"
	"strings"
	"time"

	"simds-standalone/cluster/base"
	"simds-standalone/cluster/lib"
	"simds-standalone/common"
	"simds-standalone/config"
	"simds-standalone/tracing/analyzer"
)

// for engin to init tracing
func InitTracing() {
	err := common.AppendLineCsvFile(path.Join(config.Val.OutputDir, config.Val.NetEventsLogName), analyzer.NET_EVENT_LOG_HEAD)
	if err != nil {
		panic(err)
	}
	err = common.AppendLineCsvFile(path.Join(config.Val.OutputDir, config.Val.TaskEventsLogName), analyzer.TASK_EVENT_LOG_HEAD)
	if err != nil {
		panic(err)
	}

}

type Rule struct {
	MessageType string
	FromActorID string
	ToActorID   string
	Handle      func(time.Time, *base.Message)
}

func CheckRulesThenExec(rules []Rule, t time.Time, msg *base.Message) {
	for i := range rules {
		if isRuleMatch(&rules[i], msg.Head, msg.From, msg.To) {
			rules[i].Handle(t, msg)
		}
	}
}

func isRuleMatch(r *Rule, messageType string, from string, to string) bool {
	if !common.MatchPattern(r.FromActorID, from) {
		return false
	}

	if !common.MatchPattern(r.ToActorID, to) {
		return false
	}

	if !common.MatchPattern(r.MessageType, messageType) {
		return false
	}

	return true
}

// apply the tracing rule when the actor recv message
var RecvRules []Rule = []Rule{
	{
		MessageType: "*",
		FromActorID: "*",
		ToActorID:   "*",
		Handle:      NetRecvEventRecord,
	},

	{
		MessageType: "TaskStart",
		FromActorID: "*",
		ToActorID:   "simds-taskgen*",
		Handle:      TaskEventRecord,
	},

	{
		MessageType: "TaskFinish",
		FromActorID: "*",
		ToActorID:   "simds-taskgen*",
		Handle:      TaskEventRecord,
	},
}

// apply the tracing rule when some actor send message
var SendRules []Rule = []Rule{
	{
		MessageType: "*",
		FromActorID: "*",
		ToActorID:   "*",
		Handle:      NetSendEventRecord,
	},

	{
		MessageType: "TaskDispense",
		FromActorID: "simds-taskgen*",
		ToActorID:   "*",
		Handle:      TaskEventRecord,
	},
}

// apply the tracing rule when some actor start deal message
var MsgDealRules []Rule = []Rule{}

// apply the tracing rule when some actor finishe deal message
var MsgFinishRules []Rule = []Rule{}

func TaskEventRecord(t time.Time, msg *base.Message) {
	task := msg.Body.(lib.TaskInfo)
	if strings.HasSuffix(task.Id, "preheat") {
		return
	}

	err := common.AppendLineCsvFile(
		path.Join(config.Val.OutputDir, config.Val.TaskEventsLogName),
		[]string{
			common.FormatTime(t),
			msg.Head,
			task.Id,
			msg.From,
			fmt.Sprint(task.CpuRequest),
			fmt.Sprint(task.MemoryRequest),
		},
	)
	if err != nil {
		panic(err)
	}

}

func NetSendEventRecord(t time.Time, msg *base.Message) {
	_netEventRecord(t, msg, "send")
}

func NetRecvEventRecord(t time.Time, msg *base.Message) {
	_netEventRecord(t, msg, "recv")
}

func _netEventRecord(t time.Time, msg *base.Message, eventype string) {
	switch v := msg.Body.(type) {
	case lib.TaskInfo:
		if strings.HasSuffix(v.Id, "preheat") {
			return
		}
	}

	err := common.AppendLineCsvFile(
		path.Join(config.Val.OutputDir, config.Val.NetEventsLogName),
		[]string{
			common.FormatTime(t),
			msg.Id,
			eventype,
			msg.Head,
			msg.From,
			msg.To,
		},
	)
	if err != nil {
		panic(err)
	}

}

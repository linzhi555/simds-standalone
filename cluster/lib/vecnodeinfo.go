package lib

import (
	"encoding/json"
	"log"

	"simds-standalone/cluster/base"
	"simds-standalone/common"
)

type VecNodeInfo common.Vec[NodeInfo]

func (VecNodeInfo) MessageBody() {}

func init() {
	base.InverseJsonTable["VecNodeInfo*"] = func(s string) base.MessageBody {
		var res []NodeInfo
		err := json.Unmarshal([]byte(s), &res)
		if err != nil {
			log.Panicln("try to unmarshal VecNodeInfo:", s, err)
		}
		return VecNodeInfo(res)
	}

}

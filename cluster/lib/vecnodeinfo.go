package lib

import (
	"encoding/json"
	"simds-standalone/cluster/base"
	"simds-standalone/common"
)

type VecNodeInfo common.Vec[NodeInfo]

func (VecNodeInfo) MessageBody() {}

func init() {
	base.InverseJsonTable["VecNodeInfo*"] = func(s string) base.MessageBody {
		var res VecNodeInfo
		err := json.Unmarshal([]byte(s), &res)
		if err != nil {
			panic(err)
		}
		return res
	}

}

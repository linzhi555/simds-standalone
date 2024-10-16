package lib

import (
	"encoding/json"
	"log"

	"simds-standalone/cluster/base"
)

func init() {
	base.InverseJsonTable["Signal*"] = func(s string) base.MessageBody {
		var res Signal
		err := json.Unmarshal([]byte(s), &res)
		if err != nil {
			log.Fatalln(s, err)
		}
		return res
	}
}

type Signal string

func (s Signal) MessageBody() {}

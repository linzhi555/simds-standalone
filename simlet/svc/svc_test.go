package svc_test

import (
	"log"
	"simds-standalone/simlet/svc"
	"testing"
)

func TestProto(t *testing.T) {
	m := svc.Message{From: "asdf", To: "asdfd", Content: "asdfds"}
	log.Println(m.String())
}

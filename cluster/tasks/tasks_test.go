package tasks_test

import (
	"simds-standalone/core/tasks"
	"testing"
)

func TestLua(t *testing.T) {
	l := tasks.NewLuaVm()
	err := l.DoString("print(11)")
	if err != nil {
		t.Error(err)
	}

	err = l.DoString("print(double(13))")

	if err != nil {
		t.Error(err)
	}
	err = l.DoString("print(isPrime(12))")

	if err != nil {
		t.Error(err)
	}

}

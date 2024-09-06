package common_test

import "testing"
import "simds-standalone/common"

// TestAdd tests the Add function.
func TestMatternMatch(t *testing.T) {
	testmatch(t, "simds-taskgen0", "simds-taskgen0", true)
	testmatch(t, "*", "simds-taskgen0", true)
	testmatch(t, "simds-taskgen*", "simds-taskgenasd", true)
	testmatch(t, "simds-taskgen*", "simds-taskgen3", true)
	testmatch(t, "*-taskgen*", "simds-taskgen3", true)
	testmatch(t, "sim*-ta*en*", "simds-taskgen3", true)
	testmatch(t, "*", "", true)
	testmatch(t, "asd*", "asd", true)

	testmatch(t, "Asd*", "VecAsd", false)
	testmatch(t, "VecAsd*", "Asd", false)
	testmatch(t, "NodeInfo*", "VecNodeInfoUpate", false)
	testmatch(t, "sim*taskgen", "simds-taskgen3", false)
	testmatch(t, "", "simds-taskgen3", false)
}

func _quote(s string) string {
	return "\"" + s + "\""

}

func testmatch(t *testing.T, p, s string, expect bool) {
	if expect != common.MatchPattern(p, s) {
		t.Error("error:", _quote(p), _quote(s), "expect:", expect)
	} else {
		t.Log("success:", _quote(p), _quote(s), "expect:", expect)
	}
}

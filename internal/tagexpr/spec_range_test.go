package tagexpr_test

import (
	"reflect"
	"testing"

	"github.com/telecom-cloud/client-go/internal/tagexpr"
)

func TestIssue12(t *testing.T) {
	vm := tagexpr.New("te")
	type I int
	type S struct {
		F    []I              `te:"range($, '>'+sprintf('%v:%v', #k, #v+2+len($)))"`
		Fs   [][]I            `te:"range($, range(#v, '>'+sprintf('%v:%v', #k, #v+2+##)))"`
		M    map[string]I     `te:"range($, '>'+sprintf('%s:%v', #k, #v+2+##))"`
		MFs  []map[string][]I `te:"range($, range(#v, range(#v, '>'+sprintf('%v:%v', #k, #v+2+##))))"`
		MFs2 []map[string][]I `te:"range($, range(#v, range(#v, '>'+sprintf('%v:%v', #k, #v+2+##))))"`
	}
	a := []I{2, 3}
	r := vm.MustRun(S{
		F:    a,
		Fs:   [][]I{a},
		M:    map[string]I{"m0": 2, "m1": 3},
		MFs:  []map[string][]I{{"m": a}},
		MFs2: []map[string][]I{},
	})
	assertEqual(t, []interface{}{">0:6", ">1:7"}, r.Eval("F"))
	assertEqual(t, []interface{}{[]interface{}{">0:6", ">1:7"}}, r.Eval("Fs"))
	assertEqual(t, []interface{}{[]interface{}{[]interface{}{">0:6", ">1:7"}}}, r.Eval("MFs"))
	assertEqual(t, []interface{}{}, r.Eval("MFs2"))
	assertEqual(t, true, r.EvalBool("MFs2"))

	// result may not stable for map
	got := r.Eval("M")
	if !reflect.DeepEqual([]interface{}{">m0:6", ">m1:7"}, got) &&
		!reflect.DeepEqual([]interface{}{">m1:7", ">m0:6"}, got) {
		t.Fatal(got)
	}
}

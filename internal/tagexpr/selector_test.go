package tagexpr

import (
	"testing"
)

func TestExprSelector(t *testing.T) {
	es := ExprSelector("F1.Index")
	field, ok := es.ParentField()
	if !ok {
		t.Fatal("not ok")
	}
	if "F1" != field {
		t.Fatal(field)
	}
}

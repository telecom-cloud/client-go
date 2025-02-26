package binding

import (
	"reflect"
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

type foo struct {
	f1 string
}

func TestReflect_TypeID(t *testing.T) {
	_, intType := valueAndTypeID(int(1))
	_, uintType := valueAndTypeID(uint(1))
	_, shouldBeIntType := valueAndTypeID(int(1))
	assert.DeepEqual(t, intType, shouldBeIntType)
	assert.NotEqual(t, intType, uintType)

	foo1 := foo{f1: "1"}
	foo2 := foo{f1: "2"}
	_, foo1Type := valueAndTypeID(foo1)
	_, foo2Type := valueAndTypeID(foo2)
	_, foo2PointerType := valueAndTypeID(&foo2)
	_, foo1PointerType := valueAndTypeID(&foo1)
	assert.DeepEqual(t, foo1Type, foo2Type)
	assert.NotEqual(t, foo1Type, foo2PointerType)
	assert.DeepEqual(t, foo1PointerType, foo2PointerType)
}

func TestReflect_CheckPointer(t *testing.T) {
	foo1 := foo{}
	foo1Val := reflect.ValueOf(foo1)
	err := checkPointer(foo1Val)
	if err == nil {
		t.Errorf("expect an err, but get nil")
	}

	foo2 := &foo{}
	foo2Val := reflect.ValueOf(foo2)
	err = checkPointer(foo2Val)
	if err != nil {
		t.Error(err)
	}

	foo3 := (*foo)(nil)
	foo3Val := reflect.ValueOf(foo3)
	err = checkPointer(foo3Val)
	if err == nil {
		t.Errorf("expect an err, but get nil")
	}
}

func TestReflect_DereferPointer(t *testing.T) {
	var foo1 ****foo
	foo1Val := reflect.ValueOf(foo1)
	rt := dereferPointer(foo1Val)
	if rt.Kind() == reflect.Ptr {
		t.Errorf("expect non-pointer type, but get pointer")
	}
	assert.DeepEqual(t, "foo", rt.Name())

	var foo2 foo
	foo2Val := reflect.ValueOf(foo2)
	rt2 := dereferPointer(foo2Val)
	if rt2.Kind() == reflect.Ptr {
		t.Errorf("expect non-pointer type, but get pointer")
	}
	assert.DeepEqual(t, "foo", rt2.Name())
}

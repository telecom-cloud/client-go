package binding

import (
	"testing"
)

func Test_ValidateStruct(t *testing.T) {
	type User struct {
		Age int `vd:"$>=0&&$<=130"`
	}

	user := &User{
		Age: 135,
	}
	err := DefaultValidator().ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected an error, but got nil")
	}
}

func Test_ValidateTag(t *testing.T) {
	type User struct {
		Age int `query:"age" vt:"$>=0&&$<=130"`
	}

	user := &User{
		Age: 135,
	}
	validateConfig := NewValidateConfig()
	validateConfig.ValidateTag = "vt"
	vd := NewValidator(validateConfig)
	err := vd.ValidateStruct(user)
	if err == nil {
		t.Fatalf("expected an error, but got nil")
	}

	bindConfig := NewBindConfig()
	bindConfig.Validator = vd
	binder := NewDefaultBinder(bindConfig)
	user = &User{}
	req := newMockRequest().
		SetRequestURI("http://foobar.com?age=135").
		SetHeaders("h", "header")
	err = binder.BindAndValidate(req.Req, user, nil)
	if err == nil {
		t.Fatalf("expected an error, but got nil")
	}
}

package binding

import (
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/route/param"
)

type Binder interface {
	Name() string
	Bind(*protocol.Request, interface{}, param.Params) error
	BindAndValidate(*protocol.Request, interface{}, param.Params) error
	BindQuery(*protocol.Request, interface{}) error
	BindHeader(*protocol.Request, interface{}) error
	BindPath(*protocol.Request, interface{}, param.Params) error
	BindForm(*protocol.Request, interface{}) error
	BindJSON(*protocol.Request, interface{}) error
	BindProtobuf(*protocol.Request, interface{}) error
}

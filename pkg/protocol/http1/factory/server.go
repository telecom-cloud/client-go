package factory

import (
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/protocol/http1"
	"github.com/telecom-cloud/client-go/pkg/protocol/suite"
)

var _ suite.ServerFactory = (*serverFactory)(nil)

type serverFactory struct {
	option *http1.Option
}

// New is called by Crafter during engine.Run()
func (s *serverFactory) New(core suite.Core) (server protocol.Server, err error) {
	serv := http1.NewServer()
	serv.Option = *s.option
	serv.Core = core
	return serv, nil
}

func NewServerFactory(option *http1.Option) suite.ServerFactory {
	return &serverFactory{
		option: option,
	}
}

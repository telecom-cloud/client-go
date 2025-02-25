package factory

import (
	"github.com/telecom-cloud/client-go/pkg/protocol/client"
	"github.com/telecom-cloud/client-go/pkg/protocol/http1"
	"github.com/telecom-cloud/client-go/pkg/protocol/suite"
)

var _ suite.ClientFactory = (*clientFactory)(nil)

type clientFactory struct {
	option *http1.ClientOptions
}

func (s *clientFactory) NewHostClient() (client client.HostClient, err error) {
	return http1.NewHostClient(s.option), nil
}

func NewClientFactory(option *http1.ClientOptions) suite.ClientFactory {
	return &clientFactory{
		option: option,
	}
}

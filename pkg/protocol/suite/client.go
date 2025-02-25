package suite

import "github.com/telecom-cloud/client-go/pkg/protocol/client"

type ClientFactory interface {
	NewHostClient() (hc client.HostClient, err error)
}

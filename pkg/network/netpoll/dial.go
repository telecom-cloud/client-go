package netpoll

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/cloudwego/netpoll"
	"github.com/telecom-cloud/client-go/pkg/common/errors"
	"github.com/telecom-cloud/client-go/pkg/network"
)

var errNotSupportTLS = errors.NewPublic("not support tls")

type dialer struct {
	netpoll.Dialer
}

func (d dialer) DialConnection(n, address string, timeout time.Duration, tlsConfig *tls.Config) (conn network.Conn, err error) {
	if tlsConfig != nil {
		// https
		return nil, errNotSupportTLS
	}
	c, err := d.Dialer.DialConnection(n, address, timeout)
	if err != nil {
		return nil, err
	}
	conn = newConn(c)
	return
}

func (d dialer) DialTimeout(network, address string, timeout time.Duration, tlsConfig *tls.Config) (conn net.Conn, err error) {
	if tlsConfig != nil {
		return nil, errNotSupportTLS
	}
	conn, err = d.Dialer.DialTimeout(network, address, timeout)
	return
}

func (d dialer) AddTLS(conn network.Conn, tlsConfig *tls.Config) (network.Conn, error) {
	return nil, errNotSupportTLS
}

func NewDialer() network.Dialer {
	return dialer{Dialer: netpoll.NewDialer()}
}

package dialer

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/telecom-cloud/client-go/pkg/network"
)

var defaultDialer network.Dialer

// SetDialer is used to set the global default dialer.
// Deprecated: use WithDialer instead.
func SetDialer(dialer network.Dialer) {
	defaultDialer = dialer
}

func DefaultDialer() network.Dialer {
	return defaultDialer
}

func DialConnection(network, address string, timeout time.Duration, tlsConfig *tls.Config) (conn network.Conn, err error) {
	return defaultDialer.DialConnection(network, address, timeout, tlsConfig)
}

func DialTimeout(network, address string, timeout time.Duration, tlsConfig *tls.Config) (conn net.Conn, err error) {
	return defaultDialer.DialTimeout(network, address, timeout, tlsConfig)
}

// AddTLS is used to add tls to a persistent connection, i.e. negotiate a TLS session. If conn is already a TLS
// tunnel, this function establishes a nested TLS session inside the encrypted channel.
func AddTLS(conn network.Conn, tlsConfig *tls.Config) (network.Conn, error) {
	return defaultDialer.AddTLS(conn, tlsConfig)
}

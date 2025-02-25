package dialer

import (
	"crypto/tls"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
	"github.com/telecom-cloud/client-go/pkg/network"
)

func TestDialer(t *testing.T) {
	SetDialer(&mockDialer{})
	dialer := DefaultDialer()
	assert.DeepEqual(t, &mockDialer{}, dialer)

	_, err := AddTLS(nil, nil)
	assert.NotNil(t, err)

	_, err = DialConnection("", "", 0, nil)
	assert.NotNil(t, err)

	_, err = DialTimeout("", "", 0, nil)
	assert.NotNil(t, err)
}

type mockDialer struct{}

func (m *mockDialer) DialConnection(network, address string, timeout time.Duration, tlsConfig *tls.Config) (conn network.Conn, err error) {
	return nil, errors.New("method not implement")
}

func (m *mockDialer) DialTimeout(network, address string, timeout time.Duration, tlsConfig *tls.Config) (conn net.Conn, err error) {
	return nil, errors.New("method not implement")
}

func (m *mockDialer) AddTLS(conn network.Conn, tlsConfig *tls.Config) (network.Conn, error) {
	return nil, errors.New("method not implement")
}

//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package client

import (
	"crypto/tls"
	"math/rand"
	"time"

	"github.com/telecom-cloud/client-go/pkg/network"
	"github.com/telecom-cloud/client-go/pkg/network/netpoll"
	"github.com/telecom-cloud/client-go/pkg/network/standard"
)

func newMockDialerWithCustomFunc(network, address string, timeout time.Duration, f func(network, address string, timeout time.Duration, tlsConfig *tls.Config)) network.Dialer {
	dialer := standard.NewDialer()
	if rand.Intn(2) == 0 {
		dialer = netpoll.NewDialer()
	}
	return &mockDialer{
		Dialer:           dialer,
		customDialerFunc: f,
		network:          network,
		address:          address,
		timeout:          timeout,
	}
}

//go:build !windows
// +build !windows

package dialer

import "github.com/telecom-cloud/client-go/pkg/network/netpoll"

func init() {
	defaultDialer = netpoll.NewDialer()
}

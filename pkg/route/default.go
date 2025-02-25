//go:build !windows
// +build !windows

package route

import (
	"github.com/telecom-cloud/client-go/pkg/network/netpoll"
)

func init() {
	defaultTransporter = netpoll.NewTransporter
}

//go:build !windows
// +build !windows

package netpoll

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	"github.com/telecom-cloud/client-go/pkg/common/config"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
	"github.com/telecom-cloud/client-go/pkg/common/test/mock"
)

func TestDial(t *testing.T) {
	t.Run("NetpollDial", func(t *testing.T) {
		const nw = "tcp"
		const addr = "localhost:10100"
		transporter := NewTransporter(&config.Options{
			Addr:    addr,
			Network: nw,
		})
		go transporter.ListenAndServe(func(ctx context.Context, conn interface{}) error {
			return nil
		})
		defer transporter.Close()
		time.Sleep(100 * time.Millisecond)

		dial := NewDialer()
		// DialConnection
		_, err := dial.DialConnection("tcp", "localhost:10101", time.Second, nil) // wrong addr
		assert.NotNil(t, err)

		nwConn, err := dial.DialConnection(nw, addr, time.Second, nil)
		assert.Nil(t, err)
		defer nwConn.Close()
		_, err = nwConn.Write([]byte("abcdef"))
		assert.Nil(t, err)
		// DialTimeout
		nConn, err := dial.DialTimeout(nw, addr, time.Second, nil)
		assert.Nil(t, err)
		defer nConn.Close()
	})

	t.Run("NotSupportTLS", func(t *testing.T) {
		dial := NewDialer()
		_, err := dial.AddTLS(mock.NewConn(""), nil)
		assert.DeepEqual(t, errNotSupportTLS, err)
		_, err = dial.DialConnection("tcp", "localhost:10102", time.Microsecond, &tls.Config{})
		assert.DeepEqual(t, errNotSupportTLS, err)
		_, err = dial.DialTimeout("tcp", "localhost:10102", time.Microsecond, &tls.Config{})
		assert.DeepEqual(t, errNotSupportTLS, err)
	})
}

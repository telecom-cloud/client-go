//go:build !windows
// +build !windows

package netpoll

import (
	"context"
	"net"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/telecom-cloud/client-go/pkg/common/config"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
	"github.com/telecom-cloud/client-go/pkg/network"
	"golang.org/x/sys/unix"
)

func TestTransport(t *testing.T) {
	const nw = "tcp"
	const addr = "localhost:10103"
	t.Run("TestDefault", func(t *testing.T) {
		var onConnFlag, onAcceptFlag, onDataFlag int32
		transporter := NewTransporter(&config.Options{
			Addr:    addr,
			Network: nw,
			OnConnect: func(ctx context.Context, conn network.Conn) context.Context {
				atomic.StoreInt32(&onConnFlag, 1)
				return ctx
			},
			WriteTimeout: time.Second,
			OnAccept: func(conn net.Conn) context.Context {
				atomic.StoreInt32(&onAcceptFlag, 1)
				return context.Background()
			},
		})
		go transporter.ListenAndServe(func(ctx context.Context, conn interface{}) error {
			atomic.StoreInt32(&onDataFlag, 1)
			return nil
		})
		defer transporter.Close()
		time.Sleep(100 * time.Millisecond)

		dial := NewDialer()
		conn, err := dial.DialConnection(nw, addr, time.Second, nil)
		assert.Nil(t, err)
		_, err = conn.Write([]byte("123"))
		assert.Nil(t, err)
		time.Sleep(100 * time.Millisecond)

		assert.Assert(t, atomic.LoadInt32(&onConnFlag) == 1)
		assert.Assert(t, atomic.LoadInt32(&onAcceptFlag) == 1)
		assert.Assert(t, atomic.LoadInt32(&onDataFlag) == 1)
	})

	t.Run("TestSenseClientDisconnection", func(t *testing.T) {
		var onReqFlag int32
		transporter := NewTransporter(&config.Options{
			Addr:                     addr,
			Network:                  nw,
			SenseClientDisconnection: true,
		})

		go transporter.ListenAndServe(func(ctx context.Context, conn interface{}) error {
			atomic.StoreInt32(&onReqFlag, 1)
			time.Sleep(100 * time.Millisecond)
			assert.DeepEqual(t, context.Canceled, ctx.Err())
			return nil
		})
		defer transporter.Close()
		time.Sleep(100 * time.Millisecond)

		dial := NewDialer()
		conn, err := dial.DialConnection(nw, addr, time.Second, nil)
		assert.Nil(t, err)
		_, err = conn.Write([]byte("123"))
		assert.Nil(t, err)
		err = conn.Close()
		assert.Nil(t, err)
		time.Sleep(100 * time.Millisecond)

		assert.Assert(t, atomic.LoadInt32(&onReqFlag) == 1)
	})

	t.Run("TestListenConfig", func(t *testing.T) {
		listenCfg := &net.ListenConfig{Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEADDR, 1)
				syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
			})
		}}
		transporter := NewTransporter(&config.Options{
			Addr:         addr,
			Network:      nw,
			ListenConfig: listenCfg,
		})
		go transporter.ListenAndServe(func(ctx context.Context, conn interface{}) error {
			return nil
		})
		defer transporter.Close()
	})

	t.Run("TestExceptionCase", func(t *testing.T) {
		assert.Panic(t, func() { // listen err
			transporter := NewTransporter(&config.Options{
				Network: "unknow",
			})
			transporter.ListenAndServe(func(ctx context.Context, conn interface{}) error {
				return nil
			})
		})
	})
}

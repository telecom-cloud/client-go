//go:build !windows
// +build !windows

package netpoll

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/cloudwego/netpoll"
	"github.com/telecom-cloud/client-go/pkg/common/config"
	"github.com/telecom-cloud/client-go/pkg/common/logger"
	"github.com/telecom-cloud/client-go/pkg/network"
)

func init() {
	// disable netpoll's log
	netpoll.SetLoggerOutput(io.Discard)
}

type ctxCancelKeyStruct struct{}

var ctxCancelKey = ctxCancelKeyStruct{}

func cancelContext(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	ctx = context.WithValue(ctx, ctxCancelKey, cancel)
	return ctx
}

type transporter struct {
	sync.RWMutex
	senseClientDisconnection bool
	network                  string
	addr                     string
	keepAliveTimeout         time.Duration
	readTimeout              time.Duration
	writeTimeout             time.Duration
	listener                 net.Listener
	eventLoop                netpoll.EventLoop
	listenConfig             *net.ListenConfig
	OnAccept                 func(conn net.Conn) context.Context
	OnConnect                func(ctx context.Context, conn network.Conn) context.Context
}

// For transporter switch
func NewTransporter(options *config.Options) network.Transporter {
	return &transporter{
		senseClientDisconnection: options.SenseClientDisconnection,
		network:                  options.Network,
		addr:                     options.Addr,
		keepAliveTimeout:         options.KeepAliveTimeout,
		readTimeout:              options.ReadTimeout,
		writeTimeout:             options.WriteTimeout,
		listener:                 nil,
		eventLoop:                nil,
		listenConfig:             options.ListenConfig,
		OnAccept:                 options.OnAccept,
		OnConnect:                options.OnConnect,
	}
}

// ListenAndServe binds listen address and keep serving, until an error occurs
// or the transport shutdowns
func (t *transporter) ListenAndServe(onReq network.OnData) (err error) {
	network.UnlinkUdsFile(t.network, t.addr) //nolint:errcheck
	if t.listenConfig != nil {
		t.listener, err = t.listenConfig.Listen(context.Background(), t.network, t.addr)
	} else {
		t.listener, err = net.Listen(t.network, t.addr)
	}

	if err != nil {
		panic("create netpoll listener fail: " + err.Error())
	}

	// Initialize custom option for EventLoop
	opts := []netpoll.Option{
		netpoll.WithIdleTimeout(t.keepAliveTimeout),
		netpoll.WithOnPrepare(func(conn netpoll.Connection) context.Context {
			conn.SetReadTimeout(t.readTimeout) // nolint:errcheck
			if t.writeTimeout > 0 {
				conn.SetWriteTimeout(t.writeTimeout)
			}
			ctx := context.Background()
			if t.OnAccept != nil {
				ctx = t.OnAccept(newConn(conn))
			}
			if t.senseClientDisconnection {
				ctx = cancelContext(ctx)
			}
			return ctx
		}),
	}

	if t.OnConnect != nil {
		opts = append(opts, netpoll.WithOnConnect(func(ctx context.Context, conn netpoll.Connection) context.Context {
			return t.OnConnect(ctx, newConn(conn))
		}))
	}

	if t.senseClientDisconnection {
		opts = append(opts, netpoll.WithOnDisconnect(func(ctx context.Context, connection netpoll.Connection) {
			cancelFunc, ok := ctx.Value(ctxCancelKey).(context.CancelFunc)
			if cancelFunc != nil && ok {
				cancelFunc()
			}
		}))
	}

	// Create EventLoop
	t.Lock()
	t.eventLoop, err = netpoll.NewEventLoop(func(ctx context.Context, connection netpoll.Connection) error {
		return onReq(ctx, newConn(connection))
	}, opts...)
	t.Unlock()
	if err != nil {
		panic("create netpoll event-loop fail")
	}

	// Start Server
	logger.SystemLogger().Infof("HTTP server listening on address=%s", t.listener.Addr().String())
	t.RLock()
	err = t.eventLoop.Serve(t.listener)
	t.RUnlock()
	if err != nil {
		panic("netpoll server exit")
	}

	return nil
}

// Close forces transport to close immediately (no wait timeout)
func (t *transporter) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	return t.Shutdown(ctx)
}

// Shutdown will trigger listener stop and graceful shutdown
// It will wait all connections close until reaching ctx.Deadline()
func (t *transporter) Shutdown(ctx context.Context) error {
	defer func() {
		network.UnlinkUdsFile(t.network, t.addr) //nolint:errcheck
		t.RUnlock()
	}()
	t.RLock()
	if t.eventLoop == nil {
		return nil
	}
	return t.eventLoop.Shutdown(ctx)
}

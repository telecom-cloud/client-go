package standard

import (
	"context"
	"crypto/tls"
	"net"
	"sync"
	"time"

	"github.com/telecom-cloud/client-go/pkg/common/config"
	"github.com/telecom-cloud/client-go/pkg/common/logger"
	"github.com/telecom-cloud/client-go/pkg/network"
)

type transport struct {
	// Per-connection buffer size for requests' reading.
	// This also limits the maximum header size.
	//
	// Increase this buffer if your clients send multi-KB RequestURIs
	// and/or multi-KB headers (for example, BIG cookies).
	//
	// Default buffer size is used if not set.
	readBufferSize   int
	network          string
	addr             string
	keepAliveTimeout time.Duration
	readTimeout      time.Duration
	handler          network.OnData
	ln               net.Listener
	tls              *tls.Config
	listenConfig     *net.ListenConfig
	lock             sync.Mutex
	OnAccept         func(conn net.Conn) context.Context
	OnConnect        func(ctx context.Context, conn network.Conn) context.Context
}

func (t *transport) serve() (err error) {
	network.UnlinkUdsFile(t.network, t.addr) //nolint:errcheck
	t.lock.Lock()
	if t.listenConfig != nil {
		t.ln, err = t.listenConfig.Listen(context.Background(), t.network, t.addr)
	} else {
		t.ln, err = net.Listen(t.network, t.addr)
	}
	t.lock.Unlock()
	if err != nil {
		return err
	}
	logger.SystemLogger().Infof("HTTP server listening on address=%s", t.ln.Addr().String())
	for {
		ctx := context.Background()
		conn, err := t.ln.Accept()
		var c network.Conn
		if err != nil {
			logger.SystemLogger().Errorf("Error=%s", err.Error())
			return err
		}

		if t.OnAccept != nil {
			ctx = t.OnAccept(conn)
		}

		if t.tls != nil {
			c = newTLSConn(tls.Server(conn, t.tls), t.readBufferSize)
		} else {
			c = newConn(conn, t.readBufferSize)
		}

		if t.OnConnect != nil {
			ctx = t.OnConnect(ctx, c)
		}
		go t.handler(ctx, c)
	}
}

func (t *transport) ListenAndServe(onData network.OnData) (err error) {
	t.handler = onData
	return t.serve()
}

func (t *transport) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	return t.Shutdown(ctx)
}

func (t *transport) Shutdown(ctx context.Context) error {
	defer func() {
		network.UnlinkUdsFile(t.network, t.addr) //nolint:errcheck
	}()
	t.lock.Lock()
	if t.ln != nil {
		_ = t.ln.Close()
	}
	t.lock.Unlock()
	<-ctx.Done()
	return nil
}

// For transporter switch
func NewTransporter(options *config.Options) network.Transporter {
	return &transport{
		readBufferSize:   options.ReadBufferSize,
		network:          options.Network,
		addr:             options.Addr,
		keepAliveTimeout: options.KeepAliveTimeout,
		readTimeout:      options.ReadTimeout,
		tls:              options.TLS,
		listenConfig:     options.ListenConfig,
		OnAccept:         options.OnAccept,
		OnConnect:        options.OnConnect,
	}
}

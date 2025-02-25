package netpoll

import (
	"errors"
	"io"
	"strings"
	"syscall"

	"github.com/cloudwego/netpoll"
	errs "github.com/telecom-cloud/client-go/pkg/common/errors"
	"github.com/telecom-cloud/client-go/pkg/common/logger"
	"github.com/telecom-cloud/client-go/pkg/network"
)

type Conn struct {
	network.Conn
}

func (c *Conn) ToCrafterError(err error) error {
	if errors.Is(err, netpoll.ErrConnClosed) || errors.Is(err, syscall.EPIPE) {
		return errs.ErrConnectionClosed
	}

	// only unify read timeout for now
	if errors.Is(err, netpoll.ErrReadTimeout) {
		return errs.ErrTimeout
	}
	return err
}

func (c *Conn) Peek(n int) (b []byte, err error) {
	b, err = c.Conn.Peek(n)
	err = normalizeErr(err)
	return
}

func (c *Conn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	err = normalizeErr(err)
	return n, err
}

func (c *Conn) Skip(n int) error {
	return c.Conn.Skip(n)
}

func (c *Conn) Release() error {
	return c.Conn.Release()
}

func (c *Conn) Len() int {
	return c.Conn.Len()
}

func (c *Conn) ReadByte() (b byte, err error) {
	b, err = c.Conn.ReadByte()
	err = normalizeErr(err)
	return
}

func (c *Conn) ReadBinary(n int) (b []byte, err error) {
	b, err = c.Conn.ReadBinary(n)
	err = normalizeErr(err)
	return
}

func (c *Conn) Malloc(n int) (buf []byte, err error) {
	return c.Conn.Malloc(n)
}

func (c *Conn) WriteBinary(b []byte) (n int, err error) {
	return c.Conn.WriteBinary(b)
}

func (c *Conn) Flush() error {
	return c.Conn.Flush()
}

func (c *Conn) HandleSpecificError(err error, rip string) (needIgnore bool) {
	if errors.Is(err, netpoll.ErrConnClosed) || errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
		// ignore flushing error when connection is closed or reset
		if strings.Contains(err.Error(), "when flush") {
			return true
		}
		logger.SystemLogger().Debugf("Netpoll error=%s, remoteAddr=%s", err.Error(), rip)
		return true
	}
	return false
}

func normalizeErr(err error) error {
	if errors.Is(err, netpoll.ErrEOF) {
		return io.EOF
	}

	return err
}

func newConn(c netpoll.Connection) network.Conn {
	return &Conn{Conn: c.(network.Conn)}
}

package protocol

import (
	"context"

	"github.com/telecom-cloud/client-go/pkg/network"
)

type Server interface {
	Serve(c context.Context, conn network.Conn) error
}

type StreamServer interface {
	Serve(c context.Context, conn network.StreamConn) error
}

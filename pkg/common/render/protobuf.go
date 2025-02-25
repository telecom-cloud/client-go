package render

import (
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"google.golang.org/protobuf/proto"
)

// ProtoBuf contains the given interface object.
type ProtoBuf struct {
	Data interface{}
}

var protobufContentType = "application/x-protobuf"

// Render (ProtoBuf) marshals the given interface object and writes data with custom ContentType.
func (r ProtoBuf) Render(resp *protocol.Response) error {
	r.WriteContentType(resp)
	bytes, err := proto.Marshal(r.Data.(proto.Message))
	if err != nil {
		return err
	}

	resp.AppendBody(bytes)
	return nil
}

// WriteContentType (ProtoBuf) writes ProtoBuf ContentType.
func (r ProtoBuf) WriteContentType(resp *protocol.Response) {
	writeContentType(resp, protobufContentType)
}

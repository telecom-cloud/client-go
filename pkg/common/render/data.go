package render

import "github.com/telecom-cloud/client-go/pkg/protocol"

// Data contains ContentType and bytes data.
type Data struct {
	ContentType string
	Data        []byte
}

// Render (Data) writes data with custom ContentType.
func (r Data) Render(resp *protocol.Response) (err error) {
	r.WriteContentType(resp)
	resp.AppendBody(r.Data)
	return
}

// WriteContentType (Data) writes custom ContentType.
func (r Data) WriteContentType(resp *protocol.Response) {
	writeContentType(resp, r.ContentType)
}

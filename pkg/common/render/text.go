package render

import (
	"fmt"

	"github.com/telecom-cloud/client-go/pkg/protocol"
)

// String contains the given interface object slice and its format.
type String struct {
	Format string
	Data   []interface{}
}

var plainContentType = "text/plain; charset=utf-8"

// Render (String) writes data with custom ContentType.
func (r String) Render(resp *protocol.Response) error {
	writeContentType(resp, plainContentType)
	output := r.Format
	if len(r.Data) > 0 {
		output = fmt.Sprintf(r.Format, r.Data...)
	}
	resp.AppendBodyString(output)
	return nil
}

// WriteContentType (String) writes Plain ContentType.
func (r String) WriteContentType(resp *protocol.Response) {
	writeContentType(resp, plainContentType)
}

package render

import "github.com/telecom-cloud/client-go/pkg/protocol"

// Render interface is to be implemented by JSON, XML, HTML, YAML and so on.
type Render interface {
	// Render writes data with custom ContentType.
	// Do not panic inside, RequestContext will handle it.
	Render(resp *protocol.Response) error
	// WriteContentType writes custom ContentType.
	WriteContentType(resp *protocol.Response)
}

var (
	_ Render = JSONRender{}
	_ Render = String{}
	_ Render = Data{}
)

func writeContentType(resp *protocol.Response, value string) {
	resp.Header.SetContentType(value)
}

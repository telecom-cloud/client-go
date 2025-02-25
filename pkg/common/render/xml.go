package render

import (
	"encoding/xml"

	"github.com/telecom-cloud/client-go/pkg/protocol"
)

// XML contains the given interface object.
type XML struct {
	Data interface{}
}

var xmlContentType = "application/xml; charset=utf-8"

// Render (XML) encodes the given interface object and writes data with custom ContentType.
func (r XML) Render(resp *protocol.Response) error {
	writeContentType(resp, xmlContentType)
	xmlBytes, err := xml.Marshal(r.Data)
	if err != nil {
		return err
	}

	resp.AppendBody(xmlBytes)
	return nil
}

// WriteContentType (XML) writes XML ContentType for response.
func (r XML) WriteContentType(w *protocol.Response) {
	writeContentType(w, xmlContentType)
}

package adaptor

import (
	"bytes"
	"net/http"

	"github.com/telecom-cloud/client-go/pkg/protocol"
)

// GetCompatRequest only support basic function of Request, not for all.
func GetCompatRequest(req *protocol.Request) (*http.Request, error) {
	r, err := http.NewRequest(string(req.Method()), req.URI().String(), bytes.NewReader(req.Body()))
	if err != nil {
		return r, err
	}

	h := make(map[string][]string)
	req.Header.VisitAll(func(k, v []byte) {
		h[string(k)] = append(h[string(k)], string(v))
	})

	r.Header = h
	return r, nil
}

// CopyToCrafterRequest copy uri, host, method, protocol, header, but share body reader from http.Request to protocol.Request.
func CopyToCrafterRequest(req *http.Request, hreq *protocol.Request) error {
	hreq.Header.SetRequestURI(req.RequestURI)
	hreq.Header.SetHost(req.Host)
	hreq.Header.SetMethod(req.Method)
	hreq.Header.SetProtocol(req.Proto)
	for k, v := range req.Header {
		for _, vv := range v {
			hreq.Header.Add(k, vv)
		}
	}
	if req.Body != nil {
		hreq.SetBodyStream(req.Body, hreq.Header.ContentLength())
	}
	return nil
}

//go:build !windows
// +build !windows

package protocol

import "github.com/telecom-cloud/client-go/pkg/common/logger"

func addLeadingSlash(dst, src []byte) []byte {
	// add leading slash for unix paths
	if len(src) == 0 || src[0] != '/' {
		dst = append(dst, '/')
	}

	return dst
}

// checkSchemeWhenCharIsColon check url begin with :
// Scenarios that handle protocols like "http:"
func checkSchemeWhenCharIsColon(i int, rawURL []byte) (scheme, path []byte) {
	if i == 0 {
		logger.Errorf("error happened when try to parse the rawURL(%s): missing protocol scheme", rawURL)
		return
	}
	return rawURL[:i], rawURL[i+1:]
}

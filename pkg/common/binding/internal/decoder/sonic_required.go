//go:build (linux || windows || darwin) && amd64 && !gjson
// +build linux windows darwin
// +build amd64
// +build !gjson

package decoder

import (
	"strings"

	"github.com/bytedance/sonic"
	"github.com/telecom-cloud/client-go/internal/bytesconv"
	"github.com/telecom-cloud/client-go/pkg/common/utils"
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/protocol/consts"
)

func checkRequireJSON(req *protocol.Request, tagInfo TagInfo) bool {
	if !tagInfo.Required {
		return true
	}
	ct := bytesconv.B2s(req.Header.ContentType())
	if !strings.EqualFold(utils.FilterContentType(ct), consts.MIMEApplicationJSON) {
		return false
	}
	node, _ := sonic.Get(req.Body(), stringSliceForInterface(tagInfo.JSONName)...)
	if !node.Exists() {
		idx := strings.LastIndex(tagInfo.JSONName, ".")
		if idx > 0 {
			// There should be a superior if it is empty, it will report 'true' for required
			node, _ := sonic.Get(req.Body(), stringSliceForInterface(tagInfo.JSONName[:idx])...)
			if !node.Exists() {
				return true
			}
		}
		return false
	}
	return true
}

func stringSliceForInterface(s string) (ret []interface{}) {
	x := strings.Split(s, ".")
	for _, val := range x {
		ret = append(ret, val)
	}
	return
}

func keyExist(req *protocol.Request, tagInfo TagInfo) bool {
	ct := bytesconv.B2s(req.Header.ContentType())
	if utils.FilterContentType(ct) != consts.MIMEApplicationJSON {
		return false
	}
	node, _ := sonic.Get(req.Body(), stringSliceForInterface(tagInfo.JSONName)...)
	return node.Exists()
}

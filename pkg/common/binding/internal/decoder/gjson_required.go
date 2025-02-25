//go:build gjson || !(amd64 && (linux || windows || darwin))
// +build gjson !amd64 !linux,!windows,!darwin

package decoder

import (
	"strings"

	"github.com/telecom-cloud/client-go/internal/bytesconv"
	"github.com/telecom-cloud/client-go/pkg/common/utils"
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/protocol/consts"
	"github.com/tidwall/gjson"
)

func checkRequireJSON(req *protocol.Request, tagInfo TagInfo) bool {
	if !tagInfo.Required {
		return true
	}
	ct := bytesconv.B2s(req.Header.ContentType())
	if !strings.EqualFold(utils.FilterContentType(ct), consts.MIMEApplicationJSON) {
		return false
	}
	result := gjson.GetBytes(req.Body(), tagInfo.JSONName)
	if !result.Exists() {
		idx := strings.LastIndex(tagInfo.JSONName, ".")
		// There should be a superior if it is empty, it will report 'true' for required
		if idx > 0 && !gjson.GetBytes(req.Body(), tagInfo.JSONName[:idx]).Exists() {
			return true
		}
		return false
	}
	return true
}

func keyExist(req *protocol.Request, tagInfo TagInfo) bool {
	ct := bytesconv.B2s(req.Header.ContentType())
	if utils.FilterContentType(ct) != consts.MIMEApplicationJSON {
		return false
	}
	result := gjson.GetBytes(req.Body(), tagInfo.JSONName)
	return result.Exists()
}

package utils

import (
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestPathCleanPath(t *testing.T) {
	normalPath := "/Foo/Bar/go/src/github.com/telecom-cloud/client-go/pkg/common/utils/path_test.go"
	expectedNormalPath := "/Foo/Bar/go/src/github.com/telecom-cloud/client-go/pkg/common/utils/path_test.go"
	cleanNormalPath := CleanPath(normalPath)
	assert.DeepEqual(t, expectedNormalPath, cleanNormalPath)

	singleDotPath := "/Foo/Bar/./././go/src"
	expectedSingleDotPath := "/Foo/Bar/go/src"
	cleanSingleDotPath := CleanPath(singleDotPath)
	assert.DeepEqual(t, expectedSingleDotPath, cleanSingleDotPath)

	doubleDotPath := "../../.."
	expectedDoubleDotPath := "/"
	cleanDoublePotPath := CleanPath(doubleDotPath)
	assert.DeepEqual(t, expectedDoubleDotPath, cleanDoublePotPath)

	// MultiDot can be treated as a file name
	multiDotPath := "/../...."
	expectedMultiDotPath := "/...."
	cleanMultiDotPath := CleanPath(multiDotPath)
	assert.DeepEqual(t, expectedMultiDotPath, cleanMultiDotPath)

	nullPath := ""
	expectedNullPath := "/"
	cleanNullPath := CleanPath(nullPath)
	assert.DeepEqual(t, expectedNullPath, cleanNullPath)

	relativePath := "/Foo/Bar/../go/src/../../github.com/telecom-cloud/client-go"
	expectedRelativePath := "/Foo/github.com/telecom-cloud/client-go"
	cleanRelativePath := CleanPath(relativePath)
	assert.DeepEqual(t, expectedRelativePath, cleanRelativePath)

	multiSlashPath := "///////Foo//Bar////go//src/github.com/telecom-cloud/client-go//.."
	expectedMultiSlashPath := "/Foo/Bar/go/src/github.com/cloudwego"
	cleanMultiSlashPath := CleanPath(multiSlashPath)
	assert.DeepEqual(t, expectedMultiSlashPath, cleanMultiSlashPath)

	inputPath := "/Foo/Bar/go/src/github.com/telecom-cloud/client-go/pkg/common/utils/path_test.go/."
	expectedPath := "/Foo/Bar/go/src/github.com/telecom-cloud/client-go/pkg/common/utils/path_test.go/"
	cleanedPath := CleanPath(inputPath)
	assert.DeepEqual(t, expectedPath, cleanedPath)
}

// The Function AddMissingPort can only add the missed port, don't consider the other error case.
func TestPathAddMissingPort(t *testing.T) {
	ipList := []string{"127.0.0.1", "111.111.1.1", "[0:0:0:0:0:ffff:192.1.56.10]", "[0:0:0:0:0:ffff:c0a8:101]", "www.foobar.com"}
	for _, ip := range ipList {
		assert.DeepEqual(t, ip+":443", AddMissingPort(ip, true))
		assert.DeepEqual(t, ip+":80", AddMissingPort(ip, false))
		customizedPort := ":8080"
		assert.DeepEqual(t, ip+customizedPort, AddMissingPort(ip+customizedPort, true))
		assert.DeepEqual(t, ip+customizedPort, AddMissingPort(ip+customizedPort, false))
	}
}

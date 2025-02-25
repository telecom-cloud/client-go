package utils

import (
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

// test assert func
func TestUtilsAssert(t *testing.T) {
	assertPanic := func() (panicked bool) {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		Assert(false, "should panic")
		return false
	}

	// Checking if the assertPanic function results in a panic as expected.
	// We expect a true value because it should panic.
	assert.DeepEqual(t, true, assertPanic())

	// Checking if a true assertion does not result in a panic.
	// We create a wrapper around Assert to capture if it panics when it should not.
	noPanic := func() (panicked bool) {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
			}
		}()
		Assert(true, "should not panic")
		return false
	}

	// We expect a false value because it should not panic.
	assert.DeepEqual(t, false, noPanic())
}

func TestUtilsIsTrueString(t *testing.T) {
	normalTrueStr := "true"
	upperTrueStr := "trUe"
	otherStr := "hertz"

	assert.DeepEqual(t, true, IsTrueString(normalTrueStr))
	assert.DeepEqual(t, true, IsTrueString(upperTrueStr))
	assert.DeepEqual(t, false, IsTrueString(otherStr))
}

// used for TestUtilsNameOfFunction
func testName(a int) {
}

// return the relative path for the function
func TestUtilsNameOfFunction(t *testing.T) {
	pathOfTestName := "github.com/telecom-cloud/client-go/pkg/common/utils.testName"
	pathOfIsTrueString := "github.com/telecom-cloud/client-go/pkg/common/utils.IsTrueString"
	nameOfTestName := NameOfFunction(testName)
	nameOfIsTrueString := NameOfFunction(IsTrueString)

	assert.DeepEqual(t, pathOfTestName, nameOfTestName)
	assert.DeepEqual(t, pathOfIsTrueString, nameOfIsTrueString)
}

func TestUtilsCaseInsensitiveCompare(t *testing.T) {
	lowerStr := []byte("content-length")
	upperStr := []byte("Content-Length")
	assert.DeepEqual(t, true, CaseInsensitiveCompare(lowerStr, upperStr))

	lessStr := []byte("content-type")
	moreStr := []byte("content-length")
	assert.DeepEqual(t, false, CaseInsensitiveCompare(lessStr, moreStr))

	firstStr := []byte("content-type")
	secondStr := []byte("contant-type")
	assert.DeepEqual(t, false, CaseInsensitiveCompare(firstStr, secondStr))
}

// NormalizeHeaderKey can upper the first letter and lower the other letter in
// HTTP header, invervaled by '-'.
// Example: "content-type" -> "Content-Type"
func TestUtilsNormalizeHeaderKey(t *testing.T) {
	contentTypeStr := []byte("Content-Type")
	lowerContentTypeStr := []byte("content-type")
	mixedContentTypeStr := []byte("conTENt-tYpE")
	mixedContertTypeStrWithoutNormalizing := []byte("Content-type")
	NormalizeHeaderKey(contentTypeStr, false)
	NormalizeHeaderKey(lowerContentTypeStr, false)
	NormalizeHeaderKey(mixedContentTypeStr, false)
	NormalizeHeaderKey(lowerContentTypeStr, true)

	assert.DeepEqual(t, "Content-Type", string(contentTypeStr))
	assert.DeepEqual(t, "Content-Type", string(lowerContentTypeStr))
	assert.DeepEqual(t, "Content-Type", string(mixedContentTypeStr))
	assert.DeepEqual(t, "Content-type", string(mixedContertTypeStrWithoutNormalizing))
}

// Cutting up the header Type.
// Example: "Content-Type: application/x-www-form-urlencoded\r\nDate: Fri, 6 Aug 2021 11:00:31 GMT"
// ->"Content-Type: application/x-www-form-urlencoded" and "Date: Fri, 6 Aug 2021 11:00:31 GMT"
func TestUtilsNextLine(t *testing.T) {
	multiHeaderStr := []byte("Content-Type: application/x-www-form-urlencoded\r\nDate: Fri, 6 Aug 2021 11:00:31 GMT")
	contentTypeStr, dateStr, hErr := NextLine(multiHeaderStr)
	assert.DeepEqual(t, nil, hErr)
	assert.DeepEqual(t, "Content-Type: application/x-www-form-urlencoded", string(contentTypeStr))
	assert.DeepEqual(t, "Date: Fri, 6 Aug 2021 11:00:31 GMT", string(dateStr))

	multiHeaderStrWithoutReturn := []byte("Content-Type: application/x-www-form-urlencoded\nDate: Fri, 6 Aug 2021 11:00:31 GMT")
	contentTypeStr, dateStr, hErr = NextLine(multiHeaderStrWithoutReturn)
	assert.DeepEqual(t, nil, hErr)
	assert.DeepEqual(t, "Content-Type: application/x-www-form-urlencoded", string(contentTypeStr))
	assert.DeepEqual(t, "Date: Fri, 6 Aug 2021 11:00:31 GMT", string(dateStr))

	singleHeaderStrWithFirstNewLine := []byte("\nContent-Type: application/x-www-form-urlencoded")
	firstStr, secondStr, sErr := NextLine(singleHeaderStrWithFirstNewLine)
	assert.DeepEqual(t, nil, sErr)
	assert.DeepEqual(t, string(""), string(firstStr))
	assert.DeepEqual(t, "Content-Type: application/x-www-form-urlencoded", string(secondStr))

	singleHeaderStr := []byte("Content-Type: application/x-www-form-urlencoded")
	_, _, sErr = NextLine(singleHeaderStr)
	assert.DeepEqual(t, errNeedMore, sErr)
}

func TestFilterContentType(t *testing.T) {
	contentType := "text/plain; charset=utf-8"
	contentType = FilterContentType(contentType)
	assert.DeepEqual(t, "text/plain", contentType)
}

func TestNormalizeHeaderKeyEdgeCases(t *testing.T) {
	empty := []byte("")
	NormalizeHeaderKey(empty, false)
	assert.DeepEqual(t, []byte(""), empty)
	NormalizeHeaderKey(empty, true)
	assert.DeepEqual(t, []byte(""), empty)
}

func TestFilterContentTypeEdgeCases(t *testing.T) {
	simpleContentType := "text/plain"
	assert.DeepEqual(t, "text/plain", FilterContentType(simpleContentType))
	complexContentType := "text/html; charset=utf-8; format=flowed"
	assert.DeepEqual(t, "text/html", FilterContentType(complexContentType))
}

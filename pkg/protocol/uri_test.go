package protocol

import (
	"bytes"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestURI_Username(t *testing.T) {
	var req Request
	req.SetRequestURI("http://user:pass@example.com/foo/bar")
	u := req.URI()
	user1 := string(u.Username())
	req.Header.SetRequestURIBytes([]byte("/foo/bar"))
	u = req.URI()
	user2 := string(u.Username())
	assert.DeepEqual(t, user1, user2)

	expectUser3 := "user3"
	expectUser4 := "user4"

	u.SetUsername(expectUser3)
	user3 := string(u.Username())
	assert.DeepEqual(t, expectUser3, user3)
	u.SetUsername(expectUser4)
	user4 := string(u.Username())
	assert.DeepEqual(t, expectUser4, user4)

	u.SetUsernameBytes([]byte(user3))
	assert.DeepEqual(t, expectUser3, user3)
	u.SetUsernameBytes([]byte(user4))
	assert.DeepEqual(t, expectUser4, user4)
}

func TestURI_Password(t *testing.T) {
	u := AcquireURI()
	defer ReleaseURI(u)

	expectPassword1 := "password1"
	expectPassword2 := "password2"

	u.SetPassword(expectPassword1)
	password1 := string(u.Password())
	assert.DeepEqual(t, expectPassword1, password1)
	u.SetPassword(expectPassword2)
	password2 := string(u.Password())
	assert.DeepEqual(t, expectPassword2, password2)

	u.SetPasswordBytes([]byte(password1))
	assert.DeepEqual(t, expectPassword1, password1)
	u.SetPasswordBytes([]byte(password2))
	assert.DeepEqual(t, expectPassword2, password2)
}

func TestURI_Hash(t *testing.T) {
	u := AcquireURI()
	defer ReleaseURI(u)

	expectHash1 := "hash1"
	expectHash2 := "hash2"

	u.SetHash(expectHash1)
	hash1 := string(u.Hash())
	assert.DeepEqual(t, expectHash1, hash1)
	u.SetHash(expectHash2)
	hash2 := string(u.Hash())
	assert.DeepEqual(t, expectHash2, hash2)
}

func TestURI_QueryString(t *testing.T) {
	u := AcquireURI()
	defer ReleaseURI(u)

	expectQueryString1 := "key1=value1&key2=value2"
	expectQueryString2 := "key3=value3&key4=value4"

	u.SetQueryString(expectQueryString1)
	queryString1 := string(u.QueryString())
	assert.DeepEqual(t, expectQueryString1, queryString1)
	u.SetQueryString(expectQueryString2)
	queryString2 := string(u.QueryString())
	assert.DeepEqual(t, expectQueryString2, queryString2)
}

func TestURI_Path(t *testing.T) {
	u := AcquireURI()
	defer ReleaseURI(u)

	expectPath1 := "/"
	expectPath2 := "/path1"
	expectPath3 := "/path3"

	// When Path is not set, Path defaults to "/"
	path1 := string(u.Path())
	assert.DeepEqual(t, expectPath1, path1)

	u.SetPath(expectPath2)
	path2 := string(u.Path())
	assert.DeepEqual(t, expectPath2, path2)
	u.SetPath(expectPath3)
	path3 := string(u.Path())
	assert.DeepEqual(t, expectPath3, path3)

	u.SetPathBytes([]byte(path2))
	assert.DeepEqual(t, expectPath2, path2)
	u.SetPathBytes([]byte(path3))
	assert.DeepEqual(t, expectPath3, path3)
}

func TestURI_Scheme(t *testing.T) {
	u := AcquireURI()
	defer ReleaseURI(u)

	expectScheme1 := "scheme1"
	expectScheme2 := "scheme2"

	u.SetScheme(expectScheme1)
	scheme1 := string(u.Scheme())
	assert.DeepEqual(t, expectScheme1, scheme1)
	u.SetScheme(expectScheme2)
	scheme2 := string(u.Scheme())
	assert.DeepEqual(t, expectScheme2, scheme2)

	u.SetSchemeBytes([]byte(scheme1))
	assert.DeepEqual(t, expectScheme1, scheme1)
	u.SetSchemeBytes([]byte(scheme2))
	assert.DeepEqual(t, expectScheme2, scheme2)
}

func TestURI_Host(t *testing.T) {
	u := AcquireURI()
	defer ReleaseURI(u)

	expectHost1 := "host1"
	expectHost2 := "host2"

	u.SetHost(expectHost1)
	host1 := string(u.Host())
	assert.DeepEqual(t, expectHost1, host1)
	u.SetHost(expectHost2)
	host2 := string(u.Host())
	assert.DeepEqual(t, expectHost2, host2)

	u.SetHostBytes([]byte(host1))
	assert.DeepEqual(t, expectHost1, host1)
	u.SetHostBytes([]byte(host2))
	assert.DeepEqual(t, expectHost2, host2)
}

func TestURI_PathOriginal(t *testing.T) {
	var u URI
	expectPath := "/path"
	u.Parse(nil, []byte(expectPath))
	uri := string(u.PathOriginal())
	assert.DeepEqual(t, expectPath, uri)
}

func TestArgsKV_Get(t *testing.T) {
	var argsKV argsKV
	expectKey := "key"
	expectValue := "value"
	argsKV.key = []byte(expectKey)
	argsKV.value = []byte(expectValue)
	key := string(argsKV.GetKey())
	value := string(argsKV.GetValue())
	assert.DeepEqual(t, expectKey, key)
	assert.DeepEqual(t, expectValue, value)
}

func TestURICopyToQueryArgs(t *testing.T) {
	t.Parallel()

	var u URI
	a := u.QueryArgs()
	a.Set("foo", "bar")

	var u1 URI
	u.CopyTo(&u1)
	a1 := u1.QueryArgs()

	if string(a1.Peek("foo")) != "bar" {
		t.Fatalf("unexpected query args value %q. Expecting %q", a1.Peek("foo"), "bar")
	}
	assert.DeepEqual(t, "bar", string(a1.Peek("foo")))
}

func TestURICopyTo(t *testing.T) {
	t.Parallel()

	var u URI
	var copyU URI
	u.CopyTo(&copyU)
	if !reflect.DeepEqual(&u, &copyU) { //nolint:govet
		t.Fatalf("URICopyTo fail, u: \n%+v\ncopyu: \n%+v\n", &u, &copyU) //nolint:govet
	}

	u.UpdateBytes([]byte("https://google.com/foo?bar=baz&baraz#qqqq"))
	u.CopyTo(&copyU)
	if !reflect.DeepEqual(&u, &copyU) { //nolint:govet
		t.Fatalf("URICopyTo fail, u: \n%+v\ncopyu: \n%+v\n", &u, &copyU) //nolint:govet
	}
}

func TestURILastPathSegment(t *testing.T) {
	t.Parallel()

	testURILastPathSegment(t, "", "")
	testURILastPathSegment(t, "/", "")
	testURILastPathSegment(t, "/foo/bar/", "")
	testURILastPathSegment(t, "/foobar.js", "foobar.js")
	testURILastPathSegment(t, "/foo/bar/baz.html", "baz.html")
}

func testURILastPathSegment(t *testing.T, path, expectedSegment string) {
	var u URI
	u.SetPath(path)
	segment := u.LastPathSegment()
	assert.DeepEqual(t, expectedSegment, string(segment))
}

func TestURIPathEscape(t *testing.T) {
	t.Parallel()

	testURIPathEscape(t, "/foo/bar", "/foo/bar")
	testURIPathEscape(t, "/f_o-o=b:ar,b.c&q", "/f_o-o=b:ar,b.c&q")
	testURIPathEscape(t, "/aa?bb.тест~qq", "/aa%3Fbb.%D1%82%D0%B5%D1%81%D1%82~qq")
}

func TestURIUpdate(t *testing.T) {
	t.Parallel()

	// full uri
	testURIUpdate(t, "http://foo.bar/baz?aaa=22#aaa", "https://aaa.com/bb", "https://aaa.com/bb")
	// empty uri
	testURIUpdate(t, "http://aaa.com/aaa.html?234=234#add", "", "http://aaa.com/aaa.html?234=234#add")

	// request uri
	testURIUpdate(t, "ftp://aaa/xxx/yyy?aaa=bb#aa", "/boo/bar?xx", "ftp://aaa/boo/bar?xx")

	// relative uri
	testURIUpdate(t, "http://foo.bar/baz/xxx.html?aaa=22#aaa", "bb.html?xx=12#pp", "http://foo.bar/baz/bb.html?xx=12#pp")
	testURIUpdate(t, "http://xx/a/b/c/d", "../qwe/p?zx=34", "http://xx/a/b/qwe/p?zx=34")
	testURIUpdate(t, "https://qqq/aaa.html?foo=bar", "?baz=434&aaa#xcv", "https://qqq/aaa.html?baz=434&aaa#xcv")
	testURIUpdate(t, "http://foo.bar/baz", "~a/%20b=c,тест?йцу=ке", "http://foo.bar/~a/%20b=c,%D1%82%D0%B5%D1%81%D1%82?йцу=ке")
	testURIUpdate(t, "http://foo.bar/baz", "/qwe#fragment", "http://foo.bar/qwe#fragment")
	testURIUpdate(t, "http://foobar/baz/xxx", "aaa.html#bb?cc=dd&ee=dfd", "http://foobar/baz/aaa.html#bb?cc=dd&ee=dfd")

	// hash
	testURIUpdate(t, "http://foo.bar/baz#aaa", "#fragment", "http://foo.bar/baz#fragment")

	// uri without scheme
	testURIUpdate(t, "https://foo.bar/baz", "//aaa.bbb/cc?dd", "https://aaa.bbb/cc?dd")
	testURIUpdate(t, "http://foo.bar/baz", "//aaa.bbb/cc?dd", "http://aaa.bbb/cc?dd")
}

func testURIUpdate(t *testing.T, base, update, result string) {
	var u URI
	u.Parse(nil, []byte(base))
	u.Update(update)
	s := u.String()
	assert.DeepEqual(t, result, s)
}

func testURIPathEscape(t *testing.T, path, expectedRequestURI string) {
	var u URI
	u.SetPath(path)
	requestURI := u.RequestURI()
	assert.DeepEqual(t, expectedRequestURI, string(requestURI))
}

func TestDelArgs(t *testing.T) {
	var args Args
	args.Set("foo", "bar")
	assert.DeepEqual(t, string(args.Peek("foo")), "bar")
	args.Del("foo")
	assert.DeepEqual(t, string(args.Peek("foo")), "")

	args.Set("foo2", "bar2")
	assert.DeepEqual(t, string(args.Peek("foo2")), "bar2")
	args.DelBytes([]byte("foo2"))
	assert.DeepEqual(t, string(args.Peek("foo2")), "")
}

func TestURIFullURI(t *testing.T) {
	t.Parallel()

	var args Args

	// empty scheme, path and hash
	testURIFullURI(t, "", "foobar.com", "", "", &args, "http://foobar.com/")

	// empty scheme and hash
	testURIFullURI(t, "", "aaa.com", "/foo/bar", "", &args, "http://aaa.com/foo/bar")
	// empty hash
	testURIFullURI(t, "fTP", "XXx.com", "/foo", "", &args, "ftp://xxx.com/foo")

	// empty args
	testURIFullURI(t, "https", "xx.com", "/", "aaa", &args, "https://xx.com/#aaa")

	// non-empty args and non-ASCII path
	args.Set("foo", "bar")
	args.Set("xxx", "йух")
	testURIFullURI(t, "", "xxx.com", "/тест123", "2er", &args, "http://xxx.com/%D1%82%D0%B5%D1%81%D1%82123?foo=bar&xxx=%D0%B9%D1%83%D1%85#2er")

	// test with empty args and non-empty query string
	var u URI
	u.Parse([]byte("google.com"), []byte("/foo?bar=baz&baraz#qqqq"))
	uri := u.FullURI()
	expectedURI := "http://google.com/foo?bar=baz&baraz#qqqq"
	assert.DeepEqual(t, expectedURI, string(uri))
}

func testURIFullURI(t *testing.T, scheme, host, path, hash string, args *Args, expectedURI string) {
	var u URI

	u.SetScheme(scheme)
	u.SetHost(host)
	u.SetPath(path)
	u.SetHash(hash)
	args.CopyTo(u.QueryArgs())

	uri := u.FullURI()
	assert.DeepEqual(t, expectedURI, string(uri))
}

func TestParsePathWindows(t *testing.T) {
	t.Parallel()

	testParsePathWindows(t, "/../../../../../foo", "/foo")
	testParsePathWindows(t, "/..\\..\\..\\..\\..\\foo", "/foo")
	testParsePathWindows(t, "/..%5c..%5cfoo", "/foo")
}

func TestURIPathNormalize(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.SkipNow()
	}

	t.Parallel()

	var u URI

	// double slash
	testURIPathNormalize(t, &u, "/aa//bb", "/aa/bb")

	// triple slash
	testURIPathNormalize(t, &u, "/x///y/", "/x/y/")

	// multi slashes
	testURIPathNormalize(t, &u, "/abc//de///fg////", "/abc/de/fg/")

	// encoded slashes
	testURIPathNormalize(t, &u, "/xxxx%2fyyy%2f%2F%2F", "/xxxx/yyy/")

	// dotdot
	testURIPathNormalize(t, &u, "/aaa/..", "/")

	// dotdot with trailing slash
	testURIPathNormalize(t, &u, "/xxx/yyy/../", "/xxx/")

	// multi dotdots
	testURIPathNormalize(t, &u, "/aaa/bbb/ccc/../../ddd", "/aaa/ddd")

	// dotdots separated by other data
	testURIPathNormalize(t, &u, "/a/b/../c/d/../e/..", "/a/c/")

	// too many dotdots
	testURIPathNormalize(t, &u, "/aaa/../../../../xxx", "/xxx")
	testURIPathNormalize(t, &u, "/../../../../../..", "/")
	testURIPathNormalize(t, &u, "/../../../../../../", "/")

	// encoded dotdots
	testURIPathNormalize(t, &u, "/aaa%2Fbbb%2F%2E.%2Fxxx", "/aaa/xxx")

	// double slash with dotdots
	testURIPathNormalize(t, &u, "/aaa////..//b", "/b")

	// fake dotdot
	testURIPathNormalize(t, &u, "/aaa/..bbb/ccc/..", "/aaa/..bbb/")

	// single dot
	testURIPathNormalize(t, &u, "/a/./b/././c/./d.html", "/a/b/c/d.html")
	testURIPathNormalize(t, &u, "./foo/", "/foo/")
	testURIPathNormalize(t, &u, "./../.././../../aaa/bbb/../../../././../", "/")
	testURIPathNormalize(t, &u, "./a/./.././../b/./foo.html", "/b/foo.html")
}

func testURIPathNormalize(t *testing.T, u *URI, requestURI, expectedPath string) {
	u.Parse(nil, []byte(requestURI)) //nolint:errcheck
	if string(u.Path()) != expectedPath {
		t.Fatalf("Unexpected path %q. Expected %q. requestURI=%q", u.Path(), expectedPath, requestURI)
	}
}

func testParsePathWindows(t *testing.T, path, expectedPath string) {
	var u URI
	u.Parse(nil, []byte(path))
	parsedPath := u.Path()
	if filepath.Separator == '\\' && string(parsedPath) != expectedPath {
		t.Fatalf("Unexpected Path: %q. Expected %q", parsedPath, expectedPath)
	}
}

func TestParseHostWithStr(t *testing.T) {
	expectUsername := "username"
	expectPassword := "password"

	testParseHostWithStr(t, "username", "", "")
	testParseHostWithStr(t, "username@", expectUsername, "")
	testParseHostWithStr(t, "username:password@", expectUsername, expectPassword)
	testParseHostWithStr(t, ":password@", "", expectPassword)
	testParseHostWithStr(t, ":password", "", "")
}

func testParseHostWithStr(t *testing.T, host, expectUsername, expectPassword string) {
	var u URI
	u.Parse([]byte(host), nil)
	assert.DeepEqual(t, expectUsername, string(u.Username()))
	assert.DeepEqual(t, expectPassword, string(u.Password()))
}

func TestParseURI(t *testing.T) {
	expectURI := "http://google.com/foo?bar=baz&baraz#qqqq"
	uri := string(ParseURI(expectURI).FullURI())
	assert.DeepEqual(t, expectURI, uri)
}

func TestSplitHostURI(t *testing.T) {
	cases := []struct {
		host, uri                      []byte
		wantScheme, wantHost, wantPath []byte
	}{
		{
			[]byte("example.com"), []byte("/foobar"),
			[]byte("http"), []byte("example.com"), []byte("/foobar"),
		},
		{
			[]byte("example2.com"), []byte("http://example2.com"),
			[]byte("http"), []byte("example2.com"), []byte("/"),
		},
		{
			[]byte("example2.com"), []byte("http://example3.com"),
			[]byte("http"), []byte("example3.com"), []byte("/"),
		},
		{
			[]byte("example3.com"), []byte("https://foobar.com?a=b"),
			[]byte("https"), []byte("foobar.com"), []byte("?a=b"),
		},
	}

	for _, c := range cases {
		gotScheme, gotHost, gotPath := splitHostURI(c.host, c.uri)
		if !bytes.Equal(gotScheme, c.wantScheme) || !bytes.Equal(gotHost, c.wantHost) || !bytes.Equal(gotPath, c.wantPath) {
			t.Errorf("splitHostURI(%q, %q) == (%q, %q, %q), want (%q, %q, %q)",
				c.host, c.uri, gotScheme, gotHost, gotPath, c.wantScheme, c.wantHost, c.wantPath)
		}
	}
}

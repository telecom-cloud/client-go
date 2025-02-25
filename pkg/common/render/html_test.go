package render

import (
	"html/template"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
	"github.com/telecom-cloud/client-go/pkg/common/utils"
	"github.com/telecom-cloud/client-go/pkg/protocol"
)

func TestHTMLDebug_StartChecker_timer(t *testing.T) {
	render := &HTMLDebug{
		RefreshInterval: time.Second,
		Delims:          Delims{Left: "{[{", Right: "}]}"},
		FuncMap:         template.FuncMap{},
		Files:           []string{"../../../common/testdata/template/index.tmpl"},
	}
	select {
	case <-render.reloadCh:
		t.Fatalf("should not be triggered")
	default:
	}
	render.startChecker()
	select {
	case <-time.After(render.RefreshInterval + 500*time.Millisecond):
		t.Fatalf("should be triggered in 1.5 second")
	case <-render.reloadCh:
		render.reload()
	}
}

func TestHTMLDebug_StartChecker_fs_watcher(t *testing.T) {
	f, _ := ioutil.TempFile("./", "test.tmpl")
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	render := &HTMLDebug{Files: []string{f.Name()}}
	select {
	case <-render.reloadCh:
		t.Fatalf("should not be triggered")
	default:
	}
	render.startChecker()
	f.Write([]byte("hello"))
	f.Sync()
	select {
	case <-time.After(50 * time.Millisecond):
		t.Fatalf("should be triggered immediately")
	case <-render.reloadCh:
	}
	select {
	case <-render.reloadCh:
		t.Fatalf("should not be triggered")
	default:
	}
}

func TestRenderHTML(t *testing.T) {
	resp := &protocol.Response{}

	tmpl := template.Must(template.New("").
		Delims("{[{", "}]}").
		Funcs(template.FuncMap{}).
		ParseFiles("../../../common/testdata/template/index.tmpl"))

	r := &HTMLProduction{Template: tmpl}

	html := r.Instance("index.tmpl", utils.H{
		"title": "Main website",
	})

	err := r.Close()
	assert.Nil(t, err)

	html.WriteContentType(resp)
	assert.DeepEqual(t, []byte("text/html; charset=utf-8"), resp.Header.Peek("Content-Type"))

	err = html.Render(resp)

	assert.Nil(t, err)
	assert.DeepEqual(t, []byte("text/html; charset=utf-8"), resp.Header.Peek("Content-Type"))
	assert.DeepEqual(t, []byte("<html><h1>Main website</h1></html>"), resp.Body())

	respDebug := &protocol.Response{}

	rDebug := &HTMLDebug{
		Template: tmpl,
		Delims:   Delims{Left: "{[{", Right: "}]}"},
		FuncMap:  template.FuncMap{},
		Files:    []string{"../../../common/testdata/template/index.tmpl"},
	}

	htmlDebug := rDebug.Instance("index.tmpl", utils.H{
		"title": "Main website",
	})

	err = rDebug.Close()
	assert.Nil(t, err)

	htmlDebug.WriteContentType(respDebug)
	assert.DeepEqual(t, []byte("text/html; charset=utf-8"), respDebug.Header.Peek("Content-Type"))

	err = htmlDebug.Render(respDebug)

	assert.Nil(t, err)
	assert.DeepEqual(t, []byte("text/html; charset=utf-8"), respDebug.Header.Peek("Content-Type"))
	assert.DeepEqual(t, []byte("<html><h1>Main website</h1></html>"), respDebug.Body())
}

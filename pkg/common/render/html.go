package render

import (
	"html/template"
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/telecom-cloud/client-go/pkg/common/logger"
	"github.com/telecom-cloud/client-go/pkg/protocol"
)

// Delims represents a set of Left and Right delimiters for HTML template rendering.
type Delims struct {
	// Left delimiter, defaults to {{.
	Left string
	// Right delimiter, defaults to }}.
	Right string
}

// HTMLRender interface is to be implemented by HTMLProduction and HTMLDebug.
type HTMLRender interface {
	// Instance returns an HTML instance.
	Instance(string, interface{}) Render
	Close() error
}

// HTMLProduction contains template reference and its delims.
type HTMLProduction struct {
	Template *template.Template
}

// HTML contains template reference and its name with given interface object.
type HTML struct {
	Template *template.Template
	Name     string
	Data     interface{}
}

var htmlContentType = "text/html; charset=utf-8"

// Instance (HTMLProduction) returns an HTML instance which it realizes Render interface.
func (r HTMLProduction) Instance(name string, data interface{}) Render {
	return HTML{
		Template: r.Template,
		Name:     name,
		Data:     data,
	}
}

func (r HTMLProduction) Close() error {
	return nil
}

// Render (HTML) executes template and writes its result with custom ContentType for response.
func (r HTML) Render(resp *protocol.Response) error {
	r.WriteContentType(resp)

	if r.Name == "" {
		return r.Template.Execute(resp.BodyWriter(), r.Data)
	}
	return r.Template.ExecuteTemplate(resp.BodyWriter(), r.Name, r.Data)
}

// WriteContentType (HTML) writes HTML ContentType.
func (r HTML) WriteContentType(resp *protocol.Response) {
	writeContentType(resp, htmlContentType)
}

type HTMLDebug struct {
	sync.Once
	Template        *template.Template
	RefreshInterval time.Duration

	Files   []string
	FuncMap template.FuncMap
	Delims  Delims

	reloadCh chan struct{}
	watcher  *fsnotify.Watcher
}

func (h *HTMLDebug) Instance(name string, data interface{}) Render {
	h.Do(func() {
		h.startChecker()
	})

	select {
	case <-h.reloadCh:
		h.reload()
	default:
	}

	return HTML{
		Template: h.Template,
		Name:     name,
		Data:     data,
	}
}

func (h *HTMLDebug) Close() error {
	if h.watcher == nil {
		return nil
	}
	return h.watcher.Close()
}

func (h *HTMLDebug) reload() {
	h.Template = template.Must(template.New("").
		Delims(h.Delims.Left, h.Delims.Right).
		Funcs(h.FuncMap).
		ParseFiles(h.Files...))
}

func (h *HTMLDebug) startChecker() {
	h.reloadCh = make(chan struct{})

	if h.RefreshInterval > 0 {
		go func() {
			logger.SystemLogger().Debugf("[HTMLDebug] HTML template reloader started with interval %v", h.RefreshInterval)
			for range time.Tick(h.RefreshInterval) {
				logger.SystemLogger().Debugf("[HTMLDebug] triggering HTML template reloader")
				h.reloadCh <- struct{}{}
				logger.SystemLogger().Debugf("[HTMLDebug] HTML template has been reloaded, next reload in %v", h.RefreshInterval)
			}
		}()
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	h.watcher = watcher
	for _, f := range h.Files {
		err := watcher.Add(f)
		logger.SystemLogger().Debugf("[HTMLDebug] watching file: %s", f)
		if err != nil {
			logger.SystemLogger().Errorf("[HTMLDebug] add watching file: %s, error happened: %v", f, err)
		}

	}

	go func() {
		logger.SystemLogger().Debugf("[HTMLDebug] HTML template reloader started with file watcher")
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					logger.SystemLogger().Debugf("[HTMLDebug] modified file: %s, html render template will be reloaded at the next rendering", event.Name)
					h.reloadCh <- struct{}{}
					logger.SystemLogger().Debugf("[HTMLDebug] HTML template has been reloaded")
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.SystemLogger().Errorf("error happened when watching the rendering files: %v", err)
			}
		}
	}()
}

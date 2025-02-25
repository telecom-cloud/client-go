package route

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/telecom-cloud/client-go/pkg/app"
	"github.com/telecom-cloud/client-go/pkg/common/config"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
)

func TestRouterGroupBasic(t *testing.T) {
	cfg := config.NewOptions(nil)
	router := NewEngine(cfg)
	group := router.Group("/hola", func(c context.Context, ctx *app.RequestContext) {})
	group.Use(func(c context.Context, ctx *app.RequestContext) {})

	assert.DeepEqual(t, len(group.Handlers), 2)
	assert.DeepEqual(t, "/hola", group.BasePath())
	assert.DeepEqual(t, router, group.engine)

	group2 := group.Group("manu")
	group2.Use(func(c context.Context, ctx *app.RequestContext) {}, func(c context.Context, ctx *app.RequestContext) {})

	assert.DeepEqual(t, len(group2.Handlers), 4)
	assert.DeepEqual(t, "/hola/manu", group2.BasePath())
	assert.DeepEqual(t, router, group2.engine)
}

func TestRouterGroupBasicHandle(t *testing.T) {
	performRequestInGroup(t, http.MethodGet)
	performRequestInGroup(t, http.MethodPost)
	performRequestInGroup(t, http.MethodPut)
	performRequestInGroup(t, http.MethodPatch)
	performRequestInGroup(t, http.MethodDelete)
	performRequestInGroup(t, http.MethodHead)
	performRequestInGroup(t, http.MethodOptions)
}

func performRequestInGroup(t *testing.T, method string) {
	router := NewEngine(config.NewOptions(nil))
	v1 := router.Group("v1", func(c context.Context, ctx *app.RequestContext) {})
	assert.DeepEqual(t, "/v1", v1.BasePath())

	login := v1.Group("/login/", func(c context.Context, ctx *app.RequestContext) {}, func(c context.Context, ctx *app.RequestContext) {})
	assert.DeepEqual(t, "/v1/login/", login.BasePath())

	handler := func(c context.Context, ctx *app.RequestContext) {
		ctx.String(http.StatusBadRequest, "the method was %s and index %d", string(ctx.Request.Header.Method()), ctx.GetIndex())
	}

	switch method {
	case http.MethodGet:
		v1.GET("/test", handler)
		login.GET("/test", handler)
	case http.MethodPost:
		v1.POST("/test", handler)
		login.POST("/test", handler)
	case http.MethodPut:
		v1.PUT("/test", handler)
		login.PUT("/test", handler)
	case http.MethodPatch:
		v1.PATCH("/test", handler)
		login.PATCH("/test", handler)
	case http.MethodDelete:
		v1.DELETE("/test", handler)
		login.DELETE("/test", handler)
	case http.MethodHead:
		v1.HEAD("/test", handler)
		login.HEAD("/test", handler)
	case http.MethodOptions:
		v1.OPTIONS("/test", handler)
		login.OPTIONS("/test", handler)
	default:
		panic("unknown method")
	}

	w := performRequest(router, method, "/v1/login/test")
	assert.DeepEqual(t, http.StatusBadRequest, w.Code)
	assert.DeepEqual(t, "the method was "+method+" and index 3", w.Body.String())

	w = performRequest(router, method, "/v1/test")
	assert.DeepEqual(t, http.StatusBadRequest, w.Code)
	assert.DeepEqual(t, "the method was "+method+" and index 1", w.Body.String())
}

func TestRouterGroupStatic(t *testing.T) {
	router := NewEngine(config.NewOptions(nil))
	router.Static("/", ".")
	w := performRequest(router, "GET", "/engine.go")
	fd, err := os.Open("./engine.go")
	if err != nil {
		panic(err)
	}
	assert.DeepEqual(t, http.StatusOK, w.Code)
	defer fd.Close()
	content, err := ioutil.ReadAll(fd)
	if err != nil {
		panic(err)
	}
	assert.DeepEqual(t, string(content), w.Body.String())
}

func TestRouterGroupStaticFile(t *testing.T) {
	router := NewEngine(config.NewOptions(nil))
	router.StaticFile("file", "./engine.go")
	w := performRequest(router, "GET", "/file")
	assert.DeepEqual(t, http.StatusOK, w.Code)
	fd, err := os.Open("./engine.go")
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	content, err := ioutil.ReadAll(fd)
	if err != nil {
		panic(err)
	}
	assert.DeepEqual(t, string(content), w.Body.String())
}

func TestRouterGroupInvalidStatic(t *testing.T) {
	router := &RouterGroup{
		Handlers: nil,
		basePath: "/",
		root:     true,
	}
	assert.Panic(t, func() {
		router.Static("/path/:param", "/")
	})

	assert.Panic(t, func() {
		router.Static("/path/*param", "/")
	})
}

func TestRouterGroupInvalidStaticFile(t *testing.T) {
	router := &RouterGroup{
		Handlers: nil,
		basePath: "/",
		root:     true,
	}
	assert.Panic(t, func() {
		router.StaticFile("/path/:param", "favicon.ico")
	})

	assert.Panic(t, func() {
		router.StaticFile("/path/*param", "favicon.ico")
	})
}

func TestRouterGroupTooManyHandlers(t *testing.T) {
	engine := NewEngine(config.NewOptions(nil))
	handlers1 := make([]app.HandlerFunc, 40)
	engine.Use(handlers1...)

	handlers2 := make([]app.HandlerFunc, 26)
	assert.Panic(t, func() {
		engine.Use(handlers2...)
	})
	assert.Panic(t, func() {
		engine.GET("/", handlers2...)
	})
}

func TestRouterGroupBadMethod(t *testing.T) {
	router := &RouterGroup{
		Handlers: nil,
		basePath: "/",
		root:     true,
	}
	assert.Panic(t, func() {
		router.Handle(http.MethodGet, "/")
	})
	assert.Panic(t, func() {
		router.Handle(" GET", "/")
	})
	assert.Panic(t, func() {
		router.Handle("GET ", "/")
	})
	assert.Panic(t, func() {
		router.Handle("", "/")
	})
	assert.Panic(t, func() {
		router.Handle("PO ST", "/")
	})
	assert.Panic(t, func() {
		router.Handle("1GET", "/")
	})
	assert.Panic(t, func() {
		router.Handle("PATCh", "/")
	})
}

func TestRouterGroupPipeline(t *testing.T) {
	opt := config.NewOptions([]config.Option{})
	router := NewEngine(opt)
	testRoutesInterface(t, router)

	v1 := router.Group("/v1")
	testRoutesInterface(t, v1)
}

func testRoutesInterface(t *testing.T, r IRoutes) {
	handler := func(c context.Context, ctx *app.RequestContext) {}
	assert.DeepEqual(t, r, r.Use(handler))

	assert.DeepEqual(t, r, r.Handle(http.MethodGet, "/handler", handler))
	assert.DeepEqual(t, r, r.Any("/any", handler))
	assert.DeepEqual(t, r, r.GET("/", handler))
	assert.DeepEqual(t, r, r.POST("/", handler))
	assert.DeepEqual(t, r, r.DELETE("/", handler))
	assert.DeepEqual(t, r, r.PATCH("/", handler))
	assert.DeepEqual(t, r, r.PUT("/", handler))
	assert.DeepEqual(t, r, r.OPTIONS("/", handler))
	assert.DeepEqual(t, r, r.HEAD("/", handler))

	assert.DeepEqual(t, r, r.StaticFile("/file", "."))
	assert.DeepEqual(t, r, r.Static("/static", "."))
	assert.DeepEqual(t, r, r.StaticFS("/static2", &app.FS{}))
}

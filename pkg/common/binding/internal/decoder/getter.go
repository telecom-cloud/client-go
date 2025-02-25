package decoder

import (
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/route/param"
)

type getter func(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret string, exist bool)

func path(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret string, exist bool) {
	if params != nil {
		ret, exist = params.Get(key)
	}

	if len(ret) == 0 && len(defaultValue) != 0 {
		ret = defaultValue[0]
	}
	return ret, exist
}

func postForm(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret string, exist bool) {
	if ret, exist = req.PostArgs().PeekExists(key); exist {
		return
	}

	mf, err := req.MultipartForm()
	if err == nil && mf.Value != nil {
		for k, v := range mf.Value {
			if k == key && len(v) > 0 {
				ret = v[0]
			}
		}
	}

	if len(ret) != 0 {
		return ret, true
	}
	if ret, exist = req.URI().QueryArgs().PeekExists(key); exist {
		return
	}

	if len(ret) == 0 && len(defaultValue) != 0 {
		ret = defaultValue[0]
	}

	return ret, false
}

func query(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret string, exist bool) {
	if ret, exist = req.URI().QueryArgs().PeekExists(key); exist {
		return
	}

	if len(ret) == 0 && len(defaultValue) != 0 {
		ret = defaultValue[0]
	}

	return
}

func cookie(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret string, exist bool) {
	if val := req.Header.Cookie(key); val != nil {
		ret = string(val)
		return ret, true
	}

	if len(ret) == 0 && len(defaultValue) != 0 {
		ret = defaultValue[0]
	}

	return ret, false
}

func header(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret string, exist bool) {
	if val := req.Header.Peek(key); val != nil {
		ret = string(val)
		return ret, true
	}

	if len(ret) == 0 && len(defaultValue) != 0 {
		ret = defaultValue[0]
	}

	return ret, false
}

func rawBody(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret string, exist bool) {
	exist = false
	if req.Header.ContentLength() > 0 {
		ret = string(req.Body())
		exist = true
	}
	return
}

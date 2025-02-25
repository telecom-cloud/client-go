package decoder

import (
	"github.com/telecom-cloud/client-go/internal/bytesconv"
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/route/param"
)

type sliceGetter func(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret []string)

func pathSlice(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret []string) {
	var value string
	if params != nil {
		value, _ = params.Get(key)
	}

	if len(value) == 0 && len(defaultValue) != 0 {
		value = defaultValue[0]
	}
	if len(value) != 0 {
		ret = append(ret, value)
	}

	return
}

func postFormSlice(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret []string) {
	req.PostArgs().VisitAll(func(formKey, value []byte) {
		if bytesconv.B2s(formKey) == key {
			ret = append(ret, string(value))
		}
	})
	if len(ret) > 0 {
		return
	}

	mf, err := req.MultipartForm()
	if err == nil && mf.Value != nil {
		for k, v := range mf.Value {
			if k == key && len(v) > 0 {
				ret = append(ret, v...)
			}
		}
	}
	if len(ret) > 0 {
		return
	}

	if len(ret) == 0 && len(defaultValue) != 0 {
		ret = append(ret, defaultValue...)
	}

	return
}

func querySlice(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret []string) {
	req.URI().QueryArgs().VisitAll(func(queryKey, value []byte) {
		if key == bytesconv.B2s(queryKey) {
			ret = append(ret, string(value))
		}
	})

	if len(ret) == 0 && len(defaultValue) != 0 {
		ret = append(ret, defaultValue...)
	}

	return
}

func cookieSlice(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret []string) {
	req.Header.VisitAllCookie(func(cookieKey, value []byte) {
		if bytesconv.B2s(cookieKey) == key {
			ret = append(ret, string(value))
		}
	})

	if len(ret) == 0 && len(defaultValue) != 0 {
		ret = append(ret, defaultValue...)
	}

	return
}

func headerSlice(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret []string) {
	req.Header.VisitAll(func(headerKey, value []byte) {
		if bytesconv.B2s(headerKey) == key {
			ret = append(ret, string(value))
		}
	})

	if len(ret) == 0 && len(defaultValue) != 0 {
		ret = append(ret, defaultValue...)
	}

	return
}

func rawBodySlice(req *protocol.Request, params param.Params, key string, defaultValue ...string) (ret []string) {
	if req.Header.ContentLength() > 0 {
		ret = append(ret, string(req.Body()))
	}
	return
}

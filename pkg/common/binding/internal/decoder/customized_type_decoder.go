package decoder

import (
	"reflect"

	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/route/param"
)

type CustomizeDecodeFunc func(req *protocol.Request, params param.Params, text string) (reflect.Value, error)

type customizedFieldTextDecoder struct {
	fieldInfo
	decodeFunc CustomizeDecodeFunc
}

func (d *customizedFieldTextDecoder) Decode(req *protocol.Request, params param.Params, reqValue reflect.Value) error {
	var text string
	var exist bool
	var defaultValue string
	for _, tagInfo := range d.tagInfos {
		if tagInfo.Skip || tagInfo.Key == jsonTag || tagInfo.Key == fileNameTag {
			if tagInfo.Key == jsonTag {
				defaultValue = tagInfo.Default
				if len(tagInfo.Default) != 0 && keyExist(req, tagInfo) {
					defaultValue = ""
				}
			}
			continue
		}
		text, exist = tagInfo.Getter(req, params, tagInfo.Value)
		defaultValue = tagInfo.Default
		if exist {
			break
		}
	}
	if !exist {
		return nil
	}
	if len(text) == 0 && len(defaultValue) != 0 {
		text = toDefaultValue(d.fieldType, defaultValue)
	}

	v, err := d.decodeFunc(req, params, text)
	if err != nil {
		return err
	}
	if !v.IsValid() {
		return nil
	}

	reqValue = GetFieldValue(reqValue, d.parentIndex)
	field := reqValue.Field(d.index)
	if field.Kind() == reflect.Ptr {
		t := field.Type()
		var ptrDepth int
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
			ptrDepth++
		}
		field.Set(ReferenceValue(v, ptrDepth))
		return nil
	}

	field.Set(v)
	return nil
}

func getCustomizedFieldDecoder(field reflect.StructField, index int, tagInfos []TagInfo, parentIdx []int, decodeFunc CustomizeDecodeFunc, config *DecodeConfig) ([]fieldDecoder, error) {
	for idx, tagInfo := range tagInfos {
		switch tagInfo.Key {
		case pathTag:
			tagInfos[idx].SliceGetter = pathSlice
			tagInfos[idx].Getter = path
		case formTag:
			tagInfos[idx].SliceGetter = postFormSlice
			tagInfos[idx].Getter = postForm
		case queryTag:
			tagInfos[idx].SliceGetter = querySlice
			tagInfos[idx].Getter = query
		case cookieTag:
			tagInfos[idx].SliceGetter = cookieSlice
			tagInfos[idx].Getter = cookie
		case headerTag:
			tagInfos[idx].SliceGetter = headerSlice
			tagInfos[idx].Getter = header
		case jsonTag:
			// do nothing
		case rawBodyTag:
			tagInfos[idx].SliceGetter = rawBodySlice
			tagInfos[idx].Getter = rawBody
		case fileNameTag:
			// do nothing
		default:
		}
	}
	fieldType := field.Type
	for field.Type.Kind() == reflect.Ptr {
		fieldType = field.Type.Elem()
	}
	return []fieldDecoder{&customizedFieldTextDecoder{
		fieldInfo: fieldInfo{
			index:       index,
			parentIndex: parentIdx,
			fieldName:   field.Name,
			tagInfos:    tagInfos,
			fieldType:   fieldType,
			config:      config,
		},
		decodeFunc: decodeFunc,
	}}, nil
}

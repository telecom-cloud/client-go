package decoder

import (
	"fmt"
	"reflect"

	"github.com/telecom-cloud/client-go/internal/bytesconv"
	hJson "github.com/telecom-cloud/client-go/pkg/common/json"
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/route/param"
)

type mapTypeFieldTextDecoder struct {
	fieldInfo
}

func (d *mapTypeFieldTextDecoder) Decode(req *protocol.Request, params param.Params, reqValue reflect.Value) error {
	var err error
	var text string
	var exist bool
	var defaultValue string
	for _, tagInfo := range d.tagInfos {
		if tagInfo.Skip || tagInfo.Key == jsonTag || tagInfo.Key == fileNameTag {
			if tagInfo.Key == jsonTag {
				defaultValue = tagInfo.Default
				found := checkRequireJSON(req, tagInfo)
				if found {
					err = nil
				} else {
					err = fmt.Errorf("'%s' field is a 'required' parameter, but the request does not have this parameter", tagInfo.Value)
				}
				if len(tagInfo.Default) != 0 && keyExist(req, tagInfo) {
					defaultValue = ""
				}
			}
			continue
		}
		text, exist = tagInfo.Getter(req, params, tagInfo.Value)
		defaultValue = tagInfo.Default
		if exist {
			err = nil
			break
		}
		if tagInfo.Required {
			err = fmt.Errorf("'%s' field is a 'required' parameter, but the request does not have this parameter", tagInfo.Value)
		}
	}
	if err != nil {
		return err
	}
	if len(text) == 0 && len(defaultValue) != 0 {
		text = toDefaultValue(d.fieldType, defaultValue)
	}
	if !exist && len(text) == 0 {
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
		var vv reflect.Value
		vv, err := stringToValue(t, text, req, params, d.config)
		if err != nil {
			return fmt.Errorf("unable to decode '%s' as %s: %w", text, d.fieldType.Name(), err)
		}
		field.Set(ReferenceValue(vv, ptrDepth))
		return nil
	}

	err = hJson.Unmarshal(bytesconv.S2b(text), field.Addr().Interface())
	if err != nil {
		return fmt.Errorf("unable to decode '%s' as %s: %w", text, d.fieldType.Name(), err)
	}

	return nil
}

func getMapTypeTextDecoder(field reflect.StructField, index int, tagInfos []TagInfo, parentIdx []int, config *DecodeConfig) ([]fieldDecoder, error) {
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

	return []fieldDecoder{&mapTypeFieldTextDecoder{
		fieldInfo: fieldInfo{
			index:       index,
			parentIndex: parentIdx,
			fieldName:   field.Name,
			tagInfos:    tagInfos,
			fieldType:   fieldType,
			config:      config,
		},
	}}, nil
}

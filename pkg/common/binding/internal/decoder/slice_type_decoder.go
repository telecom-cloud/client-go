package decoder

import (
	"fmt"
	"mime/multipart"
	"reflect"

	"github.com/telecom-cloud/client-go/internal/bytesconv"
	hJson "github.com/telecom-cloud/client-go/pkg/common/json"
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/route/param"
)

type sliceTypeFieldTextDecoder struct {
	fieldInfo
	isArray bool
}

func (d *sliceTypeFieldTextDecoder) Decode(req *protocol.Request, params param.Params, reqValue reflect.Value) error {
	var err error
	var texts []string
	var defaultValue string
	var bindRawBody bool
	var isDefault bool
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
				if len(tagInfo.Default) != 0 && keyExist(req, tagInfo) { //
					defaultValue = ""
				}
			}
			continue
		}
		if tagInfo.Key == rawBodyTag {
			bindRawBody = true
		}
		texts = tagInfo.SliceGetter(req, params, tagInfo.Value)
		defaultValue = tagInfo.Default
		if len(texts) != 0 {
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
	if len(texts) == 0 && len(defaultValue) != 0 {
		defaultValue = toDefaultValue(d.fieldType, defaultValue)
		texts = append(texts, defaultValue)
		isDefault = true
	}
	if len(texts) == 0 {
		return nil
	}

	reqValue = GetFieldValue(reqValue, d.parentIndex)
	field := reqValue.Field(d.index)
	// **[]**int
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			nonNilVal, ptrDepth := GetNonNilReferenceValue(field)
			field.Set(ReferenceValue(nonNilVal, ptrDepth))
		}
	}
	var parentPtrDepth int
	for field.Kind() == reflect.Ptr {
		field = field.Elem()
		parentPtrDepth++
	}

	if d.isArray {
		if len(texts) != field.Len() && !isDefault {
			return fmt.Errorf("%q is not valid value for %s", texts, field.Type().String())
		}
	} else {
		// slice need creating enough capacity
		field = reflect.MakeSlice(field.Type(), len(texts), len(texts))
	}
	// raw_body && []byte binding
	if bindRawBody && field.Type().Elem().Kind() == reflect.Uint8 {
		reqValue.Field(d.index).Set(reflect.ValueOf(req.Body()))
		return nil
	}

	// handle internal multiple pointer, []**int
	var ptrDepth int
	t := d.fieldType.Elem() // d.fieldType is non-pointer type for the field
	elemKind := t.Kind()
	for elemKind == reflect.Ptr {
		t = t.Elem()
		elemKind = t.Kind()
		ptrDepth++
	}
	if isDefault {
		err = hJson.Unmarshal(bytesconv.S2b(texts[0]), reqValue.Field(d.index).Addr().Interface())
		if err != nil {
			return fmt.Errorf("using '%s' to unmarshal field '%s: %s' failed, %v", texts[0], d.fieldName, d.fieldType.String(), err)
		}
		return nil
	}

	for idx, text := range texts {
		var vv reflect.Value
		vv, err = stringToValue(t, text, req, params, d.config)
		if err != nil {
			break
		}
		field.Index(idx).Set(ReferenceValue(vv, ptrDepth))
	}
	if err != nil {
		if !reqValue.Field(d.index).CanAddr() {
			return err
		}
		// text[0] can be a complete json content for []Type.
		err = hJson.Unmarshal(bytesconv.S2b(texts[0]), reqValue.Field(d.index).Addr().Interface())
		if err != nil {
			return fmt.Errorf("using '%s' to unmarshal field '%s: %s' failed, %v", texts[0], d.fieldName, d.fieldType.String(), err)
		}
	} else {
		reqValue.Field(d.index).Set(ReferenceValue(field, parentPtrDepth))
	}

	return nil
}

func getSliceFieldDecoder(field reflect.StructField, index int, tagInfos []TagInfo, parentIdx []int, config *DecodeConfig) ([]fieldDecoder, error) {
	if !(field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array) {
		return nil, fmt.Errorf("unexpected type %s, expected slice or array", field.Type.String())
	}
	isArray := false
	if field.Type.Kind() == reflect.Array {
		isArray = true
	}
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
	// fieldType.Elem() is the type for array/slice elem
	t := getElemType(fieldType.Elem())
	if t == reflect.TypeOf(multipart.FileHeader{}) {
		return getMultipartFileDecoder(field, index, tagInfos, parentIdx, config)
	}

	return []fieldDecoder{&sliceTypeFieldTextDecoder{
		fieldInfo: fieldInfo{
			index:       index,
			parentIndex: parentIdx,
			fieldName:   field.Name,
			tagInfos:    tagInfos,
			fieldType:   fieldType,
			config:      config,
		},
		isArray: isArray,
	}}, nil
}

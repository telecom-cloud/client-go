package decoder

import (
	"fmt"
	"github.com/telecom-cloud/client-go/pkg/common/logger"
	"reflect"

	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/route/param"
)

type fileTypeDecoder struct {
	fieldInfo
	isRepeated bool
}

func (d *fileTypeDecoder) Decode(req *protocol.Request, params param.Params, reqValue reflect.Value) error {
	fieldValue := GetFieldValue(reqValue, d.parentIndex)
	field := fieldValue.Field(d.index)

	if d.isRepeated {
		return d.fileSliceDecode(req, params, reqValue)
	}
	var fileName string
	// file_name > form > fieldName
	for _, tagInfo := range d.tagInfos {
		if tagInfo.Key == fileNameTag {
			fileName = tagInfo.Value
			break
		}
		if tagInfo.Key == formTag {
			fileName = tagInfo.Value
		}
	}
	if len(fileName) == 0 {
		fileName = d.fieldName
	}
	file, err := req.FormFile(fileName)
	if err != nil {
		logger.SystemLogger().Warnf("can not get file '%s' form request, reason: %v, so skip '%s' field binding", fileName, err, d.fieldName)
		return nil
	}
	if field.Kind() == reflect.Ptr {
		t := field.Type()
		var ptrDepth int
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
			ptrDepth++
		}
		v := reflect.New(t).Elem()
		v.Set(reflect.ValueOf(*file))
		field.Set(ReferenceValue(v, ptrDepth))
		return nil
	}

	// Non-pointer elems
	field.Set(reflect.ValueOf(*file))

	return nil
}

func (d *fileTypeDecoder) fileSliceDecode(req *protocol.Request, params param.Params, reqValue reflect.Value) error {
	fieldValue := GetFieldValue(reqValue, d.parentIndex)
	field := fieldValue.Field(d.index)
	// 如果没值，需要为其建一个值
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

	var fileName string
	// file_name > form > fieldName
	for _, tagInfo := range d.tagInfos {
		if tagInfo.Key == fileNameTag {
			fileName = tagInfo.Value
			break
		}
		if tagInfo.Key == formTag {
			fileName = tagInfo.Value
		}
	}
	if len(fileName) == 0 {
		fileName = d.fieldName
	}
	multipartForm, err := req.MultipartForm()
	if err != nil {
		logger.SystemLogger().Warnf("can not get MultipartForm from request, reason: %v, so skip '%s' field binding", fileName, err, d.fieldName)
		return nil
	}
	files, exist := multipartForm.File[fileName]
	if !exist {
		logger.SystemLogger().Warnf("the file '%s' is not existed in request, so skip '%s' field binding", fileName, d.fieldName)
		return nil
	}

	if field.Kind() == reflect.Array {
		if len(files) != field.Len() {
			return fmt.Errorf("the numbers(%d) of file '%s' does not match the length(%d) of %s", len(files), fileName, field.Len(), field.Type().String())
		}
	} else {
		// slice need creating enough capacity
		field = reflect.MakeSlice(field.Type(), len(files), len(files))
	}

	// handle multiple pointer
	var ptrDepth int
	t := d.fieldType.Elem()
	elemKind := t.Kind()
	for elemKind == reflect.Ptr {
		t = t.Elem()
		elemKind = t.Kind()
		ptrDepth++
	}

	for idx, file := range files {
		v := reflect.New(t).Elem()
		v.Set(reflect.ValueOf(*file))
		field.Index(idx).Set(ReferenceValue(v, ptrDepth))
	}
	fieldValue.Field(d.index).Set(ReferenceValue(field, parentPtrDepth))

	return nil
}

func getMultipartFileDecoder(field reflect.StructField, index int, tagInfos []TagInfo, parentIdx []int, config *DecodeConfig) ([]fieldDecoder, error) {
	fieldType := field.Type
	for field.Type.Kind() == reflect.Ptr {
		fieldType = field.Type.Elem()
	}
	isRepeated := false
	if fieldType.Kind() == reflect.Array || fieldType.Kind() == reflect.Slice {
		isRepeated = true
	}

	return []fieldDecoder{&fileTypeDecoder{
		fieldInfo: fieldInfo{
			index:       index,
			parentIndex: parentIdx,
			fieldName:   field.Name,
			tagInfos:    tagInfos,
			fieldType:   fieldType,
			config:      config,
		},
		isRepeated: isRepeated,
	}}, nil
}

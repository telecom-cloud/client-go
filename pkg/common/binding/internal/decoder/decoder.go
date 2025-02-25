package decoder

import (
	"fmt"
	"mime/multipart"
	"reflect"

	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/route/param"
)

type fieldDecoder interface {
	Decode(req *protocol.Request, params param.Params, reqValue reflect.Value) error
}

type Decoder func(req *protocol.Request, params param.Params, rv reflect.Value) error

type DecodeConfig struct {
	LooseZeroMode                      bool
	DisableDefaultTag                  bool
	DisableStructFieldResolve          bool
	EnableDecoderUseNumber             bool
	EnableDecoderDisallowUnknownFields bool
	ValidateTag                        string
	TypeUnmarshalFuncs                 map[reflect.Type]CustomizeDecodeFunc
}

func GetReqDecoder(rt reflect.Type, byTag string, config *DecodeConfig) (Decoder, bool, error) {
	var decoders []fieldDecoder
	var needValidate bool

	el := rt.Elem()
	if el.Kind() != reflect.Struct {
		return nil, false, fmt.Errorf("unsupported \"%s\" type binding", rt.String())
	}

	for i := 0; i < el.NumField(); i++ {
		if el.Field(i).PkgPath != "" && !el.Field(i).Anonymous {
			// ignore unexported field
			continue
		}

		dec, needValidate2, err := getFieldDecoder(parentInfos{[]reflect.Type{el}, []int{}, ""}, el.Field(i), i, byTag, config)
		if err != nil {
			return nil, false, err
		}
		needValidate = needValidate || needValidate2

		if dec != nil {
			decoders = append(decoders, dec...)
		}
	}

	return func(req *protocol.Request, params param.Params, rv reflect.Value) error {
		for _, decoder := range decoders {
			err := decoder.Decode(req, params, rv)
			if err != nil {
				return err
			}
		}

		return nil
	}, needValidate, nil
}

type parentInfos struct {
	Types    []reflect.Type
	Indexes  []int
	JSONName string
}

func getFieldDecoder(pInfo parentInfos, field reflect.StructField, index int, byTag string, config *DecodeConfig) ([]fieldDecoder, bool, error) {
	for field.Type.Kind() == reflect.Ptr {
		field.Type = field.Type.Elem()
	}
	// skip anonymous definitions, like:
	// type A struct {
	// 		string
	// }
	if field.Type.Kind() != reflect.Struct && field.Anonymous {
		return nil, false, nil
	}

	// JSONName is like 'a.b.c' for 'required validate'
	fieldTagInfos, newParentJSONName, needValidate := lookupFieldTags(field, pInfo.JSONName, config)
	if len(fieldTagInfos) == 0 && !config.DisableDefaultTag {
		fieldTagInfos, newParentJSONName = getDefaultFieldTags(field, pInfo.JSONName)
	}
	if len(byTag) != 0 {
		fieldTagInfos = getFieldTagInfoByTag(field, byTag)
	}

	// customized type decoder has the highest priority
	if customizedFunc, exist := config.TypeUnmarshalFuncs[field.Type]; exist {
		dec, err := getCustomizedFieldDecoder(field, index, fieldTagInfos, pInfo.Indexes, customizedFunc, config)
		return dec, needValidate, err
	}

	// slice/array field decoder
	if field.Type.Kind() == reflect.Slice || field.Type.Kind() == reflect.Array {
		dec, err := getSliceFieldDecoder(field, index, fieldTagInfos, pInfo.Indexes, config)
		return dec, needValidate, err
	}

	// map filed decoder
	if field.Type.Kind() == reflect.Map {
		dec, err := getMapTypeTextDecoder(field, index, fieldTagInfos, pInfo.Indexes, config)
		return dec, needValidate, err
	}

	// struct field will be resolved recursively
	if field.Type.Kind() == reflect.Struct {
		var decoders []fieldDecoder
		el := field.Type
		// todo: more built-in common struct binding, ex. time...
		switch el {
		case reflect.TypeOf(multipart.FileHeader{}): // file binding
			dec, err := getMultipartFileDecoder(field, index, fieldTagInfos, pInfo.Indexes, config)
			return dec, needValidate, err
		}
		if !config.DisableStructFieldResolve { // decode struct type separately
			structFieldDecoder, err := getStructTypeFieldDecoder(field, index, fieldTagInfos, pInfo.Indexes, config)
			if err != nil {
				return nil, needValidate, err
			}
			if structFieldDecoder != nil {
				decoders = append(decoders, structFieldDecoder...)
			}
		}

		// prevent infinite recursion when struct field with the same name as a struct
		if hasSameType(pInfo.Types, el) {
			return decoders, needValidate, nil
		}

		pIdx := pInfo.Indexes
		for i := 0; i < el.NumField(); i++ {
			if el.Field(i).PkgPath != "" && !el.Field(i).Anonymous {
				// ignore unexported field
				continue
			}
			var idxes []int
			if len(pInfo.Indexes) > 0 {
				idxes = append(idxes, pIdx...)
			}
			idxes = append(idxes, index)
			pInfo.Indexes = idxes
			pInfo.Types = append(pInfo.Types, el)
			pInfo.JSONName = newParentJSONName
			dec, needValidate2, err := getFieldDecoder(pInfo, el.Field(i), i, byTag, config)
			needValidate = needValidate || needValidate2
			if err != nil {
				return nil, false, err
			}
			if dec != nil {
				decoders = append(decoders, dec...)
			}
		}

		return decoders, needValidate, nil
	}

	// base type decoder
	dec, err := getBaseTypeTextDecoder(field, index, fieldTagInfos, pInfo.Indexes, config)
	return dec, needValidate, err
}

// hasSameType determine if the same type is present in the parent-child relationship
func hasSameType(pts []reflect.Type, ft reflect.Type) bool {
	for _, pt := range pts {
		if reflect.DeepEqual(getElemType(pt), getElemType(ft)) {
			return true
		}
	}
	return false
}

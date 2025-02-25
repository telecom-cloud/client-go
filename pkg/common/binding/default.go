package binding

import (
	"bytes"
	stdJson "encoding/json"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"
	"sync"

	"github.com/telecom-cloud/client-go/internal/bytesconv"
	exprValidator "github.com/telecom-cloud/client-go/internal/tagexpr/validator"
	inDecoder "github.com/telecom-cloud/client-go/pkg/common/binding/internal/decoder"
	hJson "github.com/telecom-cloud/client-go/pkg/common/json"
	"github.com/telecom-cloud/client-go/pkg/common/utils"
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/protocol/consts"
	"github.com/telecom-cloud/client-go/pkg/route/param"
	"google.golang.org/protobuf/proto"
)

const (
	queryTag           = "query"
	headerTag          = "header"
	formTag            = "form"
	pathTag            = "path"
	defaultValidateTag = "vd"
)

type decoderInfo struct {
	decoder      inDecoder.Decoder
	needValidate bool
}

var defaultBind = NewDefaultBinder(nil)

func DefaultBinder() Binder {
	return defaultBind
}

type defaultBinder struct {
	config             *BindConfig
	decoderCache       sync.Map
	queryDecoderCache  sync.Map
	formDecoderCache   sync.Map
	headerDecoderCache sync.Map
	pathDecoderCache   sync.Map
}

func NewDefaultBinder(config *BindConfig) Binder {
	if config == nil {
		bindConfig := NewBindConfig()
		bindConfig.initTypeUnmarshal()
		return &defaultBinder{
			config: bindConfig,
		}
	}
	config.initTypeUnmarshal()
	if config.Validator == nil {
		config.Validator = DefaultValidator()
	}
	return &defaultBinder{
		config: config,
	}
}

// BindAndValidate binds data from *protocol.Request to obj and validates them if needed.
// NOTE:
//
//	obj should be a pointer.
func BindAndValidate(req *protocol.Request, obj interface{}, pathParams param.Params) error {
	return DefaultBinder().BindAndValidate(req, obj, pathParams)
}

// Bind binds data from *protocol.Request to obj.
// NOTE:
//
//	obj should be a pointer.
func Bind(req *protocol.Request, obj interface{}, pathParams param.Params) error {
	return DefaultBinder().Bind(req, obj, pathParams)
}

// Validate validates obj with "vd" tag
// NOTE:
//
//	obj should be a pointer.
//	Validate should be called after Bind.
func Validate(obj interface{}) error {
	return DefaultValidator().ValidateStruct(obj)
}

func (b *defaultBinder) tagCache(tag string) *sync.Map {
	switch tag {
	case queryTag:
		return &b.queryDecoderCache
	case headerTag:
		return &b.headerDecoderCache
	case formTag:
		return &b.formDecoderCache
	case pathTag:
		return &b.pathDecoderCache
	default:
		return &b.decoderCache
	}
}

func (b *defaultBinder) bindTag(req *protocol.Request, v interface{}, params param.Params, tag string) error {
	rv, typeID := valueAndTypeID(v)
	if err := checkPointer(rv); err != nil {
		return err
	}
	rt := dereferPointer(rv)
	if rt.Kind() != reflect.Struct {
		return b.bindNonStruct(req, v)
	}

	if len(tag) == 0 {
		err := b.preBindBody(req, v)
		if err != nil {
			return fmt.Errorf("bind body failed, err=%v", err)
		}
	}
	cache := b.tagCache(tag)
	cached, ok := cache.Load(typeID)
	if ok {
		// cached fieldDecoder, fast path
		decoder := cached.(decoderInfo)
		return decoder.decoder(req, params, rv.Elem())
	}
	validateTag := defaultValidateTag
	if len(b.config.Validator.ValidateTag()) != 0 {
		validateTag = b.config.Validator.ValidateTag()
	}
	decodeConfig := &inDecoder.DecodeConfig{
		LooseZeroMode:                      b.config.LooseZeroMode,
		DisableDefaultTag:                  b.config.DisableDefaultTag,
		DisableStructFieldResolve:          b.config.DisableStructFieldResolve,
		EnableDecoderUseNumber:             b.config.EnableDecoderUseNumber,
		EnableDecoderDisallowUnknownFields: b.config.EnableDecoderDisallowUnknownFields,
		ValidateTag:                        validateTag,
		TypeUnmarshalFuncs:                 b.config.TypeUnmarshalFuncs,
	}
	decoder, needValidate, err := inDecoder.GetReqDecoder(rv.Type(), tag, decodeConfig)
	if err != nil {
		return err
	}

	cache.Store(typeID, decoderInfo{decoder: decoder, needValidate: needValidate})
	return decoder(req, params, rv.Elem())
}

func (b *defaultBinder) bindTagWithValidate(req *protocol.Request, v interface{}, params param.Params, tag string) error {
	rv, typeID := valueAndTypeID(v)
	if err := checkPointer(rv); err != nil {
		return err
	}
	rt := dereferPointer(rv)
	if rt.Kind() != reflect.Struct {
		return b.bindNonStruct(req, v)
	}

	err := b.preBindBody(req, v)
	if err != nil {
		return fmt.Errorf("bind body failed, err=%v", err)
	}
	cache := b.tagCache(tag)
	cached, ok := cache.Load(typeID)
	if ok {
		// cached fieldDecoder, fast path
		decoder := cached.(decoderInfo)
		err = decoder.decoder(req, params, rv.Elem())
		if err != nil {
			return err
		}
		if decoder.needValidate {
			err = b.config.Validator.ValidateStruct(rv.Elem())
		}
		return err
	}
	validateTag := defaultValidateTag
	if len(b.config.Validator.ValidateTag()) != 0 {
		validateTag = b.config.Validator.ValidateTag()
	}
	decodeConfig := &inDecoder.DecodeConfig{
		LooseZeroMode:                      b.config.LooseZeroMode,
		DisableDefaultTag:                  b.config.DisableDefaultTag,
		DisableStructFieldResolve:          b.config.DisableStructFieldResolve,
		EnableDecoderUseNumber:             b.config.EnableDecoderUseNumber,
		EnableDecoderDisallowUnknownFields: b.config.EnableDecoderDisallowUnknownFields,
		ValidateTag:                        validateTag,
		TypeUnmarshalFuncs:                 b.config.TypeUnmarshalFuncs,
	}
	decoder, needValidate, err := inDecoder.GetReqDecoder(rv.Type(), tag, decodeConfig)
	if err != nil {
		return err
	}

	cache.Store(typeID, decoderInfo{decoder: decoder, needValidate: needValidate})
	err = decoder(req, params, rv.Elem())
	if err != nil {
		return err
	}
	if needValidate {
		err = b.config.Validator.ValidateStruct(rv.Elem())
	}
	return err
}

func (b *defaultBinder) BindQuery(req *protocol.Request, v interface{}) error {
	return b.bindTag(req, v, nil, queryTag)
}

func (b *defaultBinder) BindHeader(req *protocol.Request, v interface{}) error {
	return b.bindTag(req, v, nil, headerTag)
}

func (b *defaultBinder) BindPath(req *protocol.Request, v interface{}, params param.Params) error {
	return b.bindTag(req, v, params, pathTag)
}

func (b *defaultBinder) BindForm(req *protocol.Request, v interface{}) error {
	return b.bindTag(req, v, nil, formTag)
}

func (b *defaultBinder) BindJSON(req *protocol.Request, v interface{}) error {
	return b.decodeJSON(bytes.NewReader(req.Body()), v)
}

func (b *defaultBinder) decodeJSON(r io.Reader, obj interface{}) error {
	decoder := hJson.NewDecoder(r)
	if b.config.EnableDecoderUseNumber {
		decoder.UseNumber()
	}
	if b.config.EnableDecoderDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	return decoder.Decode(obj)
}

func (b *defaultBinder) BindProtobuf(req *protocol.Request, v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return fmt.Errorf("%s does not implement 'proto.Message'", v)
	}
	return proto.Unmarshal(req.Body(), msg)
}

func (b *defaultBinder) Name() string {
	return "hertz"
}

func (b *defaultBinder) BindAndValidate(req *protocol.Request, v interface{}, params param.Params) error {
	return b.bindTagWithValidate(req, v, params, "")
}

func (b *defaultBinder) Bind(req *protocol.Request, v interface{}, params param.Params) error {
	return b.bindTag(req, v, params, "")
}

// best effort binding
func (b *defaultBinder) preBindBody(req *protocol.Request, v interface{}) error {
	if req.Header.ContentLength() <= 0 {
		return nil
	}
	ct := bytesconv.B2s(req.Header.ContentType())
	switch strings.ToLower(utils.FilterContentType(ct)) {
	case consts.MIMEApplicationJSON:
		return hJson.Unmarshal(req.Body(), v)
	case consts.MIMEPROTOBUF:
		msg, ok := v.(proto.Message)
		if !ok {
			return fmt.Errorf("%s can not implement 'proto.Message'", v)
		}
		return proto.Unmarshal(req.Body(), msg)
	default:
		return nil
	}
}

func (b *defaultBinder) bindNonStruct(req *protocol.Request, v interface{}) (err error) {
	ct := bytesconv.B2s(req.Header.ContentType())
	switch strings.ToLower(utils.FilterContentType(ct)) {
	case consts.MIMEApplicationJSON:
		err = hJson.Unmarshal(req.Body(), v)
	case consts.MIMEPROTOBUF:
		msg, ok := v.(proto.Message)
		if !ok {
			return fmt.Errorf("%s can not implement 'proto.Message'", v)
		}
		err = proto.Unmarshal(req.Body(), msg)
	case consts.MIMEMultipartPOSTForm:
		form := make(url.Values)
		mf, err1 := req.MultipartForm()
		if err1 == nil && mf.Value != nil {
			for k, v := range mf.Value {
				for _, vv := range v {
					form.Add(k, vv)
				}
			}
		}
		b, _ := stdJson.Marshal(form)
		err = hJson.Unmarshal(b, v)
	case consts.MIMEApplicationHTMLForm:
		form := make(url.Values)
		req.PostArgs().VisitAll(func(formKey, value []byte) {
			form.Add(string(formKey), string(value))
		})
		b, _ := stdJson.Marshal(form)
		err = hJson.Unmarshal(b, v)
	default:
		// using query to decode
		query := make(url.Values)
		req.URI().QueryArgs().VisitAll(func(queryKey, value []byte) {
			query.Add(string(queryKey), string(value))
		})
		b, _ := stdJson.Marshal(query)
		err = hJson.Unmarshal(b, v)
	}
	return
}

var _ StructValidator = (*validator)(nil)

type validator struct {
	validateTag string
	validate    *exprValidator.Validator
}

func NewValidator(config *ValidateConfig) StructValidator {
	validateTag := defaultValidateTag
	if config != nil && len(config.ValidateTag) != 0 {
		validateTag = config.ValidateTag
	}
	vd := exprValidator.New(validateTag).SetErrorFactory(defaultValidateErrorFactory)
	if config != nil && config.ErrFactory != nil {
		vd.SetErrorFactory(config.ErrFactory)
	}
	return &validator{
		validateTag: validateTag,
		validate:    vd,
	}
}

// Error validate error
type validateError struct {
	FailPath, Msg string
}

// Error implements error interface.
func (e *validateError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return "invalid parameter: " + e.FailPath
}

func defaultValidateErrorFactory(failPath, msg string) error {
	return &validateError{
		FailPath: failPath,
		Msg:      msg,
	}
}

// ValidateStruct receives any kind of type, but only performed struct or pointer to struct type.
func (v *validator) ValidateStruct(obj interface{}) error {
	if obj == nil {
		return nil
	}
	return v.validate.Validate(obj)
}

// Engine returns the underlying validator
func (v *validator) Engine() interface{} {
	return v.validate
}

func (v *validator) ValidateTag() string {
	return v.validateTag
}

var defaultValidate = NewValidator(NewValidateConfig())

func DefaultValidator() StructValidator {
	return defaultValidate
}

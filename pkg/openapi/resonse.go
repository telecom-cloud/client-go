package openapi

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/tidwall/gjson"
)

type Response struct {
	StatusCode interface{} `json:"statusCode,omitempty"`
	// Error code, which is a three part code for product.module.code
	Error string `json:"error,omitempty"`
	// Error description during failure, usually in English
	Message string `json:"message,omitempty"`
	// Error description during failure, usually in Chinese
	Description string `json:"description,omitempty"`
	// Data returned upon success
	ReturnObj interface{} `json:"returnObj"`
}

func (r *Response) DeepCopy() Response {
	return Response{
		StatusCode:  r.StatusCode,
		Error:       r.Error,
		Message:     r.Message,
		Description: r.Description,
		ReturnObj:   r.ReturnObj,
	}
}

func (r *Response) ParseStatusCode() int {
	switch v := r.StatusCode.(type) {
	case string:
		i, _ := strconv.Atoi(v)
		return i
	case int:
		return v
	default:
		return 900
	}
}

func (r *Response) BindReturnObj(jsonStr string) error {
	value := gjson.Get(jsonStr, "returnObj")
	if value.IsObject() {
		return json.Unmarshal([]byte(value.String()), &r.ReturnObj)
	}

	if value.IsArray() || value.IsBool() || value.Type == gjson.Number || value.Type == gjson.String {
		buildData := fmt.Sprintf("{\"data\": %s}", value.Raw)
		return json.Unmarshal([]byte(buildData), &r.ReturnObj)
	}

	return nil
}

package openapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/tidwall/gjson"
)

type Response struct {
	InnerResponse `json:",inline"`
	// Data returned upon success
	ReturnObj interface{} `json:"returnObj"`
}

type InnerResponse struct {
	StatusCode interface{} `json:"statusCode,omitempty"`
	// Error code, which is a three part code for product.module.code
	Error string `json:"error,omitempty"`
	// Error description during failure, usually in English
	Message string `json:"message,omitempty"`
	// Error description during failure, usually in Chinese
	Description string `json:"description,omitempty"`
}

func (r *Response) DeepCopy() Response {
	resp := Response{
		ReturnObj: r.ReturnObj,
	}
	resp.StatusCode = r.StatusCode
	resp.Error = r.Error
	resp.Message = r.Message
	resp.Description = r.Description
	return resp
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

func BindResponse(jsonStr string, dst interface{}) error {
	r, ok := dst.(*Response)
	if !ok {
		return errors.New("can not bind response")
	}

	err := json.Unmarshal([]byte(jsonStr), &r.InnerResponse)
	if err != nil {
		return err
	}

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

package openapi

import (
	"fmt"
	"testing"
)

type SliceReturnObj struct {
	Data []map[string]interface{} `json:"data,omitempty"`
}

type BoolReturnObj struct {
	Data bool `json:"data,omitempty"`
}

type NumberReturnObj struct {
	Data int64 `json:"data,omitempty"`
}

type StringReturnObj struct {
	Data string `json:"data,omitempty"`
}

type ReturnObj struct {
	Name string `json:"name,omitempty"`
	Age  int    `json:"age,omitempty"`
}

func TestUnmarshalReturnObj(t *testing.T) {
	// slice
	sliceObj := &SliceReturnObj{}
	response := Response{}
	response.ReturnObj = sliceObj
	jsonData := `{"statusCode": 800,"returnObj": [{ "id": 1, "name": "A" },{ "id": 2, "name": "B" }]}`
	err := bind(jsonData, response)
	if err != nil {
		t.Error(err)
		return
	}

	// bool
	boolObj := &BoolReturnObj{}
	response = response.DeepCopy()
	response.ReturnObj = boolObj
	boolData := `{"statusCode": 800,"returnObj": true}`
	err = bind(boolData, response)
	if err != nil {
		t.Error(err)
		return
	}

	// number
	numberObj := &NumberReturnObj{}
	response = response.DeepCopy()
	response.ReturnObj = numberObj
	numberData := `{"statusCode": 800,"returnObj": 123}`
	err = bind(numberData, response)
	if err != nil {
		t.Error(err)
		return
	}

	// string
	stringObj := &StringReturnObj{}
	response = response.DeepCopy()
	response.ReturnObj = stringObj
	strData := `{"statusCode": 800,"returnObj": "test"}`
	err = bind(strData, response)
	if err != nil {
		t.Error(err)
		return
	}

	// object
	obj := &ReturnObj{}
	response = response.DeepCopy()
	response.ReturnObj = obj
	objectData := `{"statusCode": 800,"returnObj": {"name": "test", "age": 18}}`
	err = bind(objectData, response)
	if err != nil {
		t.Error(err)
		return
	}
}

func bind(data string, response Response) error {
	err := response.BindResponse(data)
	if err != nil {
		return err
	}
	fmt.Printf("statusCode: %v, returnObj: %v\n", response.StatusCode, response.ReturnObj)
	return nil
}

func TestUnmarshalStatusCode(t *testing.T) {
	response := Response{}
	strCode := `{"statusCode": "800"}`
	err := response.BindResponse(strCode)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(response.StatusCode)

	intCode := `{"statusCode": 800}`
	err = response.BindResponse(intCode)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(response.StatusCode)
}

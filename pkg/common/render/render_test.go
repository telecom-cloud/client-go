package render

import (
	"encoding/xml"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/telecom-cloud/client-go/pkg/common/test/assert"
	"github.com/telecom-cloud/client-go/pkg/common/testdata/proto"
	"github.com/telecom-cloud/client-go/pkg/protocol"
	"github.com/telecom-cloud/client-go/pkg/protocol/consts"
)

type xmlmap map[string]interface{}

// Allows type H to be used with xml.Marshal
func (h xmlmap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{
		Space: "",
		Local: "map",
	}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for key, value := range h {
		elem := xml.StartElement{
			Name: xml.Name{Space: "", Local: key},
			Attr: []xml.Attr{},
		}
		if err := e.EncodeElement(value, elem); err != nil {
			return err
		}
	}

	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

func TestRenderJSON(t *testing.T) {
	resp := &protocol.Response{}
	data := map[string]interface{}{
		"foo":  "bar",
		"html": "<b>",
	}

	(JSONRender{data}).WriteContentType(resp)
	assert.DeepEqual(t, []byte(consts.MIMEApplicationJSONUTF8), resp.Header.Peek("Content-Type"))

	err := (JSONRender{data}).Render(resp)

	assert.Nil(t, err)
	assert.DeepEqual(t, []byte("{\"foo\":\"bar\",\"html\":\"\\u003cb\\u003e\"}"), resp.Body())
	assert.DeepEqual(t, []byte(consts.MIMEApplicationJSONUTF8), resp.Header.Peek("Content-Type"))
}

func TestRenderJSONError(t *testing.T) {
	resp := &protocol.Response{}
	data := make(chan int)

	err := (JSONRender{data}).Render(resp)
	// json: unsupported type: chan int
	assert.NotNil(t, err)
}

func TestRenderPureJSON(t *testing.T) {
	resp := &protocol.Response{}
	data := map[string]interface{}{
		"foo":  "bar",
		"html": "<b>",
	}

	(PureJSON{data}).WriteContentType(resp)
	assert.DeepEqual(t, []byte(consts.MIMEApplicationJSONUTF8), resp.Header.Peek("Content-Type"))

	err := (PureJSON{data}).Render(resp)

	assert.Nil(t, err)

	assert.DeepEqual(t, []byte("{\"foo\":\"bar\",\"html\":\"<b>\"}\n"), resp.Body())
	assert.DeepEqual(t, []byte(consts.MIMEApplicationJSONUTF8), resp.Header.Peek("Content-Type"))
}

func TestRenderPureJSONError(t *testing.T) {
	resp := &protocol.Response{}
	data := make(chan int)

	err := (PureJSON{data}).Render(resp)
	// json: unsupported type: chan int
	assert.NotNil(t, err)
}

func TestRenderProtobuf(t *testing.T) {
	resp := &protocol.Response{}
	data := proto.TestStruct{Body: []byte("Hello World")}

	(ProtoBuf{&data}).WriteContentType(resp)
	assert.DeepEqual(t, []byte("application/x-protobuf"), resp.Header.Peek("Content-Type"))

	err := (ProtoBuf{&data}).Render(resp)

	assert.Nil(t, err)
	assert.DeepEqual(t, []byte("\n\vHello World"), resp.Body())
	assert.DeepEqual(t, []byte("application/x-protobuf"), resp.Header.Peek("Content-Type"))
}

func TestRenderProtobufError(t *testing.T) {
	resp := &protocol.Response{}
	data := proto.Test{}

	err := (ProtoBuf{&data}).Render(resp)

	assert.NotNil(t, err)
}

func TestRenderString(t *testing.T) {
	resp := &protocol.Response{}

	(String{
		Format: "hello %s %d",
		Data:   []interface{}{},
	}).WriteContentType(resp)
	assert.DeepEqual(t, []byte(consts.MIMETextPlainUTF8), resp.Header.Peek("Content-Type"))

	err := (String{
		Format: "hola %s %d",
		Data:   []interface{}{"manu", 2},
	}).Render(resp)

	assert.Nil(t, err)
	assert.DeepEqual(t, []byte("hola manu 2"), resp.Body())
	assert.DeepEqual(t, []byte(consts.MIMETextPlainUTF8), resp.Header.Peek("Content-Type"))
}

func TestRenderStringLenZero(t *testing.T) {
	resp := &protocol.Response{}

	err := (String{
		Format: "hola %s %d",
		Data:   []interface{}{},
	}).Render(resp)

	assert.Nil(t, err)
	assert.DeepEqual(t, []byte("hola %s %d"), resp.Body())
	assert.DeepEqual(t, []byte(consts.MIMETextPlainUTF8), resp.Header.Peek("Content-Type"))
}

func TestRenderData(t *testing.T) {
	resp := &protocol.Response{}
	data := []byte("#!PNG some raw data")

	err := (Data{
		ContentType: "image/png",
		Data:        data,
	}).Render(resp)

	assert.Nil(t, err)
	assert.DeepEqual(t, []byte("#!PNG some raw data"), resp.Body())
	assert.DeepEqual(t, []byte(consts.MIMEImagePNG), resp.Header.Peek("Content-Type"))
}

func TestRenderXML(t *testing.T) {
	resp := &protocol.Response{}
	data := xmlmap{
		"foo": "bar",
	}

	(XML{data}).WriteContentType(resp)
	assert.DeepEqual(t, []byte(consts.MIMEApplicationXMLUTF8), resp.Header.Peek("Content-Type"))

	err := (XML{data}).Render(resp)

	assert.Nil(t, err)
	assert.DeepEqual(t, []byte("<map><foo>bar</foo></map>"), resp.Body())
	assert.DeepEqual(t, []byte(consts.MIMEApplicationXMLUTF8), resp.Header.Peek("Content-Type"))
}

func TestRenderXMLError(t *testing.T) {
	resp := &protocol.Response{}
	data := make(chan int)

	err := (XML{data}).Render(resp)

	assert.NotNil(t, err)
}

func TestRenderIndentedJSON(t *testing.T) {
	data := map[string]interface{}{
		"foo":  "bar",
		"html": "h1",
	}
	t.Run("TestHeader", func(t *testing.T) {
		resp := &protocol.Response{}
		(IndentedJSON{data}).WriteContentType(resp)
		assert.DeepEqual(t, []byte(consts.MIMEApplicationJSONUTF8), resp.Header.Peek("Content-Type"))
	})
	t.Run("TestBody", func(t *testing.T) {
		ResetStdJSONMarshal()
		resp := &protocol.Response{}
		err := (IndentedJSON{data}).Render(resp)
		assert.Nil(t, err)
		assert.DeepEqual(t, []byte("{\n    \"foo\": \"bar\",\n    \"html\": \"h1\"\n}"), resp.Body())
		assert.DeepEqual(t, []byte(consts.MIMEApplicationJSONUTF8), resp.Header.Peek("Content-Type"))
		ResetJSONMarshal(sonic.Marshal)
	})
	t.Run("TestError", func(t *testing.T) {
		resp := &protocol.Response{}
		ch := make(chan int)
		err := (IndentedJSON{ch}).Render(resp)
		assert.NotNil(t, err)
	})
}

package render

import (
	"strings"
	"testing"
)

func Test_ResetStdJSONMarshal(t *testing.T) {
	table := map[string]string{
		"testA": "hello",
		"B":     "world",
	}
	ResetStdJSONMarshal()
	jsonBytes, err := jsonMarshalFunc(table)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(jsonBytes), "\"B\":\"world\"") || !strings.Contains(string(jsonBytes), "\"testA\":\"hello\"") {
		t.Fatal("marshal struct is not equal to the string")
	}
}

func Test_DefaultJSONMarshal(t *testing.T) {
	table := map[string]string{
		"testA": "hello",
		"B":     "world",
	}
	jsonBytes, err := jsonMarshalFunc(table)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(jsonBytes), "\"B\":\"world\"") || !strings.Contains(string(jsonBytes), "\"testA\":\"hello\"") {
		t.Fatal("marshal struct is not equal to the string")
	}
}

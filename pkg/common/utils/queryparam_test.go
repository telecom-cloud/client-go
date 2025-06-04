package utils

import (
	"fmt"
	"testing"
)

func TestOptimizeQueryParam(t *testing.T) {
	params := map[string]interface{}{
		"pageNum":   0,
		"pageSize":  10,
		"sort":      "",
		"condition": "",
		"order":     "",
		"name":      "test",
	}

	OptimizeQueryParams(params)

	for k, v := range params {
		fmt.Printf("%s: %v\n", k, v)
	}
}

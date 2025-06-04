package utils

import "reflect"

func IsZeroValue(v interface{}) bool {
	return reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}

func OptimizeQueryParams(params map[string]interface{}) {
	for k, v := range params {
		if IsZeroValue(v) {
			delete(params, k)
		}
	}
}

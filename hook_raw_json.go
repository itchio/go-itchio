package itchio

import (
	"encoding/json"
	"reflect"
)

// RawJSONHookFunc re-encodes an already decoded value into a
// json.RawMessage field, so raw fields keep the exact JSON the API sent.
func RawJSONHookFunc(
	f reflect.Type,
	t reflect.Type,
	data any) (any, error) {

	if t != reflect.TypeFor[json.RawMessage]() {
		return data, nil
	}

	if data == nil {
		return json.RawMessage(nil), nil
	}

	contents, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(contents), nil
}

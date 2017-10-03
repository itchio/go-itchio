package itchio

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CamelCase(t *testing.T) {
	assert.EqualValues(t, "hello", camelcase("hello"))
	assert.EqualValues(t, "helloWorld", camelcase("hello_world"))
	assert.EqualValues(t, "pOsx", camelcase("p_osx"))
	assert.EqualValues(t, "shortText", camelcase("short_text"))
}

func Test_Camelify(t *testing.T) {
	assert.EqualValues(t, "hello", camelify("hello"))

	m1 := make(map[string]interface{})
	m1["short_text"] = "short text"
	m1["min_price"] = 1200
	m1["p_osx"] = true

	var users []interface{}

	u1 := make(map[string]interface{})
	u1["full_name"] = "John Doe"
	users = append(users, u1)

	u2 := make(map[string]interface{})
	u2["full_name"] = "Jane Fischer"
	users = append(users, u2)

	m1["user_list"] = users

	mm := camelify(m1)
	assert.EqualValues(t, "short text", mm.(map[string]interface{})["shortText"])
	assert.EqualValues(t, 1200, mm.(map[string]interface{})["minPrice"])
	assert.EqualValues(t, true, mm.(map[string]interface{})["pOsx"])
	assert.EqualValues(t, "John Doe", mm.(map[string]interface{})["userList"].([]interface{})[0].(map[string]interface{})["fullName"])
	assert.EqualValues(t, "Jane Fischer", mm.(map[string]interface{})["userList"].([]interface{})[1].(map[string]interface{})["fullName"])
}
